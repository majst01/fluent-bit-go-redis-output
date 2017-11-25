
all:
	dep ensure
	go build -buildmode=c-shared -o out_redis.so .

fast:
	dep ensure
	go build out_redis.go

clean:
	rm -rf *.so *.h *~
