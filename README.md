# eoss
一个分布式OSS（对象存储）

![img](img/EOSS架构设计.png)

# TODO

1. ~~数据的部分元数据需要另外的存储机制~~
    - 基于raft和bolt的分布式KV数据库 [raftd](https://github.com/impact-eintr/raftd) 绝赞开发中 : )
2. 一个客户端 
    - 具有文件压缩/解压的功能
    - 具有制作预览的功能
    -  缩略图
    1. **制作** 客户端制作 可能需要多个版本？
    2. **上传** 如何上传呢？暂时使用专用的api 
    3. **存储** 由es存储缩略图的hash，由raftd保存缩略图的文件({hash: file})
    4. **获取** 缩略图属于元数据，在获取元数据时分为两步：1、获取元数据本体(一个Metadata struct) 2、解析元数据中的缩略图地址，异步请求缩略图(apiServer/raftd)，在请求成功前先渲染某张固定的图片，之后替换(需要考虑请求失败怎么办)
    - 解析目录的功能(可选)
3. 断点续传功能
4. 支持目录(有难度，有必要吗)
5. ~~一个分布式的消息队列(借鉴nsq/nats/gmq)~~
    - 基于etcd(后续用raftd尝试代替)的分布式消息队列 [esq](https://github.com/impact-eintr/esq) 绝赞开发中 : )

