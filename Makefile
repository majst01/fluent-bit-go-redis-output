SHA := $(shell git rev-parse --short=8 HEAD)
GITVERSION := $(shell git describe --long --always)
BUILDDATE := $(shell date --rfc-3339=seconds)

all: test
	go build -ldflags "-X 'main.revision=$(GITVERSION)' -X 'main.builddate=$(BUILDDATE)'" -buildmode=c-shared -o out_redis.so .

test:
	go test -cover -race -coverprofile=coverage.txt -covermode=atomic

dep:
	dep ensure

clean:
	rm -rf *.so *.h *~
