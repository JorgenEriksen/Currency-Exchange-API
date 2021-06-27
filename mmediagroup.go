package CoronaAPI

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type Mmediagroup struct {
	All map[string]interface{}
}

// gets recovered data
func getRecoveredData(countryName string, r *http.Request) (Mmediagroup, error) {
	url := "https://covid-api.mmediagroup.fr/v1/history?country=" + countryName + "&status=Recovered"
	recoveredData, err := getMmediagroupData(url)
	return recoveredData, err
}

// gets confirmed data
func getConfirmedData(countryName string) (Mmediagroup, error) {
	url := "https://covid-api.mmediagroup.fr/v1/history?country=" + countryName + "&status=Confirmed"
	recoveredData, err := getMmediagroupData(url)
	return recoveredData, err
}

// gets confirmed/recovered data from mmediagroup API
func getMmediagroupData(url string) (Mmediagroup, error) {
	resp, err := http.Get(url)
	var mmediagroup Mmediagroup
	if err != nil {
		return mmediagroup, errors.New("reponse error from mmediagroup (error with extern api)")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return mmediagroup, errors.New("Error with ioutil.ReadAll (error with extern api)")
	}
	json.Unmarshal([]byte(string(body)), &mmediagroup)

	if len(mmediagroup.All) == 0 { // if country does not exist in external api
		return mmediagroup, errors.New("Can't find country. Please check the spelling and try again")
	}
	return mmediagroup, nil
}
