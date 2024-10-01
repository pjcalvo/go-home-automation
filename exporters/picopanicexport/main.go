package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
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

type metrics struct {
	pulse  *pulseCheck
	expire time.Time
	sync.RWMutex
}

func (m *metrics) getMetrics(client *http.Client, baseUrl string) *metrics {
	m.Lock()
	defer m.Unlock()

	if time.Now().Before(m.expire) {
		return m
	}

	if err := m.pulse.getPulse(client, baseUrl); err != nil {
		m.pulse.Up = false
	}

	m.expire = time.Now().Add(2 * time.Second)

	return m
}

func (m *metrics) pulseStatus() float64 {
	m.RLock()
	defer m.RUnlock()

	if m.pulse.Up {
		return 1
	}
	return 0
}

// newPoller pollsForPanics in the pico
func newPoller(baseUrl string, delay time.Duration) {
	// new object panic
	panicCheck := panicCheck{
		Panic: false,
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	slog.Info("Panic checker started ...")

	for {
		time.Sleep(delay)
		slog.Info("Checking panic...")
		err := panicCheck.getPanic(client, baseUrl)
		if err != nil {
			slog.With(err).Error("something went wrong")
			continue
		}
		if !panicCheck.Panic {
			slog.Info("No panic has ocurred...")
			continue
		}

		slog.Info("Panic ACTIONED, sending webhook...")
		err = triggerWebhook(&panicCheck)
		if err != nil {
			slog.With(err).Error("something went wrong")
		}
	}
}

// newMux for accepting incoming requests
func newMux(baseUrl string) http.Handler {
	mux := http.NewServeMux()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	m := &metrics{
		pulse: &pulseCheck{},
	}

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "pico_panic",
			Help: "Pico Panic Checker.", ConstLabels: prometheus.Labels{"status": "bool"},
		},
		func() float64 {
			return m.getMetrics(client, baseUrl).pulseStatus()
		},
	)

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

	go newPoller(picoURL, 15*time.Second)

	if err := s.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
