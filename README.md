# fluent-bit redis output plugin

## Usage

```bash
docker build -tr fluent-bit-go-redis-output
docker run -it --rm -e REDIS_HOST=172.0.0.3 -e REDIS_PORT=6379 -e REDIS_KEY=eskey fluent-bit-go-redis-output
```

## Useful links

### Redis Libraries

- [go-redis](https://github.com/go-redis/redis)
- [redigo](https://github.com/garyburd/redigo)


### TLS Socket

- [stunnel](https://github.com/liudanking/stunnel)
- [tlsproxy](https://github.com/getlantern/tlsproxy/blob/master/tlsproxy.go)

### Redis format

- [logrus-redis-hook](https://github.com/rogierlommers/logrus-redis-hook/blob/master/logrus_redis.go)

### Logstash Redis Output

- [logstash-redis-docu](https://github.com/logstash-plugins/logstash-output-redis/blob/master/docs/index.asciidoc)

## Redis server usage and availability

1. crash on connection errors

Given a list of 4 Redis databases, we pick on start a random one, if during operation this fails we panic and on restart the next hopefully working is selected.

1. rely on FLB_RETRY

With a list of redis databases we can create a list of pools, one pool per database and instead of doing a pool.Get(), call list.Get() with selects the next random redis database. If a failure occurs return FLB_RETRY and the library will retry.

## TODO

- TLS redis.DialUseTLS
- string concat instead of json.Marshal
- HA

