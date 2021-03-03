module github.com/BESTSELLER/dependabot-circleci

go 1.15

require (
	github.com/BESTSELLER/go-vault v0.1.2
	github.com/CircleCI-Public/circleci-cli v0.1.15085
	github.com/go-co-op/gocron v0.6.0
	github.com/google/go-containerregistry v0.4.1
	github.com/google/go-github/v33 v33.0.0
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/hashicorp/go-version v1.2.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/palantir/go-baseapp v0.2.3
	github.com/palantir/go-githubapp v0.6.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.20.0
	goji.io v2.0.2+incompatible
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/hashicorp/go-version => github.com/BESTSELLER/go-version v1.2.5
