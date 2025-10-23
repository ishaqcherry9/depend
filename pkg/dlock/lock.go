package dlock

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

// ---------------------  Atomic Based Locks (基于原子操作的锁) ---------------------

// SpinLock 自旋锁，适用于短时间持有的锁
type SpinLock struct {
	state  int32 // 0: unlocked, 1: locked
	closed int32 // 0: open, 1: closed, 用于标记锁是否已关闭
}

// Lock 获取锁，自旋等待直到获取锁或发生错误
func (sl *SpinLock) Lock(ctx context.Context) error {
	if atomic.LoadInt32(&sl.closed) == 1 {
		return fmt.Errorf("spinlock is closed") // 锁已关闭，返回错误
	}
	for !atomic.CompareAndSwapInt32(&sl.state, 0, 1) {
		runtime.Gosched() // 让出 CPU 时间片，避免 CPU 占用过高
	}
	return nil
}

// Unlock 释放锁
func (sl *SpinLock) Unlock(ctx context.Context) error {
	if atomic.LoadInt32(&sl.closed) == 1 {
		return fmt.Errorf("spinlock is closed") // 锁已关闭，返回错误
	}
	atomic.StoreInt32(&sl.state, 0)
	return nil
}

// TryLock 尝试获取锁，在指定时间内自旋等待
func (sl *SpinLock) TryLock(ctx context.Context, timeout time.Duration) (bool, error) {
	if atomic.LoadInt32(&sl.closed) == 1 {
		return false, fmt.Errorf("spinlock is closed") // 锁已关闭，返回错误
	}
	deadline := time.Now().Add(timeout) // 计算截止时间
	for time.Now().Before(deadline) {   // 在截止时间之前循环尝试
		if atomic.CompareAndSwapInt32(&sl.state, 0, 1) {
			return true, nil // 成功获取锁
		}
		runtime.Gosched()            // 让出 CPU 时间片
		time.Sleep(time.Millisecond) // 避免过度自旋，稍微等待
	}
	return false, nil // 超时，获取锁失败
}

// Close 关闭 SpinLock，防止继续使用
func (sl *SpinLock) Close() error {
	if atomic.CompareAndSwapInt32(&sl.closed, 0, 1) { // 确保只关闭一次
		return nil
	}
	return fmt.Errorf("spinlock already closed")
}

// AtomicMutex 互斥锁，使用原子操作实现
type AtomicMutex struct {
	state  int32 // 0: unlocked, 1: locked
	closed int32 // 0: open, 1: closed
}

// Lock 获取锁，阻塞直到获取锁或发生错误
func (m *AtomicMutex) Lock(ctx context.Context) error {
	if atomic.LoadInt32(&m.closed) == 1 {
		return fmt.Errorf("atomicmutex is closed") // 锁已关闭，返回错误
	}
	for !atomic.CompareAndSwapInt32(&m.state, 0, 1) {
		runtime.Gosched() // 让出 CPU 时间片
	}
	return nil
}

// Unlock 释放锁
func (m *AtomicMutex) Unlock(ctx context.Context) error {
	if atomic.LoadInt32(&m.closed) == 1 {
		return fmt.Errorf("atomicmutex is closed") // 锁已关闭，返回错误
	}
	atomic.StoreInt32(&m.state, 0)
	return nil
}

// TryLock 尝试获取锁，在指定时间内等待
func (m *AtomicMutex) TryLock(ctx context.Context, timeout time.Duration) (bool, error) {
	if atomic.LoadInt32(&m.closed) == 1 {
		return false, fmt.Errorf("atomicmutex is closed") // 锁已关闭，返回错误
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if atomic.CompareAndSwapInt32(&m.state, 0, 1) {
			return true, nil
		}
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
	return false, nil
}

// Close 关闭 AtomicMutex，防止继续使用
func (m *AtomicMutex) Close() error {
	if atomic.CompareAndSwapInt32(&m.closed, 0, 1) {
		return nil
	}
	return fmt.Errorf("atomicmutex already closed")
}

// CASLock 基于 CAS (Compare and Swap) 的锁
type CASLock struct {
	lock   int32 // 0: unlocked, 1: locked
	closed int32 // 0: open, 1: closed
}

// Lock 获取锁，阻塞直到获取锁或发生错误
func (l *CASLock) Lock(ctx context.Context) error {
	if atomic.LoadInt32(&l.closed) == 1 {
		return fmt.Errorf("caslock is closed") // 锁已关闭，返回错误
	}
	for !atomic.CompareAndSwapInt32(&l.lock, 0, 1) {
		runtime.Gosched() // 让出 CPU 时间片
	}
	return nil
}

// Unlock 释放锁
func (l *CASLock) Unlock(ctx context.Context) error {
	if atomic.LoadInt32(&l.closed) == 1 {
		return fmt.Errorf("caslock is closed") // 锁已关闭，返回错误
	}
	atomic.StoreInt32(&l.lock, 0)
	return nil
}

// TryLock 尝试获取锁，在指定时间内等待
func (l *CASLock) TryLock(ctx context.Context, timeout time.Duration) (bool, error) {
	if atomic.LoadInt32(&l.closed) == 1 {
		return false, fmt.Errorf("caslock is closed") // 锁已关闭，返回错误
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if atomic.CompareAndSwapInt32(&l.lock, 0, 1) {
			return true, nil
		}
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
	return false, nil
}

// Close 关闭 CASLock，防止继续使用
func (l *CASLock) Close() error {
	if atomic.CompareAndSwapInt32(&l.closed, 0, 1) {
		return nil
	}
	return fmt.Errorf("caslock already closed")
}

// ---------------------  RWLock (读写锁) ---------------------

// RWLock 读写锁，允许多个 reader 并发读，但 writer 独占写
type RWLock struct {
	readerCount int32 // 读者数量
	writer      int32 // 0: unlocked, 1: locked
	closed      int32 // 0: open, 1: closed
}

// RLock 获取读锁
func (rw *RWLock) RLock(ctx context.Context) error {
	if atomic.LoadInt32(&rw.closed) == 1 {
		return fmt.Errorf("rwlock is closed") // 锁已关闭，返回错误
	}
	for {
		count := atomic.LoadInt32(&rw.readerCount)
		if atomic.CompareAndSwapInt32(&rw.readerCount, count, count+1) {
			return nil
		}
		runtime.Gosched()
	}
}

// RUnlock 释放读锁
func (rw *RWLock) RUnlock(ctx context.Context) error {
	if atomic.LoadInt32(&rw.closed) == 1 {
		return fmt.Errorf("rwlock is closed") // 锁已关闭，返回错误
	}
	atomic.AddInt32(&rw.readerCount, -1)
	return nil
}

// Lock 获取写锁
func (rw *RWLock) Lock(ctx context.Context) error {
	if atomic.LoadInt32(&rw.closed) == 1 {
		return fmt.Errorf("rwlock is closed") // 锁已关闭，返回错误
	}
	for atomic.LoadInt32(&rw.readerCount) != 0 || !atomic.CompareAndSwapInt32(&rw.writer, 0, 1) {
		runtime.Gosched()
	}
	return nil
}

// Unlock 释放写锁
func (rw *RWLock) Unlock(ctx context.Context) error {
	if atomic.LoadInt32(&rw.closed) == 1 {
		return fmt.Errorf("rwlock is closed") // 锁已关闭，返回错误
	}
	atomic.StoreInt32(&rw.writer, 0)
	return nil
}

// TryLock 尝试获取写锁，带超时时间 (不支持读锁的 TryLock)
func (rw *RWLock) TryLock(ctx context.Context, timeout time.Duration) (bool, error) {
	if atomic.LoadInt32(&rw.closed) == 1 {
		return false, fmt.Errorf("rwlock is closed") // 锁已关闭，返回错误
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&rw.readerCount) == 0 && atomic.CompareAndSwapInt32(&rw.writer, 0, 1) {
			return true, nil
		}
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
	return false, nil
}

// Close 关闭 RWLock，防止继续使用
func (rw *RWLock) Close() error {
	if atomic.CompareAndSwapInt32(&rw.closed, 0, 1) {
		return nil
	}
	return fmt.Errorf("rwlock already closed")
}

// ---------------------  TicketLock (票据锁) ---------------------

// TicketLock 基于 Ticket 的锁，保证 FIFO (先进先出)
type TicketLock struct {
	serviceTicket int64 // 服务号，表示下一个要被服务的票据号码
	currentTicket int64 // 当前号，表示当前正在被服务的票据号码
	closed        int32 // 0: open, 1: closed
}

// Lock 获取锁，获取票据并等待轮到自己
func (tl *TicketLock) Lock(ctx context.Context) error {
	if atomic.LoadInt32(&tl.closed) == 1 {
		return fmt.Errorf("ticketlock is closed") // 锁已关闭，返回错误
	}
	ticket := atomic.AddInt64(&tl.serviceTicket, 1)     // 获取服务号
	for atomic.LoadInt64(&tl.currentTicket) != ticket { // 等待轮到当前号
		runtime.Gosched()
	}
	return nil
}

// Unlock 释放锁，增加当前号，通知下一个等待者
func (tl *TicketLock) Unlock(ctx context.Context) error {
	if atomic.LoadInt32(&tl.closed) == 1 {
		return fmt.Errorf("ticketlock is closed") // 锁已关闭，返回错误
	}
	atomic.AddInt64(&tl.currentTicket, 1) // 增加当前号
	return nil
}

// TryLock 尝试获取锁，在指定时间内等待
func (tl *TicketLock) TryLock(ctx context.Context, timeout time.Duration) (bool, error) {
	if atomic.LoadInt32(&tl.closed) == 1 {
		return false, fmt.Errorf("ticketlock is closed") // 锁已关闭，返回错误
	}
	startTime := time.Now()
	ticket := atomic.AddInt64(&tl.serviceTicket, 1) // 获取服务号

	for time.Since(startTime) < timeout {
		if atomic.LoadInt64(&tl.currentTicket) == ticket { // 等于当前号，获取锁成功
			return true, nil
		}
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}

	atomic.AddInt64(&tl.serviceTicket, -1) // 超时，需要回滚服务号，保证公平
	return false, nil
}

// Close 关闭 TicketLock，防止继续使用
func (tl *TicketLock) Close() error {
	if atomic.CompareAndSwapInt32(&tl.closed, 0, 1) {
		return nil
	}
	return fmt.Errorf("ticketlock already closed")
}

// ---------------------  ALock (数组锁) ---------------------

// ALock 数组锁，基于 Ticket 的锁，减少竞争
type ALock struct {
	flag   []int32 // 标志数组，表示每个槽是否可用
	size   int     // 数组大小，槽的数量
	queue  int64   // 请求队列，用于分配槽
	slot   int     // 当前 Goroutine 占用的槽
	closed int32   // 0: open, 1: closed
}

// NewALock 创建一个新的 ALock 实例
func NewALock(size int) *ALock {
	al := &ALock{
		flag:  make([]int32, size), // 初始化标志数组
		size:  size,
		queue: 0,
	}
	al.flag[0] = 1 // 初始化第一个元素为可用
	return al
}

// Lock 获取锁，获取一个可用槽
func (al *ALock) Lock(ctx context.Context) error {
	if atomic.LoadInt32(&al.closed) == 1 {
		return fmt.Errorf("alock is closed") // 锁已关闭，返回错误
	}
	slot := atomic.AddInt64(&al.queue, 1) % int64(al.size) // 获取槽的索引
	for atomic.LoadInt32(&al.flag[slot]) == 0 {            // 等待槽可用
		runtime.Gosched()
	}
	al.slot = int(slot) // 记录当前 Goroutine 占用的槽
	return nil
}

// Unlock 释放锁，释放当前占用的槽
func (al *ALock) Unlock(ctx context.Context) error {
	if atomic.LoadInt32(&al.closed) == 1 {
		return fmt.Errorf("alock is closed") // 锁已关闭，返回错误
	}
	atomic.StoreInt32(&al.flag[al.slot], 0) // 释放当前槽
	newSlot := (al.slot + 1) % al.size      // 获取下一个槽
	atomic.StoreInt32(&al.flag[newSlot], 1) // 释放下一个槽，使得下一个等待者可以获取锁
	return nil
}

// TryLock 尝试获取锁，在指定时间内等待
func (al *ALock) TryLock(ctx context.Context, timeout time.Duration) (bool, error) {
	if atomic.LoadInt32(&al.closed) == 1 {
		return false, fmt.Errorf("alock is closed") // 锁已关闭，返回错误
	}
	startTime := time.Now()
	slot := atomic.AddInt64(&al.queue, 1) % int64(al.size) // 获取槽的索引

	for time.Since(startTime) < timeout {
		if atomic.LoadInt32(&al.flag[slot]) == 0 { // 槽不可用
			runtime.Gosched()
			time.Sleep(time.Millisecond)
			continue
		}
		al.slot = int(slot) // 记录当前 Goroutine 占用的槽
		return true, nil
	}

	atomic.AddInt64(&al.queue, -1) // 超时，需要回滚 queue，否则会影响后续 lock 的顺序，保证公平
	return false, nil
}

// Close 关闭 ALock，防止继续使用
func (al *ALock) Close() error {
	if atomic.CompareAndSwapInt32(&al.closed, 0, 1) {
		return nil
	}
	return fmt.Errorf("alock already closed")
}

// ---------------------  CLHLock (链表锁) ---------------------

// CLHLock 实现了 Locker 接口，提供 CLH 锁的功能
type CLHLock struct {
	tail   unsafe.Pointer // 指向链表尾部的指针
	myNode unsafe.Pointer // 指向当前 Goroutine 对应的节点的指针
	closed int32          // 0: open, 1: closed
}

type clhNode struct {
	locked int32    // 节点是否已锁定的标志
	next   *clhNode // 指向下一个节点的指针
}

// NewCLHLock 创建一个新的 CLHLock 实例
func NewCLHLock() *CLHLock {
	node := new(clhNode) // 创建一个空节点
	return &CLHLock{
		tail:   unsafe.Pointer(node), // 初始化 tail 指针指向空节点
		myNode: unsafe.Pointer(node), // 初始化 myNode 指针指向空节点
	}
}

// Lock 获取锁
func (l *CLHLock) Lock(ctx context.Context) error {
	if atomic.LoadInt32(&l.closed) == 1 {
		return fmt.Errorf("clhlock is closed")
	}

	myNode := new(clhNode) // 创建一个新的节点
	myNode.locked = 1      // 设置节点为 locked 状态
	myNodePtr := unsafe.Pointer(myNode)

	// 将当前节点添加到链表尾部
	prevNodePtr := atomic.SwapPointer(&l.tail, myNodePtr)

	if prevNodePtr != nil && prevNodePtr != myNodePtr {
		// 如果前驱节点不为空，则等待前驱节点释放锁
		for atomic.LoadInt32(&(*clhNode)(prevNodePtr).locked) == 1 {
			runtime.Gosched() // 让出 CPU 时间片
		}
	}

	// 成功获取锁
	l.myNode = myNodePtr
	return nil
}

// Unlock 释放锁
func (l *CLHLock) Unlock(ctx context.Context) error {
	if atomic.LoadInt32(&l.closed) == 1 {
		return fmt.Errorf("clhlock is closed")
	}

	myNodePtr := l.myNode
	if myNodePtr == nil {
		return fmt.Errorf("unlock called on unlocked lock")
	}

	myNode := (*clhNode)(myNodePtr)
	if myNode.next != nil {
		// 如果有后继节点，则释放锁
		atomic.StoreInt32(&myNode.locked, 0)
	} else {
		// 如果没有后继节点，则尝试将 tail 指针设置为 nil
		if atomic.CompareAndSwapPointer(&l.tail, myNodePtr, nil) {
			// CAS 成功，表示当前节点是链表中唯一的节点
			return nil
		}

		// CAS 失败，表示有其他 Goroutine 正在修改链表尾部
		// 需要等待后继节点出现
		for myNode.next == nil {
			runtime.Gosched() // 让出 CPU 时间片
		}

		// 释放锁
		atomic.StoreInt32(&myNode.locked, 0)
	}

	return nil
}

// TryLock 尝试获取锁，带超时时间
func (l *CLHLock) TryLock(ctx context.Context, timeout time.Duration) (bool, error) {
	if atomic.LoadInt32(&l.closed) == 1 {
		return false, fmt.Errorf("clhlock is closed")
	}

	startTime := time.Now()
	myNode := new(clhNode) // 创建一个新的节点
	myNode.locked = 1      // 设置节点为 locked 状态
	myNodePtr := unsafe.Pointer(myNode)

	for time.Since(startTime) < timeout { // 在超时时间内尝试获取锁
		// 尝试将当前节点添加到链表尾部
		prevNodePtr := atomic.LoadPointer(&l.tail) // 读取当前的尾部节点

		// 使用 CAS 操作尝试将当前节点设置为新的尾部节点
		if atomic.CompareAndSwapPointer(&l.tail, prevNodePtr, myNodePtr) {
			// CAS 成功，表示当前节点成功添加到链表尾部
			// 现在需要等待前驱节点释放锁

			if prevNodePtr == unsafe.Pointer(nil) || prevNodePtr == myNodePtr {
				// 如果前驱节点为空，或者前驱节点是自己，表示当前链表为空，或者只有一个节点
				// 此时可以直接获取锁
				l.myNode = myNodePtr
				return true, nil
			}

			// 等待前驱节点释放锁
			for atomic.LoadInt32(&(*clhNode)(prevNodePtr).locked) == 1 && time.Since(startTime) < timeout {
				runtime.Gosched() // 让出 CPU 时间片
				time.Sleep(time.Millisecond)
			}

			// 检查是否超时
			if time.Since(startTime) >= timeout {
				// 超时，需要将当前节点从链表中移除
				// 但是由于 CLH 锁的特性，无法直接将当前节点从链表中移除
				// 这里只能放弃获取锁，并返回 false

				// 理论上，这里应该将 tail 指针恢复到 prevNodePtr，
				// 但是由于可能存在并发修改 tail 指针的情况，
				// 简单的恢复 tail 指针可能会导致链表结构出现问题。
				// 因此，这里选择直接返回 false，放弃获取锁。

				return false, nil
			}

			// 成功获取锁
			l.myNode = myNodePtr
			return true, nil
		} else {
			// CAS 失败，表示有其他 Goroutine 正在修改链表尾部
			// 需要重新尝试
			runtime.Gosched()
			time.Sleep(time.Millisecond)
		}
	}

	// 超时，获取锁失败
	return false, nil
}

// Close 关闭锁
func (l *CLHLock) Close() error {
	atomic.StoreInt32(&l.closed, 1)
	return nil
}
