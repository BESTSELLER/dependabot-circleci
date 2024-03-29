package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/rs/zerolog/log"
)

type DataDog struct {
	Series []Series `json:"series"`
}

type Series struct {
	Metric   string    `json:"metric"`
	Points   [][]int64 `json:"points"`
	Tags     []string  `json:"tags"`
	Type     string    `json:"type"`
	Host     string    `json:"host"`
	Interval int64     `json:"interval"`
}

var metricPrefix = "dependabot_circleci"

// IncrementCount incrementes a counter based on the input
func IncrementCount(metricName string, value int64, tags []string) {
	metric := metricPrefix + "." + metricName
	err := postDataDogMetric(metric, value, "count", tags)
	if err != nil {
		log.Debug().Err(err).Msgf("Could not increment datadog, metric: %s, value: %d, tags: %s", metricName, value, tags)
	}
}

// Gauge incrementes a counter based on the input
func Gauge(metricName string, value float64, tags []string) {
	metric := metricPrefix + "." + metricName
	err := postDataDogMetric(metric, int64(value), "gauge", tags)
	if err != nil {
		log.Debug().Err(err).Msgf("could not send gauge to datadog, metric: %s, value: %d, tags: %s", metricName, int64(value), tags)
	}
}

// Distribution ...
func Distribution(metricName string, value float64, tags []string) {
	metric := metricPrefix + "." + metricName
	err := postDataDogMetric(metric, int64(value), "distribution", tags)
	if err != nil {
		log.Debug().Err(err).Msgf("could not send distribution to datadog, metric: %s, value: %d, tags: %s", metricName, int64(value), tags)
	}

}

func postDataDogMetric(metric string, value int64, metricType string, tags []string) error {
	apiKey := config.AppConfig.Datadog.APIKey

	url := "https://api.datadoghq.eu/api/v1/series?api_key=" + apiKey

	series := Series{
		Metric: metric,
		Type:   metricType,
		Tags:   tags,
	}

	point := [][]int64{
		{time.Now().Unix(), value},
	}
	series.Points = point

	payload := DataDog{[]Series{series}}

	_, err := postStructAsJSON(url, payload, nil)
	if err != nil {
		return err
	}

	return nil
}

func postStructAsJSON(url string, payload interface{}, target interface{}) (string, error) {
	var myClient = http.Client{}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url, b)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	r, err := myClient.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	// check status code
	bodyBytes, _ := io.ReadAll(r.Body)
	bodyString := string(bodyBytes)

	if r.StatusCode < 200 || r.StatusCode > 299 {
		return "", fmt.Errorf("request failed, expected status: 2xx got: %d, error message: %s", r.StatusCode, bodyString)
	}

	// only decode if target is not nil
	if target != nil {
		decode := json.NewDecoder(r.Body)
		err = decode.Decode(&target)
		if err != nil {
			return "", err
		}
	}

	return bodyString, nil
}

// TimeTrackAndHistogram logs the ammount of time it take for a function to execute and send it as a Gauge to datadog.
func TimeTrackAndGauge(metric string, tags []string, start time.Time) {
	elapsed := time.Since(start)
	Gauge(metric, float64(elapsed.Milliseconds()), tags)
}
