package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

type LogInfo struct {
	PodName string `json:"pod_name"`
	Log     []byte `json:"log"`
}

func RegisterLogRouter(ctx context.Context, v1group *gin.RouterGroup) {
	group := v1group.Group("/logs")

	group.GET("get", func(c *gin.Context) {
		// 解析请求中的 podName 参数
		podNames, ok := c.GetQueryArray("podName")
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "podName parameter is missing"})
			return
		}

		// 创建日志内容读取的 channel 和 wait group
		logContentCh := make(chan LogInfo, len(podNames))
		var wg sync.WaitGroup

		// 并发读取日志文件内容
		for _, podName := range podNames {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()

				// 获取Pod的日志文件路径
				logFilePath := "/path/to/logs/" + name + ".log"

				// 检查日志文件是否存在
				if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
					logContentCh <- LogInfo{
						PodName: name,
						Log:     []byte(fmt.Sprintf("log file %s does not exist\n", logFilePath)),
					}
					return
				}

				// 读取日志文件内容
				logContent, err := os.ReadFile(logFilePath)
				if err != nil {
					logContentCh <- LogInfo{
						PodName: name,
						Log:     []byte(fmt.Sprintf("failed to read log file %s: %s\n", logFilePath, err)),
					}
					return
				}

				// 将日志文件内容写入 channel
				logContentCh <- LogInfo{
					PodName: name,
					Log:     logContent,
				}
			}(podName)
		}

		// 等待所有的 goroutine 执行完毕
		wg.Wait()

		// 从 channel 中读取日志文件内容并合并
		logMap := make(map[string][]byte)
		for i := 0; i < len(podNames); i++ {
			logInfo := <-logContentCh
			logMap[logInfo.PodName] = logInfo.Log
		}

		// 将日志文件内容返回给请求者
		c.JSON(http.StatusOK, logMap)
	})
}
