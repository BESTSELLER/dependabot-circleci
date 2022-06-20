module github.com/BESTSELLER/dependabot-circleci

go 1.16

require (
	cloud.google.com/go/bigquery v1.26.0
	github.com/BESTSELLER/go-json v0.0.2
	github.com/BESTSELLER/go-vault v0.1.2
	github.com/CircleCI-Public/circleci-cli v0.1.16535
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/docker/cli v20.10.11+incompatible // indirect
	github.com/docker/docker v20.10.11+incompatible // indirect
	github.com/gobuffalo/packr/v2 v2.8.3 // indirect
	github.com/golang-jwt/jwt/v4 v4.2.0 // indirect
	github.com/google/go-containerregistry v0.7.0
	github.com/google/go-github/v43 v43.0.0
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/hashicorp/go-version v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/palantir/go-baseapp v0.3.1
	github.com/palantir/go-githubapp v0.12.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.26.1
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a // indirect
	goji.io v2.0.2+incompatible
	golang.org/x/net v0.0.0-20211205041911-012df41ee64c // indirect
	google.golang.org/api v0.63.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.38.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/hashicorp/go-version => github.com/BESTSELLER/go-version v1.2.5
