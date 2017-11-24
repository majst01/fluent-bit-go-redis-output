# fluent-bit redis output plugin


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

## Usage

```bash
docker run -it --rm -v $PWD:/tmp -e REDIS_KEY=eskey -e REDIS_HOST=redis -e REDIS_PORT=6379 fluent/fluent-bit 
/fluent-bit/bin/fluent-bit -e /tmp/out_redis.so -i cpu -o redis
```

## Redis server usage and availability

Given a list of 4 Redis databases, we pick on start a random one, if during operation this fails we panic and on restart the next hopefully working is selected.
