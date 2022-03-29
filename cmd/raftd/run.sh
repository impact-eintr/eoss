#!/usr/bin/env bash

make

./raftd -id node01 -haddr 172.18.0.3:8001 -raddr 172.18.0.3:8101 /.raftd01
