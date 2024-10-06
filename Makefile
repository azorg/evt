# File: "Makefile"

PRJ="github.com/azorg/evt"

GIT_MESSAGE = "auto commit"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# go packages
PKGS = $(PRJ)

.PHONY: all help distclean commit tidy vendor fmt test

all: fmt test

help:
	@echo "make all       - format sources and run test"
	@echo "make help      - this help"
	@echo "make distclean - full clean (go.mod, go.sum)"
	@echo "make fmt       - format Go sources"
	@echo "make simplify  - simplify Go sources (go fmt -s)"
	@echo "make vet       - report likely mistakes (go vet)"
	@echo "make go.mod    - generate go.mod"
	@echo "make go.sum    - generate go.sum"
	@echo "make tidy      - automatic update go.sum by tidy"
	@echo "make commit    - auto commit by git"
	@echo "make test      - run test"

distclean:
	@rm -f go.mod
	@rm -f go.sum
	@#sudo rm -rf go/pkg
	@rm -rf vendor
	@go clean -modcache
	
fmt: go.mod go.sum
	@go fmt

simplify:
	@gofmt -l -w -s $(SRC)

vet:
	@#go vet
	@go vet $(PKGS)

go.mod:
	@go mod init $(PRJ)
	@touch go.mod

go.sum: go.mod Makefile #tidy
	@touch go.sum

tidy: go.mod
	@go mod tidy

commit: fmt
	git add .
	git commit -am $(GIT_MESSAGE)
	git push

test: go.mod go.sum
	@go test

# EOF: "Makefile"
