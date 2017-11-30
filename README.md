# fluent-bit redis output plugin

[![Build Status](https://travis-ci.org/majst01/fluent-bit-go-redis-output.svg?branch=master)](hhttps://travis-ci.org/majst01/fluent-bit-go-redis-output)

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

# Usage

```bash
docker run -it --rm -v /path/to/fluent-bit.conf:/fluent-bit/etc/fluent-bit.conf majst01/fluent-bit-go-redis-output
```

## Building

```bash
docker build --no-cache --tag fluent-bit-go-redis-output .
docker run -it --rm -v /path/to/fluent-bit.conf:/fluent-bit/etc/fluent-bit.conf fluent-bit-go-redis-output
```

### Configuration Options

| Key           | Description                                    | Default        |
| --------------|------------------------------------------------|----------------|
| Hosts         | Host(s) of redis servers, whitespace separated ip/host:port | 127.0.0.1:6379 |
| Password      | Optional redis password for all redis instances | "" |
| DB            | redis database (integer)  | 0 |
| UseTLS        | connect to redis with tls | False |
| TlsSkipVerify | if tls is configured skip tls certificate validation for self signed certificates | True |
| Key           | the key where to store the entries in redis | "logstash" |


Example:

add this section to fluent-bit.conf

```properties
[Output]
    Name redis
    Match *
    UseTLS true
    TLSSkipVerify true
    # if port is ommited, 6379 is used
    Hosts 172.17.0.1 172.17.0.1:6380 172.17.0.1:6381 172.17.0.1:6382 172.17.0.1:6383
    Password homer
    DB 0
    Key elastic-logstash
```

## Useful links

### Redis format

- [logrus-redis-hook](https://github.com/rogierlommers/logrus-redis-hook/blob/master/logrus_redis.go)

### Logstash Redis Output

- [logstash-redis-docu](https://github.com/logstash-plugins/logstash-output-redis/blob/master/docs/index.asciidoc)

## TODO

### Strategies for redis connection error handling

1. crash on connection errors

Given a list of 4 Redis databases, we pick on start a random one, if during operation this fails we panic and on restart the next hopefully working is selected.

1. rely on FLB_RETRY

With a list of redis databases we can create a list of pools, one pool per database and instead of doing a pool.Get(), call list.Get() with selects the next random redis database. If a failure occurs return FLB_RETRY and the library will retry.
