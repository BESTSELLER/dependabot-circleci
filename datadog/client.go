package datadog

import (
	"fmt"
	"os"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/rs/zerolog/log"
)

var client *statsd.Client
var metricPrefix = "dependabot_circleci"

// CreateClient creates a statsd client
func CreateClient() (err error) {
	client, err = statsd.New(os.Getenv("DD_AGENT_HOST"))
	if err != nil {
		return err
	}
	return nil
}

// IncrementCount incrementes a counter based on the input
func IncrementCount(metricName string, org string) {
	err := client.Incr(
		fmt.Sprintf("%s.%s", metricPrefix, metricName),
		[]string{
			"organistation:" + org,
		},
		1)
	if err != nil {
		log.Debug().Err(err).Msgf("could increment datadog counter %s", metricName)
	}
}
