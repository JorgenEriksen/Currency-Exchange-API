package main

import (
	"CoronaAPI"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	CoronaAPI.InitFirebase()
	CoronaAPI.ServerStart()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", CoronaAPI.HandleRoot)
	http.HandleFunc("/corona/v1/country/", CoronaAPI.HandleCases)
	http.HandleFunc("/corona/v1/policy/", CoronaAPI.HandleStringencyTrends)
	http.HandleFunc("/corona/v1/notifications/", CoronaAPI.HandleNotification)
	http.HandleFunc("/corona/v1/diag/", CoronaAPI.HandleDiag)
	go CoronaAPI.WebhookRoutine() // webhook check that runs every hour

	fmt.Println("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
