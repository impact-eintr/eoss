package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/api/heartbeat"
	"github.com/impact-eintr/eoss/api/locate"
	"github.com/impact-eintr/eoss/api/metadata"
	"github.com/impact-eintr/eoss/api/objects"
	"github.com/impact-eintr/eoss/cluster"
	"github.com/impact-eintr/eoss/mq/esqv1"
)

func main() {

	// 经过测试 集群的上行带宽有限(原因未知) 对于 PUT 这样的上行负载 需要进行负载均衡
	// 配置apiserver集群 这是一个高可用负载均衡集群
	// 采用gossip 协议通信 使用 一致性散列 实现负载均衡
	clus := flag.String("cluster", "", "cluster address")
	flag.Parse()

	node, err := cluster.New(os.Getenv("LISTEN_ADDRESS"), *clus)
	if err != nil {
		log.Fatalln(err)
	}

	go heartbeat.ListenHeartbeat() // 监听心跳
	go esqv1.ListenFileResponse()  // 监听文件信息

	eng := gin.Default()

	objGroup := eng.Group("/objects")
	objGroup.Use(func(ctx *gin.Context) {
		// TODO RBAC
		log.SetPrefix(fmt.Sprintf("%s@[API_SERVER] > ", node.Addr()))
		//log.Printf("[API_SERVER_CLUSTERS] 集群 %v\n", node.Members())
	})
	{
		objGroup.POST("/:name")
		objGroup.PUT("/:name", func(ctx *gin.Context) {
			// PUT请求 准备占用上行带宽
			// 需要在 Header 中提供一个时间戳
			// TODO 这个时间戳之后可以加密 提高安全性
			// TODO 在成功处理完这个 PUT 请求后将这个 hash_key 删除
			timestamp := ctx.Request.Header.Get("TimeStamp")
			if timestamp == "" {
				timestamp = time.Now().UTC().String()
				ctx.Writer.Header().Set("TimeStamp", timestamp)
			}

			addr, ok := node.ShouldProcess(ctx.Param("name") + timestamp)
			redirect := FormRedirect(ctx.Request, addr+":"+os.Getenv("LISTEN_PORT"))
			if !ok {
				ctx.Abort()
				ctx.Redirect(http.StatusTemporaryRedirect, redirect)
			}
		}, objects.Put)
		objGroup.GET("/:name", objects.Get)
		objGroup.DELETE("/:name", objects.Delete)
	}

	eng.Group("/locate", locate.Get)

	metaGroup := eng.Group("meta")
	{
		searchGroup := metaGroup.Group("search")
		{
			searchGroup.GET("/objects", metadata.GetLastestVersions)
			searchGroup.GET("/object/:name", metadata.GetAllVersions)
		}
		thumbnailGroup := metaGroup.Group("thumbnail")
		{
			thumbnailGroup.PUT("/:name", metadata.PutThumbnail)
			thumbnailGroup.GET("/:name", metadata.GetThumbnail)
			thumbnailGroup.DELETE("/:name", metadata.DelThumbnail)
		}

	}

	eng.Run(":" + os.Getenv("LISTEN_PORT"))
}

// FormRedirect returns the value for the "Location" header for a 301 response.
func FormRedirect(r *http.Request, host string) string {
	protocol := "http"
	rq := r.URL.RawQuery
	if rq != "" {
		rq = fmt.Sprintf("?%s", rq)
	}
	return fmt.Sprintf("%s://%s%s%s", protocol, host, r.URL.Path, rq)
}
