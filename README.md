# fluent-bit redis output plugin

This plugin is used to have redis output from fluent-bit. You can use fluent-bit redis instead of logstash in a configuration
where you have a redis and optional stunnel in front of your elasticsearch infrastructure. 

The configuration typically looks like:

```graphviz
fluent-bit --> stunnel --> redis <-- logstash --> elasticsearch
```

If you have multiple elastic search servers, each covered with a redis cache in front it might look like this:

```graphviz
           /-> stunnel --> redis <-- logstash --> elasticsearch 
           |
fluent-bit --> stunnel --> redis <-- logstash --> elasticsearch
           |
           \-> stunnel --> redis <-- logstash --> elasticsearch
```

## Usage

```bash
docker build --tag --no-cache fluent-bit-go-redis-output .
docker run -it --rm \
        --env REDIS_HOSTS="172.17.0.1:6380 172.17.0.1:6381 172.17.0.1:6382 172.17.0.1:6383" \
        --env REDIS_KEY=logstash \
        --env REDIS_USETLS=true \
        --env REDIS_TLSSKIP_VERIFY=true \
    fluent-bit-go-redis-output
```

## Useful links

### Redis format

- [logrus-redis-hook](https://github.com/rogierlommers/logrus-redis-hook/blob/master/logrus_redis.go)

### Logstash Redis Output

- [logstash-redis-docu](https://github.com/logstash-plugins/logstash-output-redis/blob/master/docs/index.asciidoc)

## TODO strategies for redis connection error handling

1. crash on connection errors

Given a list of 4 Redis databases, we pick on start a random one, if during operation this fails we panic and on restart the next hopefully working is selected.

1. rely on FLB_RETRY

With a list of redis databases we can create a list of pools, one pool per database and instead of doing a pool.Get(), call list.Get() with selects the next random redis database. If a failure occurs return FLB_RETRY and the library will retry.
