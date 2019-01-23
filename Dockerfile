FROM golang:1.11 AS builder

WORKDIR /go/src/github.com/majst01/fluent-bit-go-redis-output/

COPY .git Makefile Gopkg.* *.go /go/src/github.com/majst01/fluent-bit-go-redis-output/
RUN go get -u github.com/golang/dep/cmd/dep \
 && make dep all

FROM fluent/fluent-bit:1.0.3

COPY --from=builder /go/src/github.com/majst01/fluent-bit-go-redis-output/out_redis.so /fluent-bit/bin/
COPY *.conf /fluent-bit/etc/
COPY start.sh /start.sh

CMD ["/start.sh"]
