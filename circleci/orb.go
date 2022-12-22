package circleci

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/CircleCI-Public/circleci-cli/api/graphql"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func extractOrbs(orbs []*yaml.Node) map[string]*yaml.Node {
	updates := map[string]*yaml.Node{}
	for i := 0; i < len(orbs); i = i + 2 {
		orb := orbs[i+1]

		log.Debug().Msg(fmt.Sprintf("current orb version: %s", orb.Value))
		orbVersion := findNewestOrbVersion(orb.Value)
		log.Debug().Msg(fmt.Sprintf("new orb version: %s", orbVersion))

		if orb.Value != orbVersion {
			oldVersion := orb.Value
			orb.Value = orbVersion
			updates[oldVersion] = orb
		}
	}
	return updates
}

func findNewestOrbVersion(orb string) string {

	orbSplitString := strings.Split(orb, "@")

	// check if orb is always updated
	if orbSplitString[1] == "volatile" || strings.HasPrefix(orbSplitString[1], "dev:") {
		return orbSplitString[1]
	}

	client := graphql.NewClient(http.DefaultClient, "https://circleci.com/", "graphql-unstable", "", false)

	// if requests fails, return current version
	orbInfo, err := api.OrbInfo(client, orbSplitString[0])
	if err != nil {
		log.Error().Err(err).Msgf("error finding latests orb version failed for orb: %s", orbSplitString[0])
		return fmt.Sprintf("%s@%s", orbSplitString[0], orbSplitString[1])
	}

	return fmt.Sprintf("%s@%s", orbSplitString[0], orbInfo.Orb.HighestVersion)
}
