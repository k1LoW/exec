export GO111MODULE=on

default: test

ci: test sec

test:
	go test ./... -coverprofile=coverage.txt -covermode=count

sec:
	gosec ./...
