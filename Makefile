all: test
	go build -buildmode=c-shared -o out_redis.so .

test:
	go test -cover -race -coverprofile=coverage.txt -covermode=atomic

dep:
	dep ensure

clean:
	rm -rf *.so *.h *~
