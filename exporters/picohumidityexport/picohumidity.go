package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	//go:embed rootPage.html
	rootPageHTML []byte

	errInvalidResponse = errors.New("unexpected response from server")
)

type humidityValue struct {
	Value float64 `json:"value"`
}

func (tv *humidityValue) getHumidityValues(client *http.Client, url string) error {
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

	if err := json.NewDecoder(response.Body).Decode(tv); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type metrics struct {
	results *humidityValue
	up      float64
	expire  time.Time
	sync.RWMutex
}

func (m *metrics) getMetrics(client *http.Client, url string) *metrics {
	m.Lock()
	defer m.Unlock()

	if time.Now().Before(m.expire) {
		return m
	}

	m.up = 1
	if err := m.results.getHumidityValues(client, url); err != nil {
		m.up = 0
		m.results.Value = 0
	}

	m.expire = time.Now().Add(2 * time.Second)

	return m
}

func (m *metrics) value() float64 {
	m.RLock()
	defer m.RUnlock()

	return m.results.Value
}

func (m *metrics) status() float64 {
	m.RLock()
	defer m.RUnlock()
	return m.up
}

func newMux(url string) http.Handler {
	mux := http.NewServeMux()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	m := &metrics{
		results: &humidityValue{},
	}

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "pico_humidity",
			Help: "Pico Sensor Humidity.", ConstLabels: prometheus.Labels{"unit": "humidity"},
			Subsystem: "monstera_sensor",
		},
		func() float64 {
			return m.getMetrics(client, url).value()
		},
	)
	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:      "pico_up",
			Help:      "Pico Sensor Server Status.",
			Subsystem: "monstera_sensor",
		},
		func() float64 {
			return m.getMetrics(client, url).status()
		})

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write(rootPageHTML)
	})

	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

func main() {
	picoURL := os.Getenv("PICO_SERVER_URL")
	s := &http.Server{
		Addr:    ":3030",
		Handler: newMux(picoURL), ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second,
	}
	if err := s.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
