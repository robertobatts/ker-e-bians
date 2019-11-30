package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)

type FeatureCollection struct {
	_type			string		`json:type`
	features	[]Feature
}

type Feature struct {
	_type				string 		`json:type`
	geometry		Geometry
	properties 	Properties	
}

type Geometry struct {
	_type 				string 				`json:type`
	coordinates		[][]string
}

type Properties struct {
	images []string
	regulations []Regulation
	location Location
}

type Location struct {
	shstRefId 				string
	sideOfStreet 			string
	shstLocationStart float32
	shstLocationEnd 	float32
	objectId					string
	streetName				string
}

type Regulation struct {
	rule 			Rule
	timeSpans []TimeSpan
	priority 	int
}

type Rule struct {
	activity 	string
	reason 		string
	maxStay 	int
	noReturn 	int
	payment 	bool
}

type TimeSpan struct {
	daysOfWeek 	DaysOfWeek
	timesOfDay 	TimesOfDay
}

type DaysOfWeek struct {
	days []string
}

type TimesOfDay struct {
	from 	string
	to 		string
}


func GetParkingSpots() {
	url := "https://fordkerbhack.azure-api.net/features?viewport=51.536393751918915,-0.1412890195847183,51.5397303909963,-0.13700821399694973"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("ocp_apim_subscription_key", "f9693f68ac5e46e68995597f9ae48f4c")
	res, _ := client.Do(req)

	var featureCollection FeatureCollection
	json.NewDecoder(res.Body).Decode(&featureCollection)

	fmt.Println(featureCollection)

}


func main() {

}