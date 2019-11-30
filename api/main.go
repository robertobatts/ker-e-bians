package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"math"
)

var (
	rEarth = 6371.0 //radius of Earth in km
	avgCarLength = 0.005 //average car length in km
)

type FeatureCollection struct {
	Type     string 		`json:"type"`
	Features	 []Feature	`json:"features"`
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
			CarSpaces		int		`json:"carSpaces`
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


func GetParkingSpots(latitude float64, longitude float64, distance float64) []Feature {
	lat1, lon1 := addKmDistanceToCoordinates(latitude, longitude, -distance, -distance)
	lat2, lon2 := addKmDistanceToCoordinates(latitude, longitude, distance, distance)


	features := CallKerbspaceAPI(lat1, lon1, lat2, lon2).Features

	for i := 0; i < len(features); i++ {
		feature := &features[i]
		var newCoord = [][]float64{
			{feature.Geometry.Coordinates[0][1], feature.Geometry.Coordinates[0][0]},
			{feature.Geometry.Coordinates[1][1], feature.Geometry.Coordinates[1][0]},
		}
		feature.Geometry.Coordinates = newCoord
		carSpaces := getDistanceFromLatLonInKm(newCoord[0][0], newCoord[0][1], newCoord[1][0], newCoord[1][1]) / avgCarLength
		feature.Properties.CarSpaces = int(carSpaces)
	}

	return features
}

func getDistanceFromLatLonInKm(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64{
  dLat := deg2rad(lat2-lat1); 
  dLon := deg2rad(lon2-lon1); 
  var a = 
    math.Sin(dLat/2) * math.Sin(dLat/2) +
    math.Cos(deg2rad(lat1)) * math.Cos(deg2rad(lat2)) * 
    math.Sin(dLon/2) * math.Sin(dLon/2); 
  c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a)); 
  d := rEarth * c; // Distance in km
  return d;
}

func addKmDistanceToCoordinates(latitude float64, longitude float64, dx float64, dy float64) (float64, float64) {
	newLat  := latitude  + (dy / rEarth) * (180 / math.Pi);
	newLong := longitude + (dx / rEarth) * (180 / math.Pi) / math.Cos(deg2rad(latitude));

	return newLat, newLong
}

func deg2rad(deg float64) float64 {
  return deg * (math.Pi/180)
}




func main() {

	features := GetParkingSpots(51.55815224558373,-0.17980040235097644,0.05)
	json, _ := json.Marshal(&features)
	fmt.Println(string(json))
		
}