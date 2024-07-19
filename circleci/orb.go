package circleci

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/CircleCI-Public/circleci-cli/api/graphql"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func extractOrbs(orbs []*yaml.Node, parameters *map[string]*yaml.Node) map[string]*yaml.Node {
	updates := map[string]*yaml.Node{}
	for i := 0; i < len(orbs); i = i + 2 {
		orb := orbs[i+1]

		log.Debug().Msg(fmt.Sprintf("current orb version: %s", orb.Value))
		orbRoot, currentVer, newestVer := findNewestOrbVersion(orb.Value, parameters)
		log.Debug().Msg(fmt.Sprintf("new orb version: %s@%s", orbRoot, newestVer))

		if currentVer != newestVer {
			oldVersion := orb.Value
			orb.Value = newestVer
			updates[oldVersion] = orb
		}
	}
	return updates
}

func findNewestOrbVersion(currentVersion string, parameters *map[string]*yaml.Node) (orbName, currentTag, newestTag string) {
	orbSplitString := strings.Split(currentVersion, "@")
	// check if orb is always updated
	if orbSplitString[1] == "volatile" || strings.HasPrefix(orbSplitString[1], "dev:") {
		return orbSplitString[0], orbSplitString[1], orbSplitString[1]
	}
	orbName = orbSplitString[0]
	currentTag = orbSplitString[1]

	if newestTag, hit := cache[currentVersion]; hit {
		log.Debug().Msgf("Using cached version for orb: %s - %s", orbName, newestTag)
		return orbName, currentTag, newestTag
	}
	if param := ExtractParameterName(currentVersion); len(param) > 0 {
		paramDefault, found := (*parameters)[param]
		if !found {
			log.Debug().Msgf("Parameter %s not found in parameters", param)
			return orbName, currentTag, currentTag
		}
		currentTag = paramDefault.Value
	}

	CCIApiToken := ""
	if config.AppConfig.BestsellerSpecific.Running {
		log.Debug().Msg("Using Bestseller specific token to handle private orbs")
		CCIApiToken = config.AppConfig.BestsellerSpecific.Token
	}

	client := graphql.NewClient(http.DefaultClient, "https://circleci.com/", "graphql-unstable", CCIApiToken, false)

	// if requests fails, return current version
	orbInfo, err := api.OrbInfo(client, orbName)
	if err != nil {
		log.Error().Err(err).Msgf("error finding latests orb version failed for orb: %s", orbSplitString[0])
		return orbName, currentTag, currentTag
	}

	if len(orbInfo.Orb.HighestVersion) == 0 || strings.HasPrefix(orbInfo.Orb.HighestVersion, currentTag) {
		cache[currentVersion] = currentTag
		return orbName, currentTag, currentTag
	}

	newVersion := TrimSemver(currentTag, orbInfo.Orb.HighestVersion)
	cache[currentVersion] = newVersion
	return orbName, currentTag, newVersion
}
