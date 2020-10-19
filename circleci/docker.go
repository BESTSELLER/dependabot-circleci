package circleci

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"gopkg.in/yaml.v3"
)

func extractImages(images []*yaml.Node) {
	updates := map[string]*yaml.Node{}
	for i := 0; i < len(images); i++ {
		image := images[i]
		if image.Value == "image" {
			image = images[i+1]

			imageVersion := findNewestDockerVersion(image.Value)

			if image.Value != imageVersion {
				oldVersion := image.Value
				image.Value = imageVersion
				updates[oldVersion] = image
			}
			image.Value = findNewestDockerVersion(image.Value)
		}
		extractImages(image.Content)
	}
}

func findNewestDockerVersion(currentVersion string) string {

	fmt.Println(currentVersion)

	current := strings.Split(currentVersion, ":")

	// check if image has no version tag
	if len(current) == 1 {
		return currentVersion
	}

	// check if tag is latest
	if strings.ToLower(current[1]) == "latest" {
		return currentVersion
	}

	// fix this shit
	tags, err := getTags(currentVersion)
	if err != nil {
		// panic(err)
		return currentVersion
	}

	versionParts := splitVersion(currentVersion)

	var newTagsList []string
	for _, tag := range tags {
		aa := splitVersion(tag)

		if aa["prefix"] == versionParts["prefix"] && aa["suffix"] == versionParts["suffix"] {
			newTagsList = append(newTagsList, tag)
		}
	}

	newest := newTagsList[len(newTagsList)-1]

	return newest
}

func getTags(circleciTag string) ([]string, error) {
	dig, err := name.NewTag(circleciTag, name.WeakValidation)
	if err != nil {
		return nil, err
	}

	registryName := dig.Registry.RegistryStr()
	repoName := dig.Repository.RepositoryStr()

	newName, err := name.NewRepository(fmt.Sprintf("%s/%s", registryName, repoName), name.WeakValidation)
	if err != nil {
		return nil, err
	}
	tags, err := google.List(newName)
	if err != nil {
		return nil, err
	}

	return tags.Tags, nil
}

func splitVersion(version string) map[string]string {
	// Regex stolen with love from dependabot-core
	// https://github.com/dependabot/dependabot-core/blob/v0.123.0/docker/lib/dependabot/docker/update_checker.rb#L15-L27
	versionRegex := `v?(?P<version>[0-9]+(?:(?:\.[a-z0-9]+)|(?:-(?:kb)?[0-9]+))*)`
	versionWithSFX := versionRegex + `(?P<suffix>-[a-z0-9.\-]+)?$`
	versionWithPFX := `(?P<prefix>[a-z0-9.\-]+-)?` + versionRegex + `$`
	versionWithPFXSFX := `(?P<prefix>[a-z\-]+-)?` + versionRegex + `(?P<suffix>-[a-z\-]+)?$`

	nameWithVersion := versionWithPFX + `|` + versionWithSFX + `|` + versionWithPFXSFX

	var myExp = regexp.MustCompile(nameWithVersion)
	match := myExp.FindStringSubmatch(version)
	result := make(map[string]string)

	matches := myExp.SubexpNames()
	for i, name := range matches {
		if i != 0 && name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	return result
}
