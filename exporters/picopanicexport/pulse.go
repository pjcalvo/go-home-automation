package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type pulseCheck struct {
	Up bool `json:"up"`
}

func (pC *pulseCheck) getPulse(client *http.Client, baseUrl string) error {
	url := fmt.Sprintf("%s/pulse", baseUrl)
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
