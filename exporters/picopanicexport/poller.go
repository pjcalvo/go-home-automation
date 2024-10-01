package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type panicCheck struct {
	Panic     bool      `json:"panic"`
	Timestamp time.Time `json:"timestamp"`
}

func (pC *panicCheck) getPanic(client *http.Client, baseUrl string) error {
	url := fmt.Sprintf("%s/panic", baseUrl)
	response, err := client.Get(url)
	if err != nil {
		log.Println(err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf("%w: invalid status code: %s", errInvalidResponse, response.Status)
		log.Println(err)
		return err
	}

	if err := json.NewDecoder(response.Body).Decode(pC); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
