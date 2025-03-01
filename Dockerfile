FROM golang:1.24.0-alpine AS builder
WORKDIR $GOPATH/src/dependabot-circleci
COPY . .
ARG VERSION=1.0.0
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X github.com/BESTSELLER/dependabot-circleci/config.Version=${VERSION}" -o /tmp/dependabot-circleci

FROM alpine:3.21.3
COPY --from=builder /tmp/dependabot-circleci /dependabot-circleci

ENTRYPOINT ["/dependabot-circleci"]
EXPOSE 3000
