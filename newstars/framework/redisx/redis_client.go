package redisx

import (
	"context"
	"github.com/redis/go-redis/v9"
	"newstars/framework/config"
	"newstars/framework/glog"
	"sync"
)

var (
	client redis.Cmdable
	once   sync.Once
)

func getClient() redis.Cmdable {
	if client == nil {
		panic("redis client not initialized: call redisx.Init() first")
	}
	return client
}

func InitRedis(redisConfig *config.Redis) {
	once.Do(func() {
		var (
			isCluster = redisConfig.IsCluster
			addr      = redisConfig.Addr
			password  = redisConfig.Password
			db        = redisConfig.DB
		)

		if isCluster {
			client = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs: []string{
					addr,
					// 添加其他节点的终端节点地址和端口号
				},
				Password: password, // 如果有密码的话，请替换成你的 Redis 密码
			})
			pong, err := client.Ping(context.Background()).Result()
			if err != nil {
				glog.SError("redis connect ping failed, err:", err)
			} else {
				glog.SInfo("redis connect ping response:", pong)
			}
		} else {
			client = redis.NewClient(&redis.Options{
				Addr:     addr,
				Password: password, // no password set
				DB:       db,       // use default DB
			})
			pong, err := client.Ping(context.Background()).Result()
			if err != nil {
				glog.SError("redis connect ping failed, err:", err)
			} else {
				glog.SInfo("redis connect ping response:", pong)
			}
		}
		glog.SInfo("redis client connect success")
	})
}
