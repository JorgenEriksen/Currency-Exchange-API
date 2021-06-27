package CoronaAPI

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Diag struct {
	Mmediagroupapi  string
	Covidtrackerapi string
	Registered      int
	Version         string
	Uptime          float64
}

var serverStartTime time.Time

func ServerStart() {
	serverStartTime = time.Now()
}

// returns how long the server has been up in seconds
func getServerUptime() float64 {
	return time.Since(serverStartTime).Seconds()
}

// returns status code from api
func checkStatusCodeApi(r *http.Request, url string) string {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Errorf("Error in creating request:", err.Error())
	}
	client := &http.Client{}
	res, err := client.Do(request)
	statusCode := 0
	if res != nil {
		statusCode = res.StatusCode
	} else {
		statusCode = http.StatusBadRequest
	}
	return strconv.Itoa(statusCode)
}
