package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)
type FeatureCollection struct {
	Type     string `json:"type"`
	Features []struct {
		Type     string `json:"type"`
		Geometry struct {
			Type        string      `json:"type"`
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
		Properties struct {
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
	} `json:"features"`
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


func GetParkingSpots(latitude1 float64, longitude1 float64, latitude2 float64, longitude2 float64) {
	url := "https://fordkerbhack.azure-api.net/features?viewport="
	url += fmt.Sprintf("%f", latitude1) + ","
	url += fmt.Sprintf("%f", longitude1) + ","
	url += fmt.Sprintf("%f", latitude2) + ","
	url += fmt.Sprintf("%f", longitude2)

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
	fmt.Println(featureCollection)

}


func main() {
	GetParkingSpots(51.536393751918915,-0.1412890195847183,51.5397303909963,-0.13700821399694973)
}