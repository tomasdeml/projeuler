default: test

.PHONY: test
test:
	go test -v -count=1 ./...

.PHONY: test-race
test-race:
	go test -race -v -count=1 ./...
