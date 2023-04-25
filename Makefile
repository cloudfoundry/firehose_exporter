all: test

test:
	@go test -v ./...

check:
	@golangci-lint run --config .golangci.yml

coverage:
	@go test -cover -coverprofile cover.out -v ./...
	@go tool cover -func=cover.out
	@rm -f cover.out
