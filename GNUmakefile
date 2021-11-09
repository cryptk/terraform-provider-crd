VERSION=0.0.5
TEST?=$$(go list ./...)
TEST?="./crd"
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

default: build

build: fmtcheck
	go build -o build/bin/terraform-provider-crd

release: fmtcheck
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/bin/terraform-provider-crd_$(VERSION)-linux-amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/bin/terraform-provider-crd_$(VERSION)-darwin-amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o build/bin/terraform-provider-crd_$(VERSION)-darwin-arm64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o build/bin/terraform-provider-crd_$(VERSION)-windows-amd64

test: fmtcheck
	@go test -i $(TEST) || exit 1
	@echo " >> Running tests"
	@go test -v $(TEST)

gotest: fmtcheck
	@gotestsum --format testname $(TEST)

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@echo " >> Checking that code follows gofmt"
ifeq ($(GOFMT_FILES),)
	@echo "gofmt needs to be run on the following files:"
	@echo "$(GOFMT_FILES)"
	@echo "You can use the command 'make fmt' to reformat code."
	@exit 1
endif

clean:
	rm build/bin/*

.PHONY: build release test gotest fmt fmtcheck clean
