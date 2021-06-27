package CoronaAPI

import (
	"encoding/json"
	"math"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
)

type CasesPerCountry struct {
	Country               string
	Continent             string
	Scope                 string
	Confirmed             int
	Recovered             int
	Population_percentage float64
}

type PolicyStringencyTrends struct {
	Country    string
	Scope      string
	Stringency float64
	Trend      float64
}

// invalid urls
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	status := http.StatusBadRequest
	http.Error(w, "not valid url", status)
	return
}

// http://localhost:8080/corona/v1/country/{:country_name}{?scope=begin_date-end_date}
func HandleCases(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// get request
	case http.MethodGet:
		http.Header.Add(w.Header(), "content-type", "application/json")

		// gets information in url parameter
		countryName, startDate, endDate, err := getUrlData("country", r)
		if err != nil { // if error getting url data (invalid url)
			status := http.StatusBadRequest
			http.Error(w, err.Error(), status)
			return
		}

		// gets data of confirmed
		confirmedData, err := getConfirmedData(countryName)
		if err != nil { // if error with getting data
			status := http.StatusBadRequest
			http.Error(w, err.Error(), status)
			return
		}

		// gets data of recovered
		recoveredData, err := getRecoveredData(countryName, r)
		if err != nil { // if error with getting data
			status := http.StatusBadRequest
			http.Error(w, err.Error(), status)
			return
		}

		var response CasesPerCountry
		var confirmed float64
		var recovered float64
		confirmed = 0
		confirmedDates := confirmedData.All["dates"].(map[string]interface{})
		recoveredDates := recoveredData.All["dates"].(map[string]interface{})
		if startDate != "" { // if using scope
			if confirmedDates[startDate] == nil || confirmedDates[endDate] == nil {
				status := http.StatusBadRequest
				http.Error(w, "Wrong date format in scope. example of valid date: 2020-12-01-2021-01-31", status)
				return
			}
			confirmed = confirmedDates[endDate].(float64) - confirmedDates[startDate].(float64)
			recovered = recoveredDates[endDate].(float64) - recoveredDates[startDate].(float64)
			response.Scope = startDate + "-" + endDate
		} else { // if not using scope
			for k, v := range confirmedDates { // loops through confirmed dates and find the highest value
				value := v.(float64)
				if value > confirmed {
					confirmed = value
					recovered = recoveredDates[k].(float64)
				}
			}
			response.Scope = "total"
		}

		//response
		response.Continent = confirmedData.All["continent"].(string)
		response.Confirmed = int(confirmed)
		response.Country = countryName
		response.Recovered = int(recovered)
		percentPlaceholder := 100 / (confirmedData.All["population"].(float64) / confirmed)
		response.Population_percentage = math.Floor(percentPlaceholder*100) / 100
		json.NewEncoder(w).Encode(response)
		return
	default:
		return
	}
}

// http://localhost:8080/corona/v1/policy/{:country_name}{?scope=begin_date-end_date}
func HandleStringencyTrends(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// get request
	case http.MethodGet:
		http.Header.Add(w.Header(), "content-type", "application/json")

		// gets information in url parameter
		countryName, startDate, endDate, err := getUrlData("policy", r)
		if err != nil { // if error getting url data (invalid url)
			status := http.StatusBadRequest
			http.Error(w, err.Error(), status)
			return
		}

		// gets country code
		countryCode, err := getCountryCodeByName(countryName)
		if err != nil {
			status := http.StatusBadRequest
			http.Error(w, err.Error(), status)
			return
		}

		// gets stringency data on from date
		dataFromDate, err := getStringencyData(countryCode, startDate)
		if err != nil {
			status := http.StatusBadRequest
			http.Error(w, err.Error(), status)
			return
		}

		// gets stringency data on end date
		dataEndDate, err := getStringencyData(countryCode, endDate)
		if err != nil {
			status := http.StatusBadRequest
			http.Error(w, err.Error(), status)
			return
		}

		// response
		var response PolicyStringencyTrends
		response.Country = countryName
		response.Scope = startDate + "-" + endDate
		response.Stringency = dataEndDate.StringencyData.Stringency
		response.Trend = dataEndDate.StringencyData.Stringency - dataFromDate.StringencyData.Stringency
		json.NewEncoder(w).Encode(response)
		return
	default:
		return
	}
}

// http://localhost:8080/corona/v1/notifications/{id}
func HandleNotification(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		http.Header.Add(w.Header(), "content-type", "application/json")
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 5 { // if wrong url format
			http.Error(w, "Wrong format, should be '/corona/v1/notifications/{id}", http.StatusBadRequest)
		}

		id := parts[4]
		if id == "" { // if no id in url parameter
			allWebhooks, err := getWebhooks() // gets all webhooks
			if err != nil {
				http.Error(w, "Something went wrong: "+err.Error(), http.StatusBadRequest)
			}
			json.NewEncoder(w).Encode(allWebhooks)
		} else { // if id in url parameter
			webhook, err := getSingleWebhook(id) // gets webhook with id from url parameter
			if err != nil {
				http.Error(w, "Something went wrong: "+err.Error(), http.StatusBadRequest)
			}
			json.NewEncoder(w).Encode(webhook)
		}

		return
	case http.MethodPost:
		http.Header.Add(w.Header(), "content-type", "text/plain")
		var webhookRegistration WebhookRegistration
		err := json.NewDecoder(r.Body).Decode(&webhookRegistration) // gets data from body
		if err != nil {
			http.Error(w, "Something went wrong: "+err.Error(), http.StatusBadRequest)
			return
		}

		err = isValidData(webhookRegistration) // checks if the fields in body is valid
		if err != nil {
			http.Error(w, "Something went wrong: "+err.Error(), http.StatusBadRequest)
			return
		}
		var occurrences float64 = 0.0
		if webhookRegistration.Field == "stringency" { // if webhook for stringency
			// gets country code
			countryCode, err := getCountryCodeByName(webhookRegistration.Country)
			if err != nil {
				http.Error(w, "Something went wrong: "+err.Error(), http.StatusBadRequest)
				return
			}

			currentDate := time.Now().Local()
			tenDaysAgoDate := currentDate.AddDate(0, 0, -10).Format("2006-01-02") // calculates the date 10 days ago
			stringencyData, err := getStringencyData(countryCode, tenDaysAgoDate)
			if err != nil {
				http.Error(w, "Something went wrong: "+err.Error(), http.StatusBadRequest)
				return
			}
			occurrences = stringencyData.StringencyData.Stringency
		} else { // if webhook for confirmed cases
			confirmedData, err := getConfirmedData(webhookRegistration.Country)
			if err != nil {
				http.Error(w, "Something went wrong: "+err.Error(), http.StatusBadRequest)
				return
			}
			for _, v := range confirmedData.All["dates"].(map[string]interface{}) { // loops through confirmed dates and find the highest value
				value := v.(float64)
				if value > occurrences {
					occurrences = value
				}
			}
		}

		// adding the webhook to the webhooks collection in cloud firestore
		docRef, _, err := client.Collection("webhooks").Add(ctx,
			map[string]interface{}{
				"url":         webhookRegistration.Url,
				"timeout":     webhookRegistration.Timeout,
				"field":       webhookRegistration.Field,
				"country":     webhookRegistration.Country,
				"trigger":     webhookRegistration.Trigger,
				"occurrences": occurrences,
				"time":        firestore.ServerTimestamp,
			})
		if err != nil {
			http.Error(w, "Something went wrong: "+err.Error(), http.StatusBadRequest)
			return
		}
		w.Write([]byte(docRef.ID)) // response (ID generated on the document created)
		return

	case http.MethodDelete:
		http.Header.Add(w.Header(), "content-type", "text/plain")
		parts := strings.Split(r.URL.Path, "/")
		// if wrong format
		if len(parts) != 5 {
			http.Error(w, "Wrong format, should be '/corona/v1/notifications/{id}", http.StatusBadRequest)
			return
		}

		// deletes the webhook document from cloud firestore with ID from url
		id := parts[4]
		err := deleteSingleWebhook(id)
		if err != nil {
			http.Error(w, "Something went wrong: "+err.Error(), http.StatusBadRequest)
			return
		}
		w.Write([]byte("If the webhook existed, it is now deleted"))
		return
	default:
		return
	}
}

// http://localhost:8080/corona/v1/diag/
func HandleDiag(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		http.Header.Add(w.Header(), "content-type", "application/json")
		var response Diag
		// gets all webhooks
		allWebhooks, err := getWebhooks()
		if err != nil {
			http.Error(w, "Something went wrong: "+err.Error(), http.StatusBadRequest)
			return
		}
		registered := len(allWebhooks) // number of webhooks

		// checks status codes
		mmediagroupapi := checkStatusCodeApi(r, "https://covid-api.mmediagroup.fr/v1/history?country=Norway&status=Confirmed")
		covidtrackerapi := checkStatusCodeApi(r, "https://covidtrackerapi.bsg.ox.ac.uk/api/v2/stringency/actions/NOR/2020-10-10")

		// response
		response.Mmediagroupapi = mmediagroupapi
		response.Covidtrackerapi = covidtrackerapi
		response.Registered = registered
		response.Version = "v1"
		response.Uptime = getServerUptime()
		json.NewEncoder(w).Encode(response)
		return
	default:
		return
	}
}
