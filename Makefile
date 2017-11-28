all: test
	go build -buildmode=c-shared -o out_redis.so .

test:
	go test

dep:
	dep ensure

clean:
	rm -rf *.so *.h *~
