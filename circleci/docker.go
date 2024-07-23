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

func extractImages(images []*yaml.Node, parameters *map[string]*yaml.Node) map[string]*yaml.Node {
	updates := map[string]*yaml.Node{}
	for i := 0; i < len(images); i++ {
		image := images[i]
		if image.Value == "image" {
			image = images[i+1]

			log.Debug().Msg(fmt.Sprintf("current image version: %s", image.Value))
			imageName, currentTag, newestTag := findNewestDockerVersion(image.Value, parameters)
			log.Debug().Msg(fmt.Sprintf("new image version: %s:%s", imageName, newestTag))

			if currentTag != newestTag {
				oldVersion := image.Value
				image.Value = newestTag
				updates[oldVersion] = image
			}
		}
		baah := extractImages(image.Content, parameters)
		for k, v := range baah {
			updates[k] = v
		}
	}
	return updates
}

func findNewestDockerVersion(currentVersion string, parameters *map[string]*yaml.Node) (imageName, currentTag, newestTag string) {
	current := strings.Split(currentVersion, ":")

	// check if image has no version tag
	if len(current) == 1 {
		return currentVersion, "", ""
	}

	// check if tag is latest
	if strings.ToLower(current[1]) == "latest" {
		return current[0], current[1], current[1]
	}
	imageName = current[0]
	currentTag = current[1]

	if newVersion, hit := cache[currentVersion]; hit {
		log.Debug().Msgf("Using cached version for image: %s", currentVersion)
		return imageName, currentTag, newVersion
	}

	if param := ExtractParameterName(currentVersion); len(param) > 0 {
		paramDefault, found := (*parameters)[param]
		if !found {
			log.Debug().Msgf("Parameter %s not found in parameters", param)
			return imageName, currentTag, currentTag
		}
		currentTag = paramDefault.Value
	}

	// fix this shit
	tags, err := getTags(imageName)
	if err != nil {
		log.Debug().Err(err)
		return imageName, currentTag, currentTag
	}

	versionParts := splitVersion(currentTag)
	if len(versionParts) == 0 {
		return imageName, currentTag, currentTag
	}
	var newTagsList []string
	for _, tag := range tags {
		aa := splitVersion(tag)

		if aa["version"] != "" && aa["prefix"] == versionParts["prefix"] && aa["suffix"] == versionParts["suffix"] {
			newTagsList = append(newTagsList, tag)
		}

	}
	currentv, _ := version.NewVersion(versionParts["version"])
	errorList := []string{}
	versions := []*version.Version{}
	for _, raw := range newTagsList {
		v, err := version.NewVersion(raw)
		if err != nil {
			errorList = append(errorList, fmt.Sprintf("%s", err))
			continue
		}

		if len(v.Prerelease()) > 0 {
			log.Debug().Msgf("version %s skipped, prerelease", v.Original())
			continue
		}
		if trimmed := TrimSemver(currentv.Original(), v.Original()); len(trimmed) != len(v.Original()) {
			log.Debug().Msgf("version %s skipped, must have the same ammount of segments as old version %s", v.Original(), currentv.Original())
			continue
		}

		versions = append(versions, v)
	}

	if len(errorList) > 0 {
		log.Debug().Err(fmt.Errorf("you have the following errors: %s", errorList))
	}

	sort.Sort(version.Collection(versions))

	newest := versions[len(versions)-1]

	if currentv.GreaterThan(newest) {
		cache[currentVersion] = currentTag
		return imageName, currentTag, currentTag
	}
	newVersion := newest.Original()
	cache[currentVersion] = newVersion
	return imageName, currentTag, newVersion
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

	// https://github.com/dependabot/dependabot-core/blob/v0.211.0/docker/lib/dependabot/docker/update_checker.rb#L191
	// Some people suffix their versions with commit SHAs.
	commitSHA := regexp.MustCompile(`(^|\-g?)[0-9a-f]{7,}$`)
	var filteredTags []string
	for _, tag := range tags {
		if commitSHA.MatchString(tag) {
			continue
		}
		filteredTags = append(filteredTags, tag)
	}

	return filteredTags, nil
}

func splitVersion(version string) map[string]string {
	result := make(map[string]string)
	// Regex stolen with love from dependabot-core
	// https://github.com/dependabot/dependabot-core/blob/v0.211.0/docker/lib/dependabot/docker/update_checker.rb#L45-L56
	versionRegex := `v?(?P<version>[0-9]+(?:(?:\.[a-z0-9]+)|(?:-(?:kb)?[0-9]+))*)`
	versionWithSFX := versionRegex + `(?P<suffix>-[a-z0-9.\-]+)?$`
	versionWithPFX := `^(?P<prefix>[a-z0-9.\-]+-)?` + versionRegex + `$`
	versionWithPFXSFX := `^(?P<prefix>[a-z\-]+-)?` + versionRegex + `(?P<suffix>-[a-z\-]+)?$`

	nameWithVersion := versionWithPFX + `|` + versionWithSFX + `|` + versionWithPFXSFX

	var myExp = regexp.MustCompile(nameWithVersion)
	match := myExp.FindStringSubmatch(version)

	if match == nil {
		return result
	}

	matches := myExp.SubexpNames()
	for i, imgName := range matches {
		if i != 0 && imgName != "" && match[i] != "" {
			result[imgName] = match[i]
		}
	}

	return result
}
