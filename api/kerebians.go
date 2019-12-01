package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
	"strconv"
	"sort"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
)

var apiKey string

var router *chi.Mux
var db *sql.DB

const (
	rEarth         = 6371.0 //radius of Earth in km
	avgCarLength   = 0.005  //average car length in km
	port           = 8005
	dbName         = "kerbspace"
	dbPass         = ""
	dbHost         = "localhost"
	dbPort         = "3306"
	usersTableName = "users"
	usersCols      = " id, name, surname, email, password "
)

type Directions struct {
	Code   string `json:"code"`
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration int     `json:"duration"`
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
			Type        string      `json:"type"`
		} `json:"geometry"`
		Legs []struct {
			Distance float64 `json:"distance"`
			Duration int     `json:"duration"`
			Steps    []struct {
				Distance    float64 `json:"distance"`
				DrivingSide string  `json:"driving_side"`
				Duration    int     `json:"duration"`
				Geometry    struct {
					Coordinates [][]float64 `json:"coordinates"`
					Type        string      `json:"type"`
				} `json:"geometry"`
				Intersections []struct {
					Bearings []int     `json:"bearings"`
					Entry    []bool    `json:"entry"`
					Location []float64 `json:"location"`
					Out      int       `json:"out"`
				} `json:"intersections"`
				Maneuver struct {
					BearingAfter  int       `json:"bearing_after"`
					BearingBefore int       `json:"bearing_before"`
					Instruction   string    `json:"instruction"`
					Location      []float64 `json:"location"`
					Type          string    `json:"type"`
				} `json:"maneuver"`
				Mode   string `json:"mode"`
				Name   string `json:"name"`
				Weight int    `json:"weight"`
			} `json:"steps"`
			Summary string `json:"summary"`
			Weight  int    `json:"weight"`
		} `json:"legs"`
		Weight     int    `json:"weight"`
		WeightName string `json:"weight_name"`
	} `json:"routes"`
	UUID      string `json:"uuid"`
	Waypoints []struct {
		Distance float64   `json:"distance"`
		Location []float64 `json:"location"`
		Name     string    `json:"name"`
	} `json:"waypoints"`
}

// Row is an interface which is satisfied by sdb.Row and db.Rows. It allows a
// db result to be scanned into a struct.
type Row interface {
	Scan(dest ...interface{}) error
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type JourneyReq struct {
	StartLong float32 `json:"start_long"`
	StartLat  float32 `json:"start_lat"`
	EndLong   float32 `json:"end_long"`
	EndLat    float32 `json:"end_lat"`
}

type ParkingSpotsReq struct {
	Latitude 		float64 `json:"latitude"`
	Longitude  	float64 `json:"longitude"`
	Distance		float64 `json:"distance"`
}

type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
	Manifest struct {
		CreatedDate      string `json:"createdDate"`
		TimeZone         string `json:"timeZone"`
		Currency         string `json:"currency"`
		UnitHeightLength string `json:"unitHeightLength"`
		UnitWeight       string `json:"unitWeight"`
		Authority        struct {
			Name  string `json:"name"`
			URL   string `json:"url"`
			Phone string `json:"phone"`
		} `json:"authority"`
	} `json:"manifest"`
}

type Feature struct {
	Type     string `json:"type"`
	Geometry struct {
		Type        string      `json:"type"`
		Coordinates [][]float64 `json:"coordinates"`
	} `json:"geometry"`
	Properties struct {
		CarSpaces   int           `json:"carSpaces`
		Images      []interface{} `json:"images"`
		Regulations []struct {
			Rule struct {
				Activity string `json:"activity"`
				Reason   string `json:"reason"`
				MaxStay  int    `json:"maxStay"`
				NoReturn int    `json:"noReturn"`
				Payment  bool   `json:"payment"`
			} `json:"rule"`
			UserClasses []struct {
				Classes []string `json:"classes"`
			} `json:"userClasses"`
			TimeSpans []struct {
				DaysOfWeek struct {
					Days []string `json:"days"`
				} `json:"daysOfWeek"`
				TimesOfDay struct {
					From string `json:"from"`
					To   string `json:"to"`
				} `json:"timesOfDay"`
			} `json:"timeSpans"`
			Priority int `json:"priority"`
		} `json:"regulations"`
		Location struct {
			ShstRefID         string  `json:"shstRefId"`
			SideOfStreet      string  `json:"sideOfStreet"`
			ShstLocationStart float64 `json:"shstLocationStart"`
			ShstLocationEnd   float64 `json:"shstLocationEnd"`
			ObjectID          string  `json:"objectId"`
			StreetName        string  `json:"streetName"`
		} `json:"location"`
	} `json:"properties"`
}

func CallKerbspaceAPI(latitude1 float64, longitude1 float64, latitude2 float64, longitude2 float64) FeatureCollection {
	url := "https://fordkerbhack.azure-api.net/features?viewport="
	url += fmt.Sprintf("%f", latitude1) + ","
	url += fmt.Sprintf("%f", longitude1) + ","
	url += fmt.Sprintf("%f", latitude2) + ","
	url += fmt.Sprintf("%f", longitude2)
	fmt.Println(url)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Ocp-Apim-Subscription-Key", "f9693f68ac5e46e68995597f9ae48f4c")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := client.Do(req)

	var featureCollection FeatureCollection
	err := json.NewDecoder(resp.Body).Decode(&featureCollection)
	if err != nil {
		fmt.Println(err.Error())
	}
	return featureCollection
}

func GetParkingSpots(latitude float64, longitude float64, distance float64, reason string) []Feature {

	if distance == 0 {
		distance = 0.25
	}

	lat1, lon1 := addKmDistanceToCoordinates(latitude, longitude, -distance, -distance)
	lat2, lon2 := addKmDistanceToCoordinates(latitude, longitude, distance, distance)

	features := CallKerbspaceAPI(lat1, lon1, lat2, lon2).Features

	newFeatures := []Feature{}
	for i := 0; i < len(features); i++ {
		feature := &features[i]
		if matchReason(*feature, reason) {
			var newCoord = swapCoordinates(feature.Geometry.Coordinates)

			feature.Geometry.Coordinates = newCoord
			carSpaces := getDistanceFromLatLonInKm(newCoord[0][0], newCoord[0][1], newCoord[1][0], newCoord[1][1]) / avgCarLength
			feature.Properties.CarSpaces = int(carSpaces)

			newFeatures = append(newFeatures, *feature)
		}
	}

	return newFeatures
}

func matchReason(feature Feature, reason string) bool {
	if reason == "" {
		return true
	}

	regulations := feature.Properties.Regulations
	for _, regulation :=  range regulations {
		if regulation.Rule.Reason == "unrestricted" {
			return true
		}
		if regulation.Rule.Reason == reason {
			return true
		}
	}
	return false
}

func swapCoordinates(coordinates [][]float64) [][]float64 {
	for i := 0; i < len(coordinates); i++ {
		a := coordinates[i][0]
		coordinates[i][0] = coordinates[i][1]
		coordinates[i][1] = a

		//fmt.Printf("%f,%f,abc%v\n", coordinates[i][0], coordinates[i][1], i)
	}
	return coordinates
}

func getDistanceFromLatLonInKm(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	dLat := deg2rad(lat2 - lat1)
	dLon := deg2rad(lon2 - lon1)
	var a = math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(deg2rad(lat1))*math.Cos(deg2rad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := rEarth * c // Distance in km
	return d
}

func addKmDistanceToCoordinates(latitude float64, longitude float64, dx float64, dy float64) (float64, float64) {
	newLat := latitude + (dy/rEarth)*(180/math.Pi)
	newLong := longitude + (dx/rEarth)*(180/math.Pi)/math.Cos(deg2rad(latitude))

	return newLat, newLong
}

func deg2rad(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func init() {
	router = chi.NewRouter()
	router.Use(middleware.Recoverer)
	dbSource := fmt.Sprintf("root:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true", dbPass, dbHost, dbPort, dbName)

	var err error
	db, err = sql.Open("mysql", dbSource)
	if err != nil {
		panic(err)
	}
}

func routers() *chi.Mux {
	router.Post("/users", createUser)
	router.Get("/route", route)
	router.Get("/routewithpark", routeWithPark)
	router.Get("/parkingspots", parkingSpots)

	return router
}

func main() {
	err := populateAPIDetails()
	if err != nil {
		panic(err)
	}

	routers()
	fmt.Printf("Server listen at port:%d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), Logger())
}

func populateAPIDetails() error {
	filename := "APIKey.txt"
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Cannot open key file %s\n", filename)
		return err
	}

	reader := bufio.NewReader(f)

	var line string
	for i := 0; i < 3; i++ {
		line, err = reader.ReadString('\n')
		fmt.Printf(" > Read %d characters\n", len(line))

		if len(line) > 50 {
			// Process the line here.
			fmt.Println(" > > " + line[:50])
		}

		line = strings.Replace(line, "\n", "", 1)
		if i == 0 {
			apiKey = line
		}

		if err != nil {
			break
		}
	}

	fmt.Printf("Err:%+v\n", err)
	if err != nil && err != io.EOF {
		fmt.Printf(" > Failed!: %v\n", err)
		return err
	}

	fmt.Printf("==> apiKey:%s\n", apiKey)
	return nil
}

// respondwithError return error message
func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondwithJSON(w, code, map[string]string{"message": msg})
}

// respondwithJSON write json response format
func respondwithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	//fmt.Printf("Responding with code:%+v and payload:%+v\n", code, payload)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.WriteHeader(code)
	w.Write(response)
}

// Logger return log message
func Logger() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(time.Now(), r.Method, r.URL)
		router.ServeHTTP(w, r) // dispatch the request
	})
}

func createUser(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("createUser called\n")
	var user User
	json.NewDecoder(r.Body).Decode(&user)
	defer r.Body.Close()

	fmt.Printf("User: %+v\n", user)
	query, err := db.Prepare("insert " + usersTableName + " (name, surname, email, password) " +
		"values (?, ?, ?, ?)")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := query.Exec(user.Name, user.Surname, user.Email, user.Password)
	defer query.Close()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := res.LastInsertId()

	userCreated, err := scanUsers(db.QueryRowContext(r.Context(), "select"+
		usersCols+"from "+usersTableName+" where id=?", id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	msg := fmt.Sprintf("successfully created, id:%d", id)
	fmt.Printf("msg:%s\n", msg)
	fmt.Printf("userCreated:%+v\n", userCreated)
	respondwithJSON(w, http.StatusCreated, userCreated)
}

func scanUsers(row Row) (*User, error) {
	var result User

	err := row.Scan(
		&result.ID,
		&result.Name,
		&result.Surname,
		&result.Email,
		&result.Password,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &result, err
}

func callMapBoxAPI(coordinates [][]float64) Directions {
	url := "https://api.mapbox.com/directions/v5/mapbox/driving/"
	for i := 0; i < len(coordinates); i++ {
		url += fmt.Sprintf("%f", coordinates[i][1]) + "," + fmt.Sprintf("%f", coordinates[i][0])
		if i != len(coordinates) - 1 {
			url += ";"
		}
	}

	url += "?steps=true&geometries=geojson&access_token=pk.eyJ1Ijoicm9iZXJ0b2JhdHRzIiwiYSI6ImNrM2xraXd0NzBkeWEzbm40YTVnYTFhb2kifQ.FiHpv9X4KvHDi8tFjIgZVg"

	fmt.Println(url)
	resp, _ := http.Get(url)
	fmt.Println(resp)
	
	var directions Directions
	err := json.NewDecoder(resp.Body).Decode(&directions)
	if err != nil {
		fmt.Println(err.Error())
	}

	return directions
}

func getPath(coordinates[][] float64) [][]float64 {

	directions := callMapBoxAPI(coordinates)
	//just to keep it simple, we give the user only one path
	//for i := 0; i < len(optimizedTrip.Trips); i++ {
		coord := directions.Routes[0].Geometry.Coordinates
		coord = swapCoordinates(coord)
	//}
	return coord
}

func getPathWithPark(startLat float64, startLong float64, endLat float64, endLong float64, distance float64, reason string) [][]float64 {
	parkingSpots := GetParkingSpots(endLat, endLong, distance, reason)
	
	coordinates := [][]float64{{startLat, startLong}}

	sort.Slice(parkingSpots, func(i, j int) bool {return parkingSpots[i].Properties.CarSpaces > parkingSpots[j].Properties.CarSpaces})

	size := len(parkingSpots)
	if (size > 21) {
		size = 21
	}
	for i := 0; i < size; i++ {
		coordinates = append(coordinates, parkingSpots[i].Geometry.Coordinates[0])
	}
	coordinates = append(coordinates, []float64{endLat, endLong})

	return getPath(coordinates)
}

func route(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("route called\n")

	startLat, _ := strconv.ParseFloat(r.URL.Query().Get("startLat"), 64)
	startLong, _ := strconv.ParseFloat(r.URL.Query().Get("startLong"), 64)
	endLat, _ := strconv.ParseFloat(r.URL.Query().Get("endLat"), 64)
	endLong, _ := strconv.ParseFloat(r.URL.Query().Get("endLong"), 64)

	result := getPath([][]float64{
		{startLat, startLong},
		//{51.522050, -0.108824},
		{endLat, endLong},
	})

	msg := fmt.Sprintf("successfully run")
	fmt.Printf("msg:%s\n", msg)
	respondwithJSON(w, http.StatusOK, result)
}

func routeWithPark(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("routeWithPark called\n")

	startLat, _ := strconv.ParseFloat(r.URL.Query().Get("startLat"), 64)
	startLong, _ := strconv.ParseFloat(r.URL.Query().Get("startLong"), 64)
	endLat, _ := strconv.ParseFloat(r.URL.Query().Get("endLat"), 64)
	endLong, _ := strconv.ParseFloat(r.URL.Query().Get("endLong"), 64)
	distance, _ := strconv.ParseFloat(r.URL.Query().Get("distance"), 64)
	reason := r.URL.Query().Get("reason")


	result := getPathWithPark(startLat, startLong, endLat, endLong, distance, reason)

	msg := fmt.Sprintf("successfully run")
	fmt.Printf("msg:%s\n", msg)
	respondwithJSON(w, http.StatusOK, result)
}

func parkingSpots(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("parkingSpots called\n")
	latitude, _ := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	longitude, _ := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	distance, _ := strconv.ParseFloat(r.URL.Query().Get("distance"), 64)
	reason := r.URL.Query().Get("reason")
	result := GetParkingSpots(latitude, longitude, distance, reason)

	msg := fmt.Sprintf("successfully run")
	fmt.Printf("msg:%s\n", msg)
	respondwithJSON(w, http.StatusOK, result)
}