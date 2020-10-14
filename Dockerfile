FROM golang:1.15 as build
WORKDIR $GOPATH/src/github.com/BESTELLER/dependabot-circleci
COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 go mod vendor
RUN GO111MODULE=on CGO_ENABLED=0 go install -mod=vendor

FROM alpine
WORKDIR /
COPY --from=build /go/bin/dependabot-circleci /

CMD /dependabot-circleci