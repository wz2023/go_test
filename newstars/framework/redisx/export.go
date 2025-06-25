package redisx

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

func Pipeline() redis.Pipeliner {
	return getClient().Pipeline()
}

func Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return getClient().Pipelined(ctx, fn)
}

func TxPipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return getClient().TxPipelined(ctx, fn)
}

func TxPipeline() redis.Pipeliner {
	return getClient().TxPipeline()
}

// 字符串方法
func Get(ctx context.Context, key string) *redis.StringCmd {
	return getClient().Get(ctx, key)
}

func GetEx(ctx context.Context, key string, expire time.Duration) *redis.StringCmd {
	return getClient().GetEx(ctx, key, expire)
}

func GetDel(ctx context.Context, key string) *redis.StringCmd {
	return getClient().GetDel(ctx, key)
}

func Set(ctx context.Context, key string, value interface{}, expire time.Duration) *redis.StatusCmd {
	return getClient().Set(ctx, key, value, expire)
}

func SetNX(ctx context.Context, key string, value interface{}, expire time.Duration) *redis.BoolCmd {
	return getClient().SetNX(ctx, key, value, expire)
}

func SetXX(ctx context.Context, key string, value interface{}, expire time.Duration) *redis.BoolCmd {
	return getClient().SetXX(ctx, key, value, expire)
}

func MGet(ctx context.Context, keys ...string) *redis.SliceCmd {
	return getClient().MGet(ctx, keys...)
}

func MSet(ctx context.Context, values ...interface{}) *redis.StatusCmd {
	return getClient().MSet(ctx, values...)
}

func Incr(ctx context.Context, key string) *redis.IntCmd {
	return getClient().Incr(ctx, key)
}

func IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	return getClient().IncrBy(ctx, key, value)
}

func Decr(ctx context.Context, key string) *redis.IntCmd {
	return getClient().Decr(ctx, key)
}

func DecrBy(ctx context.Context, key string, decrement int64) *redis.IntCmd {
	return getClient().DecrBy(ctx, key, decrement)
}

// 哈希方法
func HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return getClient().HGet(ctx, key, field)
}

func HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return getClient().HSet(ctx, key, values...)
}

func HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd {
	return getClient().HMGet(ctx, key, fields...)
}

func HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd {
	return getClient().HMSet(ctx, key, values...)
}

func HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	return getClient().HGetAll(ctx, key)
}

func HIncrBy(ctx context.Context, key, field string, incr int64) *redis.IntCmd {
	return getClient().HIncrBy(ctx, key, field, incr)
}

func HKeys(ctx context.Context, key string) *redis.StringSliceCmd {
	return getClient().HKeys(ctx, key)
}

// 列表方法
func LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return getClient().LPush(ctx, key, values...)
}

func RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return getClient().RPush(ctx, key, values...)
}

func LPop(ctx context.Context, key string) *redis.StringCmd {
	return getClient().LPop(ctx, key)
}

func RPop(ctx context.Context, key string) *redis.StringCmd {
	return getClient().RPop(ctx, key)
}

func LLen(ctx context.Context, key string) *redis.IntCmd {
	return getClient().LLen(ctx, key)
}

func LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return getClient().LRange(ctx, key, start, stop)
}

// 集合方法
func SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return getClient().SAdd(ctx, key, members...)
}

func SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return getClient().SMembers(ctx, key)
}

func SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	return getClient().SIsMember(ctx, key, member)
}

// 有序集合方法
func ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	return getClient().ZAdd(ctx, key, members...)
}

func ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return getClient().ZRange(ctx, key, start, stop)
}

func ZRank(ctx context.Context, key, member string) *redis.IntCmd {
	return getClient().ZRank(ctx, key, member)
}

// 键操作
func Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	return getClient().Exists(ctx, keys...)
}

func Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return getClient().Del(ctx, keys...)
}

func Expire(ctx context.Context, key string, expire time.Duration) *redis.BoolCmd {
	return getClient().Expire(ctx, key, expire)
}

func TTL(ctx context.Context, key string) *redis.DurationCmd {
	return getClient().TTL(ctx, key)
}

func Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	return getClient().Scan(ctx, cursor, match, count)
}

// 流方法
func XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd {
	return getClient().XAdd(ctx, a)
}

func XRead(ctx context.Context, a *redis.XReadArgs) *redis.XStreamSliceCmd {
	return getClient().XRead(ctx, a)
}

// HyperLogLog 方法
func PFAdd(ctx context.Context, key string, els ...interface{}) *redis.IntCmd {
	return getClient().PFAdd(ctx, key, els...)
}

func PFCount(ctx context.Context, keys ...string) *redis.IntCmd {
	return getClient().PFCount(ctx, keys...)
}

// 脚本方法
func Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	return getClient().Eval(ctx, script, keys, args...)
}

func EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	return getClient().EvalSha(ctx, sha1, keys, args...)
}

// 发布/订阅
func Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	return getClient().Publish(ctx, channel, message)
}

// 集群方法
func ClusterSlots(ctx context.Context) *redis.ClusterSlotsCmd {
	return getClient().ClusterSlots(ctx)
}

func ClusterNodes(ctx context.Context) *redis.StringCmd {
	return getClient().ClusterNodes(ctx)
}
