package circleci

import (
	"fmt"
	"log"
	"strings"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/CircleCI-Public/circleci-cli/api/graphql"
	"gopkg.in/yaml.v3"
)

func extractOrbs(orbs []*yaml.Node) map[string]*yaml.Node {
	updates := map[string]*yaml.Node{}
	for i := 0; i < len(orbs); i = i + 2 {
		orb := orbs[i+1]
		orbVersion := findNewestOrbVersion(orb.Value)

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
	if orbSplitString[1] == "volatile" {
		return "volatile"
	}

	client := graphql.NewClient("https://circleci.com/", "graphql-unstable", "", false)

	// if requests fails, return current version
	orbInfo, err := api.OrbInfo(client, orbSplitString[0])
	if err != nil {
		log.Printf("finding latests orb version failed: %v", err)
		return fmt.Sprintf("%s@%s", orbSplitString[0], orbSplitString[1])
	}

	return fmt.Sprintf("%s@%s", orbSplitString[0], orbInfo.Orb.HighestVersion)
}
