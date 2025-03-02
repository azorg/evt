# File: "Makefile"

PRJ="github.com/azorg/evt"

GIT_MESSAGE = "WiP: auto commit"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# go packages
PKGS = $(PRJ)

.PHONY: all help distclean commit tidy vendor fmt test

all: fmt doc test

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

clean:
	@rm -f doc.txt
	@rm -f doc.md

distclean: clean
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

doc: README.txt README.md

README.txt: *.go
	go doc -all > README.txt

README.md: *.go ~/go/bin/gomarkdoc
	~/go/bin/gomarkdoc -o README.md

~/go/bin/gomarkdoc:
	go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest

# EOF: "Makefile"
