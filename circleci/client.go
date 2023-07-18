package circleci

import (
	"net/http"

	"github.com/CircleCI-Public/circleci-cli/api/graphql"
)

var client *graphql.Client

func NewClient(token string) (*graphql.Client, error) {
	if token == "" {
		client = graphql.NewClient(http.DefaultClient, "https://circleci.com/", "graphql-unstable", "", false)
	} else {
		client = graphql.NewClient(http.DefaultClient, "https://circleci.com/", "graphql-unstable", token, false)
	}
	return client, nil
}
