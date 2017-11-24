all:
	go build -buildmode=c-shared -o out_redis.so .

fast:
	go build out_redis.go

clean:
	rm -rf *.so *.h *~
