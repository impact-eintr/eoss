#!/usr/bin/env bash

make

./esqd -http_addr="172.18.0.3:9501" -tcp_addr="172.18.0.3:9502" -raftd_endpoint="172.18.0.4:8001" -node_id=1 -node_weight=1 -data_save_path=./data -enable_cluster=true -enable_raftd=true
