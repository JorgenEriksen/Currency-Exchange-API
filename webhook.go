package CoronaAPI

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type WebhookRegistration struct {
	ID          string
	Url         string    `json:"url"`
	Timeout     float64   `json:"timeout"`
	Time        time.Time `json:"time"`
	Field       string    `json:"field"`
	Country     string    `json:"country"`
	Trigger     string    `json:"trigger"`
	Occurrences float64   `json:"occurrences"`
}

var ctx context.Context
var client *firestore.Client

// initialize firebase/firestore
func InitFirebase() {
	ctx = context.Background()
	opt := option.WithCredentialsFile("./serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalln(err)
	}
	client, err = app.Firestore(ctx)
}

// gets a single webhook data
func getSingleWebhook(id string) (WebhookRegistration, error) {
	data, err := client.Collection("webhooks").Doc(id).Get(ctx)
	var webhook WebhookRegistration
	if err != nil {
		return webhook, errors.New("Can't find webhook with id " + id)
	}
	webhook = mapToWebhookStruct(data.Data(), data.Ref.ID)
	return webhook, nil
}

// deletes a single webhook
func deleteSingleWebhook(id string) error {
	_, err := client.Collection("webhooks").Doc(id).Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

// gets all webhooks
func getWebhooks() ([]WebhookRegistration, error) {
	iter := client.Collection("webhooks").Documents(ctx)
	var allWebooks []WebhookRegistration
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return allWebooks, err
		}
		webhook := mapToWebhookStruct(doc.Data(), doc.Ref.ID)
		allWebooks = append(allWebooks, webhook)
	}
	return allWebooks, nil
}

// check if webhook data is valid
func isValidData(webhookData WebhookRegistration) error {
	if len(webhookData.Url) < 1 {
		return errors.New("Missing or invalid url data")
	} else if webhookData.Timeout < 1 {
		return errors.New("Missing or invalid timeout data")
	} else if webhookData.Field != "stringency" && webhookData.Field != "confirmed" {
		return errors.New("Missing or invalid field data")
	} else if len(webhookData.Country) < 1 {
		return errors.New("Missing or invalid country data")
	} else if webhookData.Trigger != "ON_CHANGE" && webhookData.Trigger != "ON_TIMEOUT" && webhookData.Trigger != "ON_UPDATE" {
		return errors.New("Missing or invalid trigger data")
	}
	return nil
}

// webhook routine
func WebhookRoutine() {
	webhooks, err := getWebhooks()
	if err != nil {
		log.Fatalln(err)
	}
	currentDate := time.Now().Local()
	tenDaysAgoDate := currentDate.AddDate(0, 0, -10).Format("2006-01-02") // gets the date 10 days ago

	// for each webhook
	for i := 0; i < len(webhooks); i++ {
		currentTime := time.Now().Local()
		var whenToNotificate time.Time = webhooks[i].Time.Add(time.Duration(webhooks[i].Timeout) * time.Minute) // gets the date for when to notificate
		if webhooks[i].Field == "stringency" {                                                                  // if stringency webhook
			// gets country code
			countryCode, err := getCountryCodeByName(webhooks[i].Country)
			if err != nil {
				log.Fatalln(err)
			}

			// gets stringency data
			stringencyData, err := getStringencyData(countryCode, tenDaysAgoDate)
			if err != nil {
				log.Fatalln(err)
			}

			// check if it's time to notificate
			if (stringencyData.StringencyData.Stringency != webhooks[i].Occurrences || webhooks[i].Trigger == "ON_TIMEOUT") && currentTime.After(whenToNotificate) {
				webhooks[i].Occurrences = stringencyData.StringencyData.Stringency                   // so the newest data is in the notification
				sendNotification(webhooks[i])                                                        // sends notification
				updateWebhook(webhooks[i].ID, currentTime, stringencyData.StringencyData.Stringency) // update webhook in firestore
			}
		} else if webhooks[i].Field == "confirmed" { // if confirmed webhook
			confirmedData, err := getConfirmedData(webhooks[i].Country)
			if err != nil {
				log.Fatalln(err)
			}
			// finds the highest occurrences of confirmed
			var highestOccurrences float64 = 0.0
			for _, v := range confirmedData.All["dates"].(map[string]interface{}) { // loops through confirmed dates and find the highest value
				value := v.(float64)
				if value > highestOccurrences {
					highestOccurrences = value
				}
			}
			// check if it's time to notificate
			if (highestOccurrences != webhooks[i].Occurrences || webhooks[i].Trigger == "ON_TIMEOUT") && currentTime.After(whenToNotificate) {
				webhooks[i].Occurrences = highestOccurrences                   // so the newest data is in the notification
				sendNotification(webhooks[i])                                  // sends notification
				updateWebhook(webhooks[i].ID, currentTime, highestOccurrences) // update webhook in firestore
			}
		}
	}
	time.Sleep(time.Duration(3600) * time.Second) // runs every hour
	go WebhookRoutine()
}

// sends notification to client
func sendNotification(webhook WebhookRegistration) {
	json, err := json.Marshal(webhook)
	if err != nil {
		log.Fatalln("An error has occurred: %s", err)
	}
	req, err := http.NewRequest(http.MethodPost, webhook.Url, bytes.NewReader([]byte(json))) // actualyl sends post request to klient
	req.Header.Add("content-type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error in HTTP request: " + err.Error())
		return
	}
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Something is wrong with invocation response: " + err.Error())
		return
	}
}

// updates webhook in firestore
func updateWebhook(id string, newTime time.Time, newOccurences float64) {
	_, err := client.Collection("webhooks").Doc(id).Set(ctx, map[string]interface{}{
		"time":        newTime,
		"occurrences": newOccurences,
	}, firestore.MergeAll)
	if err != nil {
		log.Fatalln("An error has occurred: %s", err)
	}
}
