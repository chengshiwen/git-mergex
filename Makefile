# Makefile

VERSION     := 0.1.0
LDFLAGS     ?= "-s -w -X github.com/chengshiwen/git-mergex/cmd.Version=$(VERSION)"
GOBUILD_ENV = GO111MODULE=on CGO_ENABLED=0
GOX         = go run github.com/mitchellh/gox
TARGETS     := darwin/amd64 darwin/arm64
DIST_DIRS   := find * -maxdepth 0 -type d -exec

.PHONY: build cross-build release test lint down tidy clean

all: build

build:
	$(GOBUILD_ENV) go build -o bin/git-mergex -a -ldflags $(LDFLAGS)

cross-build: clean
	$(GOBUILD_ENV) $(GOX) -ldflags $(LDFLAGS) -parallel=2 -output="bin/git-mergex-{{.OS}}-{{.Arch}}/git-mergex" -osarch='$(TARGETS)' .

release: cross-build
	( \
		cd bin && \
		$(DIST_DIRS) cp ../LICENSE {} \; && \
		$(DIST_DIRS) cp ../README.md {} \; && \
		$(DIST_DIRS) tar -zcf {}.tar.gz {} \; && \
		$(DIST_DIRS) zip -r {}.zip {} \; && \
		$(DIST_DIRS) rm -rf {} \; && \
		sha256sum * > sha256sums.txt \
	)

test:
	go test -v ./...

lint:
	golangci-lint run --enable=golint --disable=errcheck --disable=typecheck && goimports -l -w . && go fmt ./... && go vet ./...

down:
	go list ./... && go mod verify

tidy:
	rm -f go.sum && go mod tidy -v

clean:
	rm -rf bin