all:
	go build ./...
	go vet ./...

.PHONY: all
