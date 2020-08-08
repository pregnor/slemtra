.DEFAULT_GOAL=all

.PHONY: all
all: check-dependencies analyze-diff build test

.PHONY: analyze
analyze:
	golangci-lint run ./...
	go list -u -m -mod=readonly -json all\
		| go-mod-outdated -ci -direct -update

.PHONY: analyze-diff
analyze-new:
	golangci-lint run --new ./...
	go list -u -m -mod=readonly -json all\
		| go-mod-outdated -ci -direct -update

.PHONY: build
build:
	go install ./...

.PHONY:
check-dependencies:
	which \
		go \
		go-mod-outdated \
		golangci-lint

.PHONY: count-analysis
count-analysis:
	golangci-lint run ./...\
		| grep -E "^[0-9a-zA-Z/\.\-\_].*\([0-9a-zA-Z]+\)$$"\
		| sed -E 's/^[0-9a-zA-Z\/\.\-\_].*\(([0-9a-zA-Z]+)\)$$/\1/g'\
		| sort\
		| uniq -c\
		| sort -bgr

.PHONY: test
test:
	go test -count 1 -cover -race -run ".*" -timeout 10s ./...
