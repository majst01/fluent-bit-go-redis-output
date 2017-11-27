all:
	go build -buildmode=c-shared -o out_redis.so .

dep:
	dep ensure

clean:
	rm -rf *.so *.h *~
