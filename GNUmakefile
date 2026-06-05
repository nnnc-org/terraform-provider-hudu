NAMESPACE  := nnnc-org
NAME       := hudu
BINARY     := terraform-provider-$(NAME)
VERSION    := 0.1.0
OS_ARCH    := $(shell go env GOOS)_$(shell go env GOARCH)

INSTALL_DIR := $(HOME)/.terraform.d/plugins/registry.terraform.io/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)

.PHONY: default build install test testacc fmt vet tidy docs clean

default: build

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(INSTALL_DIR)
	mv $(BINARY) $(INSTALL_DIR)/$(BINARY)

test:
	go test ./... -count=1

testacc:
	TF_ACC=1 go test ./internal/provider/... -v -count=1 -timeout 120m

fmt:
	gofmt -s -w .

vet:
	go vet ./...

tidy:
	go mod tidy

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate \
		--provider-name $(NAME) \
		--rendered-provider-name Hudu

clean:
	rm -f $(BINARY)
