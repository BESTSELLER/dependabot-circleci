package circleci

import (
	"fmt"
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

// func extractImages(images []*yaml.Node) map[string]*yaml.Node {
// 	updates := map[string]*yaml.Node{}
// 	for i := 0; i < len(images); i = i + 2 {
// 		image := images[i]
// 		if image.Value == "image" {
// 			image = images[i+1]

// 			imageVersion := findNewestDockerVersion(image.Value)

// 			if image.Value != imageVersion {
// 				oldVersion := image.Value
// 				image.Value = imageVersion
// 				updates[oldVersion] = image
// 			}
// 		}
// 		extractImages(image.Content)
// 	}
// 	return updates
// }

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
		panic(err)
	}
	newest := tags[len(tags)-1]

	fmt.Println(newest)

	return "latest"
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
