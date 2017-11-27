FROM golang:1.9 AS builder

COPY Makefile redis.go out_redis.go Gopkg.toml /go/src/github.com/majst01/fluent-bit-go-redis-output/

RUN cd /go/src/github.com/majst01/fluent-bit-go-redis-output/ \
 && go get -u github.com/golang/dep/cmd/dep \
 && make dep all

FROM fluent/fluent-bit

COPY --from=builder /go/src/github.com/majst01/fluent-bit-go-redis-output/out_redis.so /fluent-bit/bin/
COPY start.sh /start.sh

CMD ["/start.sh"]
