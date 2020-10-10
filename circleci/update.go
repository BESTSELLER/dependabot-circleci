package circleci

import (
	"strings"

	"gopkg.in/yaml.v3"
)

func GetUpdates(node *yaml.Node) map[string]*yaml.Node {
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

	return orbUpdates
}

func ReplaceVersion(orb *yaml.Node, oldVersion string, content string) string {

	lines := strings.Split(content, "\n")
	lineNumber := orb.Line + -1
	theLine := lines[lineNumber]
	lines[lineNumber] = strings.ReplaceAll(theLine, oldVersion, orb.Value)

	output := strings.Join(lines, "\n")

	return output
}
