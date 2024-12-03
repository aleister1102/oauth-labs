.PHONY: lint
lint:
	goimports-reviser -format ./...
	golangci-lint run

.PHONY: test
test:
	go test -v ./...
