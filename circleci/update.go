package circleci

import (
	"strings"

	"gopkg.in/yaml.v3"
)

func getDockerUpdates(node *yaml.Node) map[string]*yaml.Node {
	updates := map[string]*yaml.Node{}

	for i, nextHole := range node.Content {
		if nextHole.Value == "executors" || nextHole.Value == "jobs" {

			// check if there is a docker image
			if i+1 >= len(node.Content) {
				return updates
			}

			dockers := node.Content[i+1]
			updates := extractImages(dockers.Content)
			for k, v := range updates {
				updates[k] = v
			}
			return updates
		}

		next := getDockerUpdates(nextHole)
		for k, v := range next {
			updates[k] = v
		}
	}

	return updates
}
func getOrbUpdates(node *yaml.Node) map[string]*yaml.Node {
	updates := map[string]*yaml.Node{}

	for i, nextHole := range node.Content {
		if nextHole.Value == "orbs" {
			orbs := node.Content[i+1]
			updates := extractOrbs(orbs.Content)
			for k, v := range updates {
				updates[k] = v
			}
			return updates
		}

		next := getOrbUpdates(nextHole)
		for k, v := range next {
			updates[k] = v
		}
	}

	return updates
}

// GetUpdates returns a list of updated yaml nodes
func GetUpdates(node *yaml.Node) (map[string]*yaml.Node, map[string]*yaml.Node) {
	return getOrbUpdates(node), getDockerUpdates(node)
}

// ReplaceVersion replaces a specific line in the yaml
func ReplaceVersion(orb *yaml.Node, oldVersion string, content string) string {

	lines := strings.Split(content, "\n")
	lineNumber := orb.Line + -1
	theLine := lines[lineNumber]
	lines[lineNumber] = strings.ReplaceAll(theLine, oldVersion, orb.Value)

	output := strings.Join(lines, "\n")

	return output
}
