BINARY_NAME=jfind

.PHONY: build
build:
	go build -o $(BINARY_NAME)

.PHONY: build-all
build-all:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe

.PHONY: clean
clean:
	go clean
	rm -f $(BINARY_NAME)*

.PHONY: test
test:
	go test -v ./...
