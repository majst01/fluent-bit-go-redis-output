#!/usr/bin/env sh

set -e

ls -l /fluent-bit/bin/

/fluent-bit/bin/fluent-bit -c /fluent-bit/etc/fluent-bit.conf \
                           -e /fluent-bit/bin/out_redis.so \
                           -i cpu