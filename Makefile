default: services test

.PHONY: test
test:
	CGO_ENABLED=1 go test -benchmem -bench=. -v ./... -race -coverprofile=coverage.out -covermode=atomic && go tool cover -func=coverage.out

.PHONY: lint
lint:
	golangci-lint run

.PHONY: cover
cover:
	go list -f '{{if len .TestGoFiles}}"go test -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"{{end}}' ./... | xargs -L 1 sh -c

.PHONY: publish_cover
publish_cover: cover
	go get -d golang.org/x/tools/cmd/cover
	go get github.com/modocache/gover
	go get github.com/mattn/goveralls
	gover
	@goveralls -coverprofile=gover.coverprofile -service=travis-ci -repotoken=$(COVERALLS_TOKEN)

.PHONY: clean
clean:
	@find . -name \.coverprofile -type f -delete
	@rm -f gover.coverprofile
