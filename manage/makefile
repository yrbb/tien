COMMIT_HASH=$(shell git rev-parse --short HEAD || echo "GitNotFound")
COMPILE=$(shell date '+%Y-%m-%d %H:%M:%S') by $(shell go version)
LDFLAGS="-X \"github.com/yrbb/rain.Version=${COMMIT_HASH}\" -X \"github.com/yrbb/rain.Compile=$(COMPILE)\""

.PHONY: all run build build-linux clean 
all: build

run: build
	@./bin/example test -c=config/config.toml

build: clean
	@go build -ldflags ${LDFLAGS} -o ./bin/example ./main.go

build-linux: clean
	@GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS} -o ./bin/example ./main.go

clean:
	@rm -rf bin
