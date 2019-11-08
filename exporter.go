package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	leds "github.com/hodgesds/goleds"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace            = "led"
	exporter             = "led_exporter"
	eventConfigNamespace = "events"
)

var port int

func init() {
	flag.IntVar(&port, "port", 9342, "HTTP port")
}

// LEDCollector is an interface that embeds the prometheus Collector interface.
type LEDCollector interface {
	prometheus.Collector
}

type ledCollector struct {
	leds          []*leds.LED
	brightness    *prometheus.Desc
	maxBrightness *prometheus.Desc
}

// NewLEDCollector returns a new PerfCollector.
func NewLEDCollector() (LEDCollector, error) {
	leds, err := leds.LEDs()
	if err != nil {
		return nil, err
	}

	brightness := prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			"led",
			"brightness",
		),
		"LED brightness",
		[]string{"led"},
		nil,
	)

	maxBrightness := prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			"led",
			"max_brightness",
		),
		"LED max brightness",
		[]string{"led"},
		nil,
	)

	return &ledCollector{
		brightness:    brightness,
		maxBrightness: maxBrightness,
		leds:          leds,
	}, nil
}

// Describe implements the prometheus.Collector interface.
func (c *ledCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.brightness
	ch <- c.maxBrightness
}

// Collect implements prometheus.Collector interface.
func (c *ledCollector) Collect(ch chan<- prometheus.Metric) {
	for _, led := range c.leds {
		n := name(led.Name())
		brightness, err := led.Brightness()
		if err != nil {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.brightness,
			prometheus.GaugeValue,
			float64(brightness),
			n,
		)
		maxBrightness, err := led.MaxBrightness()
		if err != nil {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.maxBrightness,
			prometheus.GaugeValue,
			float64(maxBrightness),
			n,
		)
	}
}

func name(s string) string {
	return strings.Replace(strings.Replace(s, ":", "_", -1), "-", "_", -1)
}

func main() {
	flag.Parse()
	collector, err := NewLEDCollector()
	if err != nil {
		log.Fatal(err)
	}

	prometheus.MustRegister(collector)

	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>LED Exporter</title></head>
             <body>
             <h1>LED Exporter</h1>
             <p><a href=/metrics>Metrics</a></p>
             </body>
             </html>`))
	})
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}

}
