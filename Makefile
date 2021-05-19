NAME := proxy
MEMORY := memory
VERSION := v0.0.3-rc1
REVISION := $(shell git rev-parse --short HEAD)
SRCS    := $(shell find . -type d -name archive -prune -o -type f -name '*.go')
LDFLAGS := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -extldflags \"-static\""

bin/$(NAME): $(SRCS)
	go build -o bin/$(NAME)

bin/$(NAME)/static: $(SRCS)
	go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o bin/$(NAME)

.PHONY: deps
deps:
	go get -v

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf dsit/*
	rm -rf oidc-plugin/*

.PHONY: cross-build
cross-build: deps
	for os in darwin linux windows; do \
		for arch in amd64 386; do \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$$os-$$arch/$(NAME); \
		done; \
	done

DIST_DIRS := find * -type d -exec
.PHONY: dist
dist:
	cd dist && \
	$(DIST_DIRS) cp ../LICENSE {} \; && \
	$(DIST_DIRS) cp ../README.md {} \; && \
	$(DIST_DIRS) tar -zcf $(NAME)-$(VERSION)-{}.tar.gz {} \; && \
	$(DIST_DIRS) zip -r $(NAME)-$(VERSION)-{}.zip {} \; && \
	cd ..

plugin/$(MEMORY):
	go build -o oidc-plugin/$(MEMORY) examples/memory/main.go

plugin/$(MEMORY)/static:
	go build -a -tags netgo -installsuffix netgo -o oidc-plugin/$(MEMORY) examples/memory/main.go