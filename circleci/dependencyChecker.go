package circleci

import (
	"fmt"
	"log"
	"strings"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/CircleCI-Public/circleci-cli/api/graphql"
	"gopkg.in/yaml.v3"
)

func GetUpdates(node *yaml.Node) map[string]*yaml.Node {
	//fmt.Printf("start: %d\n", len(orbUpdates))
	orbUpdates := map[string]*yaml.Node{}

	for i, nextHole := range node.Content {
		if nextHole.Value == "orbs" {
			orbs := node.Content[i+1]
			orbUpdates := extractOrbs(orbs.Content)
			for k, v := range orbUpdates {
				orbUpdates[k] = v
			}
			return orbUpdates
		}

		// *** ready for docker image check ***
		// if nextHole.Value == "executors" {
		// 	orbs := node.Content[i+1]
		// 	extractImages(orbs.Content)
		// }
		// if nextHole.Value == "jobs" {
		// 	orbs := node.Content[i+1]
		// 	extractImages(orbs.Content)
		// }

		next := GetUpdates(nextHole)
		for k, v := range next {
			orbUpdates[k] = v
		}
	}

	//fmt.Printf("end: %d\n", len(orbUpdates))
	return orbUpdates
}

func ReplaceVersion(orb *yaml.Node, oldVersion string, content string) string {

	lines := strings.Split(content, "\n")
	lineNumber := orb.Line + -1
	theLine := lines[lineNumber]
	lines[lineNumber] = strings.ReplaceAll(theLine, oldVersion, orb.Value)

	output := strings.Join(lines, "\n")
	fmt.Println(output)

	return output
}

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

func extractImages(orbs []*yaml.Node) {
	for i := 0; i < len(orbs); i++ {
		orb := orbs[i]
		if orb.Value == "image" {
			orb = orbs[i+1]

			orb.Value = findNewestDockerVersion(orb.Value)
		}
		extractImages(orb.Content)
	}
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

func findNewestDockerVersion(currentVersion string) string {

	fmt.Println(currentVersion)

	current := strings.Split(currentVersion, ":")

	fmt.Println(current)
	fmt.Println(len(current))

	// check if image has no version tag
	if len(current) == 1 {
		return currentVersion
	}

	// check if tag is latest
	if strings.ToLower(current[1]) == "latest" {
		return currentVersion
	}

	// define registry
	registry := "registry.hub.docker.com"
	registryImage := strings.Split(current[0], "/")

	if len(registryImage) > 2 {
		registry = fmt.Sprintf("%s", registryImage[0])
	}

	fmt.Println(registry)

	// query that damn registry for newer versions

	return "latest"
	// This one is a bit tricky actually! Watchtower seems to do this by utilising a docker client, but then we need
	// Docker in docker i guess ? Maybe there is a smart api endpoint, all registries should use the same to communicate with docker i guess ?
}
