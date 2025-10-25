package cache

import (
	"context"
	"time"

	"github.com/ishaqcherry9/depend/pkg/encoding"
	"github.com/redis/go-redis/v9"
)

type RedisOriginCmds interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Decr(ctx context.Context, key string) *redis.IntCmd
	HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd
	HSetNX(ctx context.Context, key, field string, value interface{}) *redis.BoolCmd
	HMGet(ctx context.Context, key string, field ...string) *redis.SliceCmd
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	HGet(ctx context.Context, key, field string) *redis.StringCmd
	HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd
	LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	LPop(ctx context.Context, key string) *redis.StringCmd
	RPop(ctx context.Context, key string) *redis.StringCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
	ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd
	ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd
	ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd
	ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	SetArgs(ctx context.Context, key string, value interface{}, a redis.SetArgs) *redis.StatusCmd
	Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error)
	Pipeline() redis.Pipeliner
	TxPipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error)
	TxPipeline() redis.Pipeliner
	Process(ctx context.Context, cmd redis.Cmder) error
	Options() *redis.Options
	PoolStats() *redis.PoolStats
	PSubscribe(ctx context.Context, channels ...string) *redis.PubSub
	SSubscribe(ctx context.Context, channels ...string) *redis.PubSub
	NewSearchBuilder(ctx context.Context, index, query string) *redis.SearchBuilder
	NewAggregateBuilder(ctx context.Context, index, query string) *redis.AggregateBuilder
	NewCreateIndexBuilder(ctx context.Context, index string) *redis.CreateIndexBuilder
	NewDropIndexBuilder(ctx context.Context, index string) *redis.DropIndexBuilder
	RedisClient(ctx context.Context) *redis.Client
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
}

type RedisOriginCmdCache struct {
	client            *redis.Client
	KeyPrefix         string
	encoding          encoding.Encoding
	DefaultExpireTime time.Duration
	newObject         func() interface{}
}

func NewRedisOriginCmdCache(client *redis.Client, keyPrefix string) *RedisOriginCmdCache {
	return &RedisOriginCmdCache{
		client:    client,
		KeyPrefix: keyPrefix,
	}
}

// https://redis.com.cn/commands.html reids命令集, 后续还需要其它,可以在这里拓展添加

// Set 命令(同时是会跟expire结合使用) : SET key value [expiration options]
// 1. SET mykey "myvalue"
// 2. SET mykey "myvalue" EX 60
func (c *RedisOriginCmdCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return c.client.Set(ctx, key, value, expiration)
}

// Get 命令: 从 Redis 数据库中检索与指定键关联的字符串值
// GET mykey
func (c *RedisOriginCmdCache) Get(ctx context.Context, key string) *redis.StringCmd {
	return c.client.Get(ctx, key)
}

// Del 命令: 用于从 Redis 数据库中删除指定的键值对。 该命令可以同时删除多个 key，并返回成功删除的 key 的数量
// DEL key[key ...]
func (c *RedisOriginCmdCache) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	if len(keys) == 0 {
		return c.client.Del(ctx)
	}
	return c.client.Del(ctx, keys...)
}

// SetNX 命令: 用于在 Redis 数据库中设置键值对，但仅当 key 不存在时才设置。 如果 key 已经存在，则该命令不执行任何操作。该命令返回 1 表示设置成功，返回 0 表示 key 已经存在，设置失败
// SET key value [expiration] NX
func (c *RedisOriginCmdCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return c.client.SetNX(ctx, key, value, expiration)
}

// Exists 命令: 在执行诸如 GET、SET 或 DEL 等操作之前，可以使用 Exists 命令来验证 key 是否存在，避免不必要的操作或错误。
// EXISTS key1
// EXISTS key1 key2 key 3
func (c *RedisOriginCmdCache) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	if len(keys) == 0 {
		return c.client.Exists(ctx)
	}
	return c.client.Exists(ctx, keys...)
}

// Expire 命令: 为指定的 key 设置过期时间（TTL，Time To Live），以秒为单位
// EXPIRE key seconds[NX|XX|GT|LT]
/*
	Redis 通过两种方式删除过期 key：
		被动删除： 当客户端尝试访问 key 时，如果 key 已经过期，则 Redis 会删除该 key。
		主动删除： Redis 会定期随机抽取一些 key，检查它们是否过期，如果过期则删除。
	使用场景
		缓存控制： 可以利用过期时间来控制缓存的有效期，自动清理过期数据.
		会话管理： 可以为会话 key 设置过期时间，实现会话的自动失效。
		限流： 可以结合计数器和过期时间，实现简单的限流策略。
*/
func (c *RedisOriginCmdCache) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return c.client.Expire(ctx, key, expiration)
}

// Incr 命令: Incr 命令用于原子性地递增 Redis 中存储的数字。它对应于 Redis 原生的 INCR 命令。
// INCR keyname -- Incr 命令返回 key 增加后的值
/*
	使用场景
		计数器: 可以用于实现各种计数器，例如页面访问量、点赞数等.
		频率限制: 通过递增计数器来跟踪和执行频率限制.
		原子操作: 在多客户端环境中，确保递增计数器的原子性.
		会话管理: 用于生成唯一的会话 ID 或序列号
*/
func (c *RedisOriginCmdCache) Incr(ctx context.Context, key string) *redis.IntCmd {
	return c.client.Incr(ctx, key)
}

// Decr 命令: Decr 命令对应于 Redis 的原生命令 DECR. DECR key 将 key 中存储的数字值减 1
// Decr keyname
func (c *RedisOriginCmdCache) Decr(ctx context.Context, key string) *redis.IntCmd {
	return c.client.Decr(ctx, key)
}

// HSet 命令: 用于设置哈希表中指定字段的值。如果哈希表不存在，会创建一个新的哈希表并设置值。如果字段已存在，旧值将被覆盖
// HSET key field value
/*
	1. HSet("myhash", "key1", "value1", "key2", "value2")
	2. HSet("myhash", []string{"key1", "value1", "key2", "value2"})
	3. HSet("myhash", map[string]interface{}{"key1": "value1", "key2": "value2"})
	type MyHash struct { Key1 string `redis:"key1"`; Key2 int `redis:"key2"` }
	4. HSet("myhash", MyHash{"value1", "value2"}) Warn: redis-server >= 4.0
*/
func (c *RedisOriginCmdCache) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.client.HSet(ctx, key, values...)
}

// HSetNX 命令: HSetNX 用于设置存储在指定键中的哈希表中某个字段的值，但前提是该字段尚未存在。这对于避免覆盖哈希表中现有的数据非常有用。
// HSETNX key field value
/*
	HSETNX myhash field1 "hello"
	HSETNX myhash field1 "world"
	HGET myhash field1
	返回的结果是hello
*/
func (c *RedisOriginCmdCache) HSetNX(ctx context.Context, key, field string, value interface{}) *redis.BoolCmd {
	return c.client.HSetNX(ctx, key, field, value)
}

// HGet 命令: 用于获取哈希表中指定字段的值
// HGET key field
func (c *RedisOriginCmdCache) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return c.client.HGet(ctx, key, field)
}

// HMSet 命令: 用于同时设置哈希表中多个字段的值 注意：在 Redis 4.0.0 之后，HMSET 被认为已弃用，推荐使用 HSET 命令代替
// 可以使用 HSet 结合 map 来实现类似的功能
// HMSET key field value[field value ...]
// 返回值：该命令返回一个整数回复：
// 如果该字段是哈希表中的一个新字段，并且值已成功设置，则返回 1。
// 如果该字段已存在，并且未执行任何操作，则返回 0。
func (c *RedisOriginCmdCache) HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd {
	return c.client.HMSet(ctx, key, values...)
}

// HMGet 命令: 用于同时获取哈希表中多个字段的值
// HMGET key field[field ...]
func (c *RedisOriginCmdCache) HMGet(ctx context.Context, key string, field ...string) *redis.SliceCmd {
	return c.client.HMGet(ctx, key, field...)
}

// HDel 命令: 用于删除哈希表中的一个或多个指定字段
// HDEL key field[field ...]
func (c *RedisOriginCmdCache) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	if len(fields) == 0 {
		return c.client.HDel(ctx, key)
	}
	return c.client.HDel(ctx, key, fields...)
}

// LPush 命令:LPush 命令将一个或多个值插入到列表的头部（左侧）。如果 key 不存在，则在执行 push 操作之前会创建一个空列表。如果 key 对应的值不是列表，则返回错误
// LPUSH key element[element ...]
// 返回值：执行 push 操作后，列表的长度.
func (c *RedisOriginCmdCache) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.client.LPush(ctx, key, values...)
}

// RPush 命令: RPush 命令将一个或多个值追加到列表的尾部（右侧）。如果 key 不存在，则在执行 push 操作之前会创建一个空列表。如果 key 对应的值不是列表，则返回错误
// RPUSH key element[element ...]
// 返回值：执行 push 操作后，列表的长度.
func (c *RedisOriginCmdCache) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.client.RPush(ctx, key, values...)
}

// LPop 命令: LPop 命令移除并返回列表的第一个元素（列表头部）。可以指定移除元素的数量。如果 key 不存在，则返回 nil
// LPOP key[count]
// 返回值：被移除的元素的值。如果列表为空，则返回 nil
func (c *RedisOriginCmdCache) LPop(ctx context.Context, key string) *redis.StringCmd {
	return c.client.LPop(ctx, key)
}

// RPop 命令: RPop 命令移除并返回列表的最后一个元素（列表尾部）。可以指定移除元素的数量。如果 key 不存在，则返回 nil
// RPOP key[count]
// 返回值：被移除的元素的值。如果列表为空，则返回 nil
func (c *RedisOriginCmdCache) RPop(ctx context.Context, key string) *redis.StringCmd {
	return c.client.RPop(ctx, key)
}

// SAdd 命令: SAdd 命令将一个或多个成员添加到存储在 key 的集合中。 已经存在于集合中的成员将被忽略。如果 key 不存在，则在添加指定成员之前创建一个新集合。当存储在 key 中的值不是集合时，返回错误
// SADD key member[member ...]
// 返回值：添加到集合中的元素数量，不包括集合中已存在的元素。
func (c *RedisOriginCmdCache) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return c.client.SAdd(ctx, key, members...)
}

// SMembers 命令: SMembers 命令返回存储在 key 的集合中的所有成员。
// SMEMBERS key
// 返回值：集合中所有成员的列表。
func (c *RedisOriginCmdCache) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return c.client.SMembers(ctx, key)
}

// ZAdd 命令: ZAdd 命令将一个或多个成员添加到有序集合中，或者更新已经存在成员的分数（score）。有序集合中的成员根据它们的分数进行排序（从小到大）。如果 key 不存在，则创建一个新的有序集合。
/*
	原生命令：ZADD key[NX|XX][GT|LT][CH][INCR] score member[score member ...]
		NX: 只添加新成员，不更新已存在成员。
		XX: 只更新已存在成员的分数，不添加新成员。
		GT: 仅在提供的 score 大于当前 score 时才更新现有元素。此标志不阻止添加新元素。
		LT: 仅在提供的 score 小于当前 score 时才更新现有元素。此标志不阻止添加新元素。
		CH: 修改返回值为已更改的元素数量（added + updated）。
		INCR: 将成员的分数增加指定数值。
*/
// 返回值：添加到有序集合中的新成员的数量，不包括已更新分数的成员。如果使用了 CH 选项，则返回更改的元素数量。
func (c *RedisOriginCmdCache) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	return c.client.ZAdd(ctx, key, members...)
}

// ZRange 命令: ZRange 命令返回有序集合中指定范围内的元素。 元素被认为是从最低到最高分排序。
// ZRANGE key start stop[WITHSCORES]
// 返回值：指定范围内的元素列表（可选地，包括它们的分数）。
func (c *RedisOriginCmdCache) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return c.client.ZRange(ctx, key, start, stop)
}

// ZRangeWithScores 命令: ZRangeWithScores 命令返回有序集合中指定范围内的元素及其对应的分数。
// ZRANGE key start stop WITHSCORES
// 返回值：指定范围内的元素和分数列表，格式为 element1, score1, element2, score2,...。
func (c *RedisOriginCmdCache) ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	return c.client.ZRangeWithScores(ctx, key, start, stop)
}

// ZIncrBy 命令:ZIncrBy 命令为有序集合中指定成员的分数增加指定数值。如果成员不存在，则添加该成员，并将其初始分数设为 0
// ZINCRBY key increment member
// 返回值：增加之后，成员的新的分数。
func (c *RedisOriginCmdCache) ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd {
	return c.client.ZIncrBy(ctx, key, increment, member)
}

// ZRem 命令: ZRem 命令从有序集合中移除指定的成员。
// ZREM key member[member ...]
// 返回值：从有序集合中移除的成员数量。
func (c *RedisOriginCmdCache) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return c.client.ZRem(ctx, key, members...)
}

// Publish 命令: Publish 命令用于将消息发布到指定频道。任何订阅该频道的客户端都将收到该消息。
// 原生命令：PUBLISH channel message
// 返回值：收到消息的订阅者数量。
func (c *RedisOriginCmdCache) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	return c.client.Publish(ctx, channel, message)
}

// Subscribe 命令: Subscribe 命令用于让客户端订阅一个或多个频道。一旦客户端订阅了频道，它将开始接收发布到这些频道的消息。
// 原生命令：SUBSCRIBE channel[channel ...]
// 返回值：一个 *redis.PubSub 类型的指针，用于接收发布到订阅频道的消息。
func (c *RedisOriginCmdCache) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.client.Subscribe(ctx)
	}
	return c.client.Subscribe(ctx, channels...)
}

// SetArgs 命令:功能：SetArgs 命令用于设置键的值，并允许使用各种选项，如设置过期时间、仅在键不存在时设置 (NX) 或仅在键存在时设置 (XX)。
// 原生命令：SET key value[expiration][NX|XX][GET]
// 返回值: redis.StatusCmd，可以通过调用其 Err() 方法检查是否成功。
func (c *RedisOriginCmdCache) SetArgs(ctx context.Context, key string, value interface{}, a redis.SetArgs) *redis.StatusCmd {
	return c.client.SetArgs(ctx, key, value, a)
}

// Pipelined 命令:Pipelined 和 Pipeline 用于批量执行 Redis 命令，减少客户端与服务器之间的网络往返次数，从而提高性能.
// Pipelined：Pipelined 是一个函数，它接受一个函数作为参数，该函数定义了要在管道中执行的命令。Redis 客户端会自动将这些命令放入管道中，并在函数执行完毕后一次性发送到服务器.
// 原生命令：管道化不是一个单独的 Redis 命令，而是一种客户端优化技术，它允许客户端将多个命令打包在一起发送到服务器，服务器依次执行这些命令并将结果一次性返回给客户端。
// 返回值：Pipelined 返回一个命令切片和一个错误。
func (c *RedisOriginCmdCache) Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return c.client.Pipelined(ctx, fn)
}

// Pipeline 命令:Pipelined 和 Pipeline 用于批量执行 Redis 命令，减少客户端与服务器之间的网络往返次数，从而提高性能.
// Pipeline 命令: Pipeline 是一个接口，允许你手动将命令添加到管道中，然后使用 Exec 方法一次性执行它们.
// 原生命令：管道化不是一个单独的 Redis 命令，而是一种客户端优化技术，它允许客户端将多个命令打包在一起发送到服务器，服务器依次执行这些命令并将结果一次性返回给客户端。
// 返回值：Pipeline 的 Exec 方法返回一个命令切片和一个错误。
func (c *RedisOriginCmdCache) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

// TxPipelined 命令: TxPipelined 和 TxPipeline 类似于 Pipelined 和 Pipeline，但它们在事务中执行命令。事务提供了一种将多个命令作为一个原子操作执行的方式。
// TxPipelined：TxPipelined 是一个函数，它接受一个函数作为参数，该函数定义了要在事务中执行的命令。Redis 客户端会自动将这些命令放入管道中，并在函数执行完毕后一次性发送到服务器。
// 原生命令：MULTI, command1, command2,..., EXEC。MULTI 命令表示事务的开始，EXEC 命令表示事务的结束。
// 返回值：TxPipeline 的 Exec 方法返回一个命令切片和一个错误。
func (c *RedisOriginCmdCache) TxPipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return c.client.TxPipelined(ctx, fn)
}

// TxPipeline 命令: TxPipeline 是一个接口，允许你手动将命令添加到事务管道中，然后使用 Exec 方法一次性执行它们。
// 原生命令：MULTI, command1, command2,..., EXEC。MULTI 命令表示事务的开始，EXEC 命令表示事务的结束。
// 返回值：TxPipeline 的 Exec 方法返回一个命令切片和一个错误。
func (c *RedisOriginCmdCache) TxPipeline() redis.Pipeliner {
	return c.client.TxPipeline()
}

// Process 命令:功能：Process 方法用于执行单个 Redis 命令。它接受一个 redis.Cmder 接口类型的参数，该接口表示要执行的命令。这个方法通常在较低层次的抽象中使用，允许更细粒度的控制命令的执行过程。
// 原生命令：取决于传递给 Process 方法的 Cmder 接口的具体命令。
// 返回值：返回一个 error，表示命令执行过程中是否发生错误。
func (c *RedisOriginCmdCache) Process(ctx context.Context, cmd redis.Cmder) error {
	return c.client.Process(ctx, cmd)
}

// Options 命令:功能：Options 结构体用于配置 Redis 客户端的各种选项，例如连接地址、密码、数据库、连接池大小、超时时间等。通过 Options，可以自定义客户端的行为以满足特定的应用需求。
/*
	字段（部分）：
		Addr：Redis 服务器的地址。
		Password：连接 Redis 服务器所需的密码。
		DB：要选择的数据库编号。
		PoolSize：连接池的大小。
		DialTimeout：建立连接的超时时间。
		ReadTimeout：读取数据的超时时间。
		WriteTimeout：写入数据的超时时间。
*/
func (c *RedisOriginCmdCache) Options() *redis.Options {
	return c.client.Options()
}

// PoolStats 命令:功能：PoolStats 结构体提供了关于 Redis 客户端连接池的统计信息。这些信息可以帮助你监控和调整连接池的性能。
/*
	字段：
		Hits：连接池中成功获取连接的次数。
		Misses：连接池中未能找到空闲连接而需要创建新连接的次数。
		Timeouts：获取连接超时的次数。
		TotalConns：连接池中总的连接数量。
		IdleConns：连接池中空闲的连接数量。
		StaleConns：从连接池中移除的过期连接数量。
*/
func (c *RedisOriginCmdCache) PoolStats() *redis.PoolStats {
	return c.client.PoolStats()
}

// PSubscribe 命令:功能: PSubscribe 命令用于订阅与给定模式匹配的频道。与 Subscribe 命令不同，PSubscribe 允许使用通配符模式来一次性订阅多个频道。
// 原生命令：PSUBSCRIBE pattern[pattern ...]
/*
	模式匹配：
		h?llo 订阅 hello, hallo 和 hxllo
		h*llo 订阅 hllo 和 heeeello
		h[ae]llo 订阅 hello 和 hallo, 但不订阅 hillo
*/
//返回值: 与 Subscribe 类似，PSubscribe 返回一个 *redis.PubSub 类型的指针，用于接收与模式匹配的频道上发布的消息。
func (c *RedisOriginCmdCache) PSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.client.PSubscribe(ctx)
	}
	return c.client.PSubscribe(ctx, channels...)
}

// SSubscribe 命令:功能：SSubscribe 命令用于订阅指定的 shard channels。在 Redis 集群中，shard channels 通过相同的算法分配给槽位，该算法用于将键分配给槽位。客户端可以订阅覆盖某个槽位的节点（主节点 / 复制节点）以接收发布的消息。
// 原生命令：SSUBSCRIBE shardchannel[shardchannel ...]
// 约束：在给定的 SSubscribe 调用中，所有指定的 shard channels 需要属于单个槽位。客户端可以通过单独的 SSubscribe 调用来订阅跨不同槽位的频道。
// 返回值：此命令不返回任何内容。相反，对于每个 shard channel，都会推送一条消息，其中第一个元素是字符串 "ssubscribe"，以确认命令成功。
func (c *RedisOriginCmdCache) SSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.client.SSubscribe(ctx)
	}
	return c.client.SSubscribe(ctx, channels...)
}

// NewSearchBuilder 命令:命令:功能：NewSearchBuilder 用于构建复杂的搜索查询。它提供了一种流式（链式）的 API，允许你构造搜索请求的各个部分，例如要搜索的索引、查询字符串、返回字段、排序方式、分页等。这个构建器模式使得创建复杂的搜索查询更加清晰和易于维护。
// 原生命令：FT.SEARCH index_name query_string 以及其他相关的搜索选项。
func (c *RedisOriginCmdCache) NewSearchBuilder(ctx context.Context, index, query string) *redis.SearchBuilder {
	return c.client.NewSearchBuilder(ctx, index, query)
}

// NewAggregateBuilder 功能：NewAggregateBuilder 用于构建聚合查询。聚合查询用于对搜索结果进行分组、排序、转换和统计，类似于 SQL 中的 GROUP BY 和聚合函数。NewAggregateBuilder 允许你通过链式调用定义聚合管道的各个阶段，例如分组、归约、排序、应用表达式、过滤和限制结果集.
// 原生命令：FT.AGGREGATE index_name query_string 以及其他聚合选项.
func (c *RedisOriginCmdCache) NewAggregateBuilder(ctx context.Context, index, query string) *redis.AggregateBuilder {
	return c.client.NewAggregateBuilder(ctx, index, query)
}

// NewCreateIndexBuilder 命令: 功能：NewCreateIndexBuilder 用于构建创建索引的命令。通过此构建器，你可以指定索引的名称、要索引的字段、字段的类型（TEXT、TAG、NUMERIC、GEO 等）以及其他索引选项。索引对于 Redis Search 的性能至关重要.
// 原生命令：FT.CREATE index_name 以及相关的 schema 定义和索引选项
func (c *RedisOriginCmdCache) NewCreateIndexBuilder(ctx context.Context, index string) *redis.CreateIndexBuilder {
	return c.client.NewCreateIndexBuilder(ctx, index)
}

// NewDropIndexBuilder 命令: 功能：NewDropIndexBuilder 用于构建删除索引的命令。删除索引会移除与该索引相关的所有数据和配置。
// 原生命令：FT.DROPINDEX index_name。
func (c *RedisOriginCmdCache) NewDropIndexBuilder(ctx context.Context, index string) *redis.DropIndexBuilder {
	return c.client.NewDropIndexBuilder(ctx, index)
}

// RedisClient 返回底层原始的Client，用于需要直接访问底层 API 的特殊场景
func (c *RedisOriginCmdCache) RedisClient(ctx context.Context) *redis.Client {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client
}

// Eval 允许执行 Lua 脚本（EVAL）——传入的 keys 会被包装器加上前缀。
// 功能：Eval 命令用于执行 Lua 脚本。它允许你将一段 Lua 脚本发送到 Redis 服务器执行，从而实现复杂的逻辑和原子操作。Eval 命令是 Redis 扩展功能的强大工具。
// 原生命令：EVAL script numkeys key[key ...] arg[arg ...]
// numkeys: 整数，代表 key 的个数，用于告诉 redis 哪些参数是 key
func (c *RedisOriginCmdCache) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	if len(keys) == 0 {
		return c.client.Eval(ctx, script, nil, args...)
	}
	return c.client.Eval(ctx, script, keys, args...)
}
