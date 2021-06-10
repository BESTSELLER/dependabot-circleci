module github.com/BESTSELLER/dependabot-circleci

go 1.16

require (
	github.com/BESTSELLER/go-vault v0.1.2
	github.com/CircleCI-Public/circleci-cli v0.1.15224
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/google/go-containerregistry v0.5.1
	github.com/google/go-github/v35 v35.3.0
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/hashicorp/go-version v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/palantir/go-baseapp v0.2.3
	github.com/palantir/go-githubapp v0.7.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.22.0
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a // indirect
	goji.io v2.0.2+incompatible
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/hashicorp/go-version => github.com/BESTSELLER/go-version v1.2.5
