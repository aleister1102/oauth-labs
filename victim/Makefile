.PHONY: all
all:
	go build -ldflags="-s -w" -o ./bin/server ./cmd/server/main.go

.PHONY: lint
lint:
	goimports-reviser -format ./...
	fieldalignment -fix ./...
	golangci-lint run

.PHONY: test
test:
	go test -v ./...

