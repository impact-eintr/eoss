#!/usr/bin/env bash

curl -v 127.0.0.1:8080/objects/$1 --output $2
