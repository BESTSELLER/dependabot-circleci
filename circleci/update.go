package circleci

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Update struct {
	Type        string
	CurrentName string
	FileUpdates map[string]FileUpdate
}

func (update Update) SplitCurrentVersion() []string {
	var oldVersion []string
	var separator string
	if update.Type == "orb" {
		separator = "@"
	} else {
		separator = ":"
	}
	oldVersion = strings.Split(update.CurrentName, separator)
	return oldVersion
}

type FileUpdate struct {
	SHA        *string
	Content    *string
	Parameters *map[string]*yaml.Node
	Node       *yaml.Node
}

var cache = map[string]string{}

func getDockerUpdates(node *yaml.Node, parameters *map[string]*yaml.Node) map[string]*yaml.Node {
	updates := map[string]*yaml.Node{}

	for i, nextHole := range node.Content {
		switch nextHole.Value {
		case "executors":
		case "jobs":
		case "docker":
			// check if there is a docker image
			if i+1 >= len(node.Content) {
				return updates
			}
			dockers := node.Content[i+1]
			updates := extractImages(dockers.Content, parameters)
			for k, v := range updates {
				updates[k] = v
			}
			return updates
		}

		next := getDockerUpdates(nextHole, parameters)
		for k, v := range next {
			updates[k] = v
		}
	}

	return updates
}

// Recurse until we find a block called parameters then we extract a map with the default values
func extractParameters(yamlNode []*yaml.Node) map[string]*yaml.Node {
	for a, nextHole := range yamlNode {
		if nextHole.Value == "parameters" {
			if parametersBlock := yamlNode[a+1].Content; len(parametersBlock) > 0 {
				results := map[string]*yaml.Node{}
				for pi, param := range parametersBlock {
					if len(param.Value) > 0 {
						for c, k := range parametersBlock[pi+1].Content {
							if k.Value == "default" {
								results[param.Value] = parametersBlock[pi+1].Content[c+1]
							}
						}
					}
				}
				return results
			}
		} else if len(nextHole.Content) > 0 {
			if results := extractParameters(nextHole.Content); results != nil {
				return results
			}
		}
	}
	return nil
}

func getOrbUpdates(node *yaml.Node, parameters *map[string]*yaml.Node) map[string]*yaml.Node {
	updates := map[string]*yaml.Node{}

	for i, nextHole := range node.Content {
		if nextHole.Value == "orbs" {
			orbs := node.Content[i+1]
			updates := extractOrbs(orbs.Content, parameters)
			for k, v := range updates {
				updates[k] = v
			}
			return updates
		}

		next := getOrbUpdates(nextHole, parameters)
		for k, v := range next {
			updates[k] = v
		}
	}

	return updates
}

// ScanFileUpdates returns a map of updates found in a file, the key is the original version
func ScanFileUpdates(updates *map[string]Update, content, path, SHA *string) error {
	// unmarshal
	var nodeContent yaml.Node
	err := yaml.Unmarshal([]byte(*content), &nodeContent)
	if err != nil {
		return err
	}

	parameters := extractParameters(nodeContent.Content)
	for k, orbUpdate := range getOrbUpdates(&nodeContent, &parameters) {
		if _, contains := (*updates)[orbUpdate.Value]; !contains {
			(*updates)[orbUpdate.Value] = Update{
				Type:        "orb",
				CurrentName: k,
				FileUpdates: make(map[string]FileUpdate),
			}
		}
		(*updates)[orbUpdate.Value].FileUpdates[*path] = FileUpdate{
			SHA:        SHA,
			Node:       orbUpdate,
			Content:    content,
			Parameters: &parameters,
		}
	}

	for k, dockerUpdate := range getDockerUpdates(&nodeContent, &parameters) {
		if _, contains := (*updates)[dockerUpdate.Value]; !contains {
			(*updates)[dockerUpdate.Value] = Update{
				Type:        "docker",
				CurrentName: k,
				FileUpdates: make(map[string]FileUpdate),
			}
		}
		(*updates)[dockerUpdate.Value].FileUpdates[*path] = FileUpdate{
			SHA:        SHA,
			Node:       dockerUpdate,
			Content:    content,
			Parameters: &parameters,
		}
	}
	return nil
}

// ReplaceVersion replaces a specific line in the yaml
func ReplaceVersion(lineNumber int, oldVersion, newVersion, content string) string {
	lines := strings.Split(content, "\n")
	theLine := lines[lineNumber]
	lines[lineNumber] = strings.ReplaceAll(theLine, oldVersion, newVersion)
	output := strings.Join(lines, "\n")
	return output
}

func ExtractParameterName(param string) string {
	r := regexp.MustCompile(`<<\s*parameters\.(\w+)\s*>>`)
	match := r.FindStringSubmatch(param)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
