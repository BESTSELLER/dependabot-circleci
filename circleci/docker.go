package circleci

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
)

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
	
	// fix this shit
	listTags("harbor.bestsellerit.com/library/harpocrates:1.0.0")
	listTags("node:14.14-alpine")
	
	// query that damn registry for newer versions

	return "latest"
	// This one is a bit tricky actually! Watchtower seems to do this by utilising a docker client, but then we need
	// Docker in docker i guess ? Maybe there is a smart api endpoint, all registries should use the same to communicate with docker i guess ?
}

func listTags(circleciTag string) {
	dig, err := name.NewTag(circleciTag, name.WeakValidation)
	if err != nil {
		panic(err)
	}

	registryName := dig.Registry.RegistryStr()
	repoName := dig.Repository.RepositoryStr()

	newName, err := name.NewRepository(fmt.Sprintf("%s/%s", registryName, repoName), name.WeakValidation)
	if err != nil {
		panic(err)
	}
	tags, err := google.List(newName)
	if err != nil {
		panic(err)
	}

	for _, tag := range tags.Tags {
		fmt.Println(tag)
	}
}

