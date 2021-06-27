package CoronaAPI

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type CovidTracker struct {
	StringencyData Stringency `json:"stringencyData"`
}

type Stringency struct {
	Date_value        string
	Country_code      string
	Confirmed         int
	Deaths            int
	Stringency_actual float64
	Stringency        float64
}

// gets the stringenct data on a spesific date
func getStringencyData(countryCode string, date string) (CovidTracker, error) {
	url := "https://covidtrackerapi.bsg.ox.ac.uk/api/v2/stringency/actions/" + countryCode + "/" + date
	resp, err := http.Get(url)
	var covidTracker CovidTracker
	if err != nil {
		return covidTracker, errors.New("reponse error from covidtracker (error with extern api)")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return covidTracker, errors.New("Error with ioutil.ReadAll (error with extern api)")
	}
	json.Unmarshal([]byte(body), &covidTracker)
	return covidTracker, nil
}
