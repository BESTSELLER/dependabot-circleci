FROM golang:1.17.1-alpine AS builder
WORKDIR $GOPATH/src/dependabot-circleci
COPY . .
ARG VERSION=1.0.0
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X gh.version=${VERSION}" -o /tmp/dependabot-circleci

FROM alpine
COPY --from=builder /tmp/dependabot-circleci /dependabot-circleci

CMD ["/dependabot-circleci"]
EXPOSE 3000