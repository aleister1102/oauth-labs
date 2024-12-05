#!/bin/sh

chown 1000:1000 /app/config.yaml
exec /usr/local/bin/gosu 1000:1000 /app/server /app/config.yaml
