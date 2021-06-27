package CoronaAPI

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type RestCountries struct {
	Name       string `json:"name"`
	Alpha3Code string `json:"alpha3Code"`
}

// gets data in url
func getUrlData(endpointName string, r *http.Request) (string, string, string, error) {
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) != 5 {
		return "", "", "", errors.New("Wrong format, should be '/corona/v1/" + endpointName + "/{:country_name}{?scope=begin_date-end_date}'")
	}
	countryName := strings.Title(strings.ToLower(parts[4])) // makes all the letter lowercase, and then makes the first letter uppercase
	scopeQuery := r.URL.Query().Get("scope")
	var startDate, endDate = "", ""
	var err error
	if len(scopeQuery) > 0 {
		startDate, endDate, err = getStartAndEndDate(scopeQuery)
		if err != nil { // if error with queryl
			return "", "", "", errors.New(err.Error())
		}
	}

	return countryName, startDate, endDate, nil
}

// gets start and end date from date in url
func getStartAndEndDate(dates string) (string, string, error) {
	datesData := strings.Split(dates, "-")
	if len(datesData) != 6 {
		return "", "", errors.New("Wrong date format in scope. example of valid date: 2020-12-01-2021-01-31")
	}
	startDate := datesData[0] + "-" + datesData[1] + "-" + datesData[2]
	endDate := datesData[3] + "-" + datesData[4] + "-" + datesData[5]
	return startDate, endDate, nil
}

// gets country code by country name with restcountries API
func getCountryCodeByName(countryName string) (string, error) {
	resp, err := http.Get("https://restcountries.eu/rest/v2/name/" + countryName + "?fullText=true")
	var restCountry []RestCountries
	if err != nil {
		return "", errors.New("reponse error from restcountries (error with extern api)")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("Error with ioutil.ReadAll (error with extern api)")
	}
	json.Unmarshal([]byte(string(body)), &restCountry)
	if len(restCountry) < 1 {
		return "", errors.New("Can't find country with name: " + countryName)
	}
	return restCountry[0].Alpha3Code, nil
}

// converts map to WebhookRegistration struct
func mapToWebhookStruct(mapData map[string]interface{}, ID string) WebhookRegistration {
	var newWebhook WebhookRegistration
	newWebhook.ID = ID
	newWebhook.Url = mapData["url"].(string)
	newWebhook.Timeout = mapData["timeout"].(float64)
	newWebhook.Field = mapData["field"].(string)
	newWebhook.Country = mapData["country"].(string)
	newWebhook.Trigger = mapData["trigger"].(string)
	newWebhook.Occurrences = mapData["occurrences"].(float64)
	newWebhook.Time = mapData["time"].(time.Time)
	return newWebhook
}

/*
Url     string  `json:"url"`
Timeout float64 `json:"timeout"`
Field   string  `json:"field"`
Country string  `json:"country"`
Trigger string  `json:"trigger"`
*/
