#!/usr/bin/env cat
# 在终端里设置对应的环境变量

# es 服务的地址 host:port
export ES_SERVER=

# raftd 服务的地址 host:port
export RAFTD_SERVER=

# esq服务的地址 host:port
export ESQ_SERVER=

# rabbitmq服务的地址 host:port
export RABBITMQ_SERVER=

# 当前服务在集群中的位置 host + port
export LISTEN_ADDRESS=
export LISTEN_PORT=

# 文件存储地址前缀
export STORAGE_ROOT=

# 设定服务日志输出级别
export GIN_MODE=release
