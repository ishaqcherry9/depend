package utils

import (
	"fmt"
	"time"
)

// Snowflake 结构体，用于生成唯一 ID
type Snowflake struct {
	epoch         int64 // 起始时间戳 (毫秒)
	workerID      int64 // 工作节点 ID (0-31)
	datacenterID  int64 // 数据中心 ID (0-31)
	sequenceMask  int64 // 序列号掩码 (4095 = 2^12 - 1)
	sequence      int64 // 序列号 (0-sequenceMask)
	lastTimestamp int64 // 上次生成 ID 的时间戳 (毫秒)

	// 可配置的位数
	timestampBits  uint8 // 时间戳位数
	datacenterBits uint8 // 数据中心位数
	workerBits     uint8 // 机器ID位数
	sequenceBits   uint8 // 序列号位数

	maxWorkerID       int64 // 最大worker ID
	maxDatacenterID   int64 // 最大datacenter ID
	workerIDShift     uint8
	datacenterIDShift uint8
	timestampShift    uint8
}

// Options 定义 Snowflake 的配置选项
type Options struct {
	Epoch          int64 // 起始时间戳 (毫秒)
	WorkerID       int64 // 工作节点 ID (0-maxWorkerID)
	DatacenterID   int64 // 数据中心 ID (0-maxDatacenterID)
	TimestampBits  uint8 // 时间戳位数
	DatacenterBits uint8 // 数据中心位数
	WorkerBits     uint8 // 机器ID位数
	SequenceBits   uint8 // 序列号位数
}

// NewSnowflake 创建一个新的 Snowflake 实例
// epoch: 起始时间戳 (毫秒)，建议使用一个固定的时间点
// workerID: 工作节点 ID (0-31)，用于区分不同的工作节点
// datacenterID: 数据中心 ID (0-31)，用于区分不同的数据中心
func NewSnowflake(options Options) (*Snowflake, error) {
	// 检查 options 参数有效性
	if options.Epoch == 0 {
		options.Epoch = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	}

	if options.DatacenterBits == 0 {
		options.DatacenterBits = 5
	}

	if options.WorkerBits == 0 {
		options.WorkerBits = 5
	}

	if options.SequenceBits == 0 {
		options.SequenceBits = 12
	}

	maxWorkerID := int64(-1 ^ (-1 << options.WorkerBits))
	maxDatacenterID := int64(-1 ^ (-1 << options.DatacenterBits))

	// 检查 workerID 和 datacenterID 的有效性
	if options.WorkerID < 0 || options.WorkerID > maxWorkerID {
		return nil, fmt.Errorf("workerID must be between 0 and %d", maxWorkerID)
	}
	if options.DatacenterID < 0 || options.DatacenterID > maxDatacenterID {
		return nil, fmt.Errorf("datacenterID must be between 0 and %d", maxDatacenterID)
	}

	// 移位计算
	workerIDShift := options.SequenceBits
	datacenterIDShift := options.SequenceBits + options.WorkerBits
	timestampShift := options.SequenceBits + options.WorkerBits + options.DatacenterBits

	return &Snowflake{
		epoch:         options.Epoch,
		workerID:      options.WorkerID,
		datacenterID:  options.DatacenterID,
		sequenceMask:  int64(-1 ^ (-1 << options.SequenceBits)),
		sequence:      0,
		lastTimestamp: 0,

		timestampBits:  options.TimestampBits,
		datacenterBits: options.DatacenterBits,
		workerBits:     options.WorkerBits,
		sequenceBits:   options.SequenceBits,

		maxWorkerID:       maxWorkerID,
		maxDatacenterID:   maxDatacenterID,
		workerIDShift:     workerIDShift,
		datacenterIDShift: datacenterIDShift,
		timestampShift:    timestampShift,
	}, nil
}

// GenerateID 生成一个唯一的 ID
func (sf *Snowflake) GenerateID() (int64, error) {
	// 获取当前时间戳 (毫秒)
	timestamp := time.Now().UnixMilli()

	// 解决时钟回拨问题
	if timestamp < sf.lastTimestamp {
		// 如果当前时间戳小于上次生成 ID 的时间戳，说明发生了时钟回拨
		// 可以选择等待时钟同步，或者直接返回错误
		// 这里选择等待时钟同步
		timestamp = sf.handleClockBackwards(timestamp)
		return 0, fmt.Errorf("clock is moving backwards, rejecting request until %d", sf.lastTimestamp)
	}

	// 如果当前时间戳等于上次生成 ID 的时间戳，则需要增加序列号
	if timestamp == sf.lastTimestamp {
		// 使用原子操作增加序列号
		sf.sequence = (sf.sequence + 1) & sf.sequenceMask

		// 如果序列号溢出，则需要等待下一毫秒
		if sf.sequence == 0 {
			timestamp = sf.waitUntilNextMillis(sf.lastTimestamp)
		}
	} else {
		// 如果当前时间戳大于上次生成 ID 的时间戳，则重置序列号
		sf.sequence = 0
	}

	// 记录上次生成 ID 的时间戳
	sf.lastTimestamp = timestamp

	// 使用位运算来构建 ID
	id := (timestamp-sf.epoch)<<sf.timestampShift | // 时间戳 (42 bits)
		(sf.datacenterID << sf.datacenterIDShift) | // 数据中心 ID (5 bits)
		(sf.workerID << sf.workerIDShift) | // 工作节点 ID (5 bits)
		sf.sequence // 序列号 (12 bits)

	return id, nil
}

// handleClockBackwards 处理时钟回拨问题
func (sf *Snowflake) handleClockBackwards(timestamp int64) int64 {
	// 更精细的时钟回拨处理
	// 如果回拨时间较短，则等待一段时间
	// 如果回拨时间较长，则可以选择拒绝生成 ID，或者切换到备用方案
	// 这里选择等待一段时间
	if sf.lastTimestamp-timestamp < 5 { // 5ms内等待
		fmt.Println("Clock is moving backwards, waiting for sync...")
		return sf.waitUntilNextMillis(sf.lastTimestamp)
	} else {
		panic(fmt.Sprintf("Clock has moved backwards %d ms. Rejecting id generate request", sf.lastTimestamp-timestamp))
	}
}

// waitUntilNextMillis 等待下一毫秒
func (sf *Snowflake) waitUntilNextMillis(lastTimestamp int64) int64 {
	timestamp := time.Now().UnixMilli()
	for timestamp <= lastTimestamp {
		timestamp = time.Now().UnixMilli()
		time.Sleep(time.Millisecond) // 避免CPU空转
	}
	return timestamp
}
