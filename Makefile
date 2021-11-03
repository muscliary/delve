.DEFAULT_GOAL=test

GO_SRC := $(shell find . -type f -not -path './_fixtures/*' -not -path './vendor/*' -not -path './_scripts/*' -not -path './localtests/*' -name '*.go')

check-cert:
	@go run _scripts/make.go check-cert

build: $(GO_SRC)
	@go run _scripts/make.go build

install: $(GO_SRC)
	@go run _scripts/make.go install

uninstall:
	@go run _scripts/make.go uninstall

test: vet
	@go run _scripts/make.go test

vet:
	@go vet $$(go list ./... | grep -v native)

test-proc-run:
	@go run _scripts/make.go test -s proc -r $(RUN)

test-integration-run:
	@go run _scripts/make.go test -s service/test -r $(RUN)

vendor:
	@go run _scripts/make.go vendor

.PHONY: vendor test-integration-run test-proc-run test check-cert install build vet uninstall
