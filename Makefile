SHA := $(shell git rev-parse --short=8 HEAD)
GITVERSION := $(shell git describe --long)
VERSION := $(shell echo $(GITVERSION) | cut -f1 -d-)
SERIAL := $(shell echo $(GITVERSION) | cut -f2 -d-).$(SHA)
BUILDDATE := $(shell date --rfc-3339=seconds)

all: test
	go build -ldflags "-X 'main.revision=$(VERSION)-$(SERIAL)' -X 'main.builddate=$(BUILDDATE)'" -buildmode=c-shared -o out_redis.so .

test:
	go test -cover -race -coverprofile=coverage.txt -covermode=atomic

dep:
	dep ensure

clean:
	rm -rf *.so *.h *~
