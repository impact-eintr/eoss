#!/usr/bin/env bash

a=`cat $1 | openssl dgst -sha256 -binary| base64 |sed ':t;N;s/\n//;b t'`
curl -v 127.0.0.1:8080/objects/$2 -XPUT -T $1  -H "Digest: SHA-256=$a"
