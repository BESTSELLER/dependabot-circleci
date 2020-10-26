package circleci

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/hashicorp/go-version"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func extractImages(images []*yaml.Node) {
	updates := map[string]*yaml.Node{}
	for i := 0; i < len(images); i++ {
		image := images[i]
		if image.Value == "image" {
			image = images[i+1]

			imageVersion := findNewestDockerVersion(image.Value)
			log.Debug().Msg(fmt.Sprintf("new: %s", imageVersion))

			if image.Value != imageVersion {
				oldVersion := image.Value
				image.Value = imageVersion
				updates[oldVersion] = image
			}
		}
		extractImages(image.Content)
	}
}

func findNewestDockerVersion(currentVersion string) string {
	log.Debug().Msg(fmt.Sprintf("old: %s", currentVersion))

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
		log.Debug().Err(err)
		return currentVersion
	}

	versionParts := splitVersion(current[1])
	if len(versionParts) == 0 {
		return currentVersion
	}

	var newTagsList []string
	for _, tag := range tags {
		aa := splitVersion(tag)

		if aa["version"] != "" && aa["prefix"] == versionParts["prefix"] && aa["suffix"] == versionParts["suffix"] {
			newTagsList = append(newTagsList, tag)
		}

	}

	errorList := []string{}
	versions := []*version.Version{}
	for _, raw := range newTagsList {
		v, err := version.NewVersion(raw)
		if err != nil {
			errorList = append(errorList, fmt.Sprintf("%s", err))
			continue
		}

		versions = append(versions, v)
	}

	if len(errorList) > 0 {
		log.Debug().Err(fmt.Errorf("You have the following errors: %s", strings.Join(errorList, "\n")))
	}

	sort.Sort(version.Collection(versions))

	newest := versions[len(versions)-1]

	currentv, _ := version.NewVersion(versionParts["version"])
	if currentv.GreaterThan(newest) {
		return currentVersion
	}

	return fmt.Sprintf("%s:%s", current[0], newest.Original())
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
	tags, err := remote.List(newName)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func splitVersion(version string) map[string]string {
	result := make(map[string]string)
	// Regex stolen with love from dependabot-core
	// https://github.com/dependabot/dependabot-core/blob/v0.123.0/docker/lib/dependabot/docker/update_checker.rb#L15-L27
	versionRegex := `v?(?P<version>([0-9]+)\.([0-9]+)\.([0-9]+)?(?:\+[0-9A-Za-z-]+)?)`
	versionWithSFX := versionRegex + `(?P<suffix>-[a-z0-9.\-]+)?$`
	versionWithPFX := `(?P<prefix>[a-z0-9.\-]+-)?` + versionRegex + `$`
	versionWithPFXSFX := `(?P<prefix>[a-z\-]+-)?` + versionRegex + `(?P<suffix>-[a-z\-]+)?$`

	nameWithVersion := versionWithPFX + `|` + versionWithSFX + `|` + versionWithPFXSFX

	var myExp = regexp.MustCompile(nameWithVersion)
	match := myExp.FindStringSubmatch(version)

	if match == nil {
		return result
	}

	matches := myExp.SubexpNames()
	for i, name := range matches {
		if i != 0 && name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	return result
}
