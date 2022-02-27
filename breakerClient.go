package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

var cb *gobreaker.CircuitBreaker

func init() {
	var settings gobreaker.Settings
	settings.Name = "Breaker"
	settings.ReadyToTrip = func(counts gobreaker.Counts) bool {
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 10 && failureRatio >= 0.2
	}
	// settings.Timeout = time.Second
	// defaults timeout is 60 seconds. State machine doesn't recheck the site for a minute saving resources
	settings.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
		if to == gobreaker.StateOpen {
			fmt.Println("Transitioning:: to Open State =>SITE IS DOWN")
		}
		if from == gobreaker.StateOpen && to == gobreaker.StateHalfOpen {
			fmt.Println("Transitioning:: Open --> Half Open =>RE-CHECKING")
		}
		if from == gobreaker.StateHalfOpen && to == gobreaker.StateClosed {
			fmt.Println("Transitioning::Half Open --> Closed =>SITE IS BACK UP")
		}
	}
	cb = gobreaker.NewCircuitBreaker(settings)
}

func Get(url string) ([]byte, error) {
	body, err := cb.Execute(func() (interface{}, error) {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return body, nil
	})
	if err != nil {
		return nil, err
	}

	return body.([]byte), nil
}

func main() {
	url := "http://localhost:8000"
	var body []byte
	var err error
	for {
		body, err = Get(url)
		if err != nil {
			fmt.Printf("%s is DOWN. Handle it accordingly now...\n", url)
		} else {
			fmt.Println(string(body))
		}
		time.Sleep(time.Second)
	}

}
