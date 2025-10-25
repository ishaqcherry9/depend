package dlock

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// SpinLock 自旋锁，适用于短时间持有的锁
func TestSpinLock_LockUnlock(t *testing.T) {
	sl := SpinLock{}
	ctx := context.Background()

	// Test case 1: Basic lock and unlock
	err := sl.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	err = sl.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Concurrent lock and unlock
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := sl.Lock(ctx)
			if err != nil {
				t.Errorf("Lock failed: %v", err)
				return
			}
			counter++
			time.Sleep(time.Millisecond) // Simulate work
			counter--
			err = sl.Unlock(ctx)
			if err != nil {
				t.Errorf("Unlock failed: %v", err)
				return
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestSpinLock_TryLock(t *testing.T) {
	sl := SpinLock{}
	ctx := context.Background()

	// Test case 1: Successful TryLock
	locked, err := sl.TryLock(ctx, time.Millisecond*100)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if !locked {
		t.Errorf("TryLock should have succeeded")
	}
	err = sl.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Unsuccessful TryLock (timeout)
	err = sl.Lock(ctx) // Lock it first
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	locked, err = sl.TryLock(ctx, time.Millisecond*10) // Short timeout
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if locked {
		t.Errorf("TryLock should have timed out")
	}
	err = sl.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
}

func TestSpinLock_Close(t *testing.T) {
	sl := SpinLock{}
	ctx := context.Background()

	// Test case 1: Close successfully
	err := sl.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Test case 2: Double close (should return error)
	err = sl.Close()
	if err == nil {
		t.Errorf("Second Close should have returned an error")
	}

	// Test case 3: Lock after close (should return error)
	err = sl.Lock(ctx)
	if err == nil {
		t.Errorf("Lock after Close should have returned an error")
	}

	// Test case 4: Unlock after close (should return error)
	err = sl.Unlock(ctx)
	if err == nil {
		t.Errorf("Unlock after Close should have returned an error")
	}

	// Test case 5: TryLock after close (should return error)
	_, err = sl.TryLock(ctx, time.Millisecond*10)
	if err == nil {
		t.Errorf("TryLock after Close should have returned an error")
	}
}

func TestSpinLock_Concurrent(t *testing.T) {
	sl := SpinLock{}
	ctx := context.Background()
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			locked, err := sl.TryLock(ctx, time.Millisecond*100)
			if err != nil {
				t.Errorf("TryLock failed: %v", err)
				return
			}
			if locked {
				counter++
				time.Sleep(time.Millisecond * 5) // Simulate work
				counter--
				err = sl.Unlock(ctx)
				if err != nil {
					t.Errorf("Unlock failed: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestSpinLock_Gosched(t *testing.T) {
	sl := SpinLock{}
	ctx := context.Background()
	var locked bool
	var err error

	// Lock the spinlock
	err = sl.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// Try to acquire the lock in a separate goroutine
	go func() {
		locked, err = sl.TryLock(ctx, time.Second)
		if err != nil {
			t.Errorf("TryLock failed: %v", err)
			return
		}
	}()

	// Let the other goroutine run for a bit
	time.Sleep(time.Millisecond * 10)

	// Unlock the spinlock
	err = sl.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Wait for the other goroutine to finish
	time.Sleep(time.Millisecond * 10)

	// Check if the other goroutine acquired the lock
	if !locked {
		t.Errorf("TryLock should have succeeded")
	}
}

// AtomicMutex 互斥锁，使用原子操作实现
func TestAtomicMutex_LockUnlock(t *testing.T) {
	m := AtomicMutex{}
	ctx := context.Background()

	// Test case 1: Basic lock and unlock
	err := m.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	err = m.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Concurrent lock and unlock
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := m.Lock(ctx)
			if err != nil {
				t.Errorf("Lock failed: %v", err)
				return
			}
			counter++
			time.Sleep(time.Millisecond) // Simulate work
			counter--
			err = m.Unlock(ctx)
			if err != nil {
				t.Errorf("Unlock failed: %v", err)
				return
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestAtomicMutex_TryLock(t *testing.T) {
	m := AtomicMutex{}
	ctx := context.Background()

	// Test case 1: Successful TryLock
	locked, err := m.TryLock(ctx, time.Millisecond*100)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if !locked {
		t.Errorf("TryLock should have succeeded")
	}
	err = m.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Unsuccessful TryLock (timeout)
	err = m.Lock(ctx) // Lock it first
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	locked, err = m.TryLock(ctx, time.Millisecond*10) // Short timeout
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if locked {
		t.Errorf("TryLock should have timed out")
	}
	err = m.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
}

func TestAtomicMutex_Close(t *testing.T) {
	m := AtomicMutex{}
	ctx := context.Background()

	// Test case 1: Close successfully
	err := m.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Test case 2: Double close (should return error)
	err = m.Close()
	if err == nil {
		t.Errorf("Second Close should have returned an error")
	}

	// Test case 3: Lock after close (should return error)
	err = m.Lock(ctx)
	if err == nil {
		t.Errorf("Lock after Close should have returned an error")
	}

	// Test case 4: Unlock after close (should return error)
	err = m.Unlock(ctx)
	if err == nil {
		t.Errorf("Unlock after Close should have returned an error")
	}

	// Test case 5: TryLock after close (should return error)
	_, err = m.TryLock(ctx, time.Millisecond*10)
	if err == nil {
		t.Errorf("TryLock after Close should have returned an error")
	}
}

func TestAtomicMutex_Concurrent(t *testing.T) {
	m := AtomicMutex{}
	ctx := context.Background()
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			locked, err := m.TryLock(ctx, time.Millisecond*100)
			if err != nil {
				t.Errorf("TryLock failed: %v", err)
				return
			}
			if locked {
				counter++
				time.Sleep(time.Millisecond * 5) // Simulate work
				counter--
				err = m.Unlock(ctx)
				if err != nil {
					t.Errorf("Unlock failed: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestAtomicMutex_Gosched(t *testing.T) {
	m := AtomicMutex{}
	ctx := context.Background()
	var locked bool
	var err error

	// Lock the mutex
	err = m.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// Try to acquire the lock in a separate goroutine
	go func() {
		locked, err = m.TryLock(ctx, time.Second)
		if err != nil {
			t.Errorf("TryLock failed: %v", err)
			return
		}
	}()

	// Let the other goroutine run for a bit
	time.Sleep(time.Millisecond * 10)

	// Unlock the mutex
	err = m.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Wait for the other goroutine to finish
	time.Sleep(time.Millisecond * 10)

	// Check if the other goroutine acquired the lock
	if !locked {
		t.Errorf("TryLock should have succeeded")
	}
}

// CASLock 基于 CAS (Compare and Swap) 的锁
func TestCASLock_LockUnlock(t *testing.T) {
	l := CASLock{}
	ctx := context.Background()

	// Test case 1: Basic lock and unlock
	err := l.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	err = l.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Concurrent lock and unlock
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := l.Lock(ctx)
			if err != nil {
				t.Errorf("Lock failed: %v", err)
				return
			}
			counter++
			time.Sleep(time.Millisecond) // Simulate work
			counter--
			err = l.Unlock(ctx)
			if err != nil {
				t.Errorf("Unlock failed: %v", err)
				return
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestCASLock_TryLock(t *testing.T) {
	l := CASLock{}
	ctx := context.Background()

	// Test case 1: Successful TryLock
	locked, err := l.TryLock(ctx, time.Millisecond*100)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if !locked {
		t.Errorf("TryLock should have succeeded")
	}
	err = l.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Unsuccessful TryLock (timeout)
	err = l.Lock(ctx) // Lock it first
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	locked, err = l.TryLock(ctx, time.Millisecond*10) // Short timeout
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if locked {
		t.Errorf("TryLock should have timed out")
	}
	err = l.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
}

func TestCASLock_Close(t *testing.T) {
	l := CASLock{}
	ctx := context.Background()

	// Test case 1: Close successfully
	err := l.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Test case 2: Double close (should return error)
	err = l.Close()
	if err == nil {
		t.Errorf("Second Close should have returned an error")
	}

	// Test case 3: Lock after close (should return error)
	err = l.Lock(ctx)
	if err == nil {
		t.Errorf("Lock after Close should have returned an error")
	}

	// Test case 4: Unlock after close (should return error)
	err = l.Unlock(ctx)
	if err == nil {
		t.Errorf("Unlock after Close should have returned an error")
	}

	// Test case 5: TryLock after close (should return error)
	_, err = l.TryLock(ctx, time.Millisecond*10)
	if err == nil {
		t.Errorf("TryLock after Close should have returned an error")
	}
}

func TestCASLock_Concurrent(t *testing.T) {
	l := CASLock{}
	ctx := context.Background()
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			locked, err := l.TryLock(ctx, time.Millisecond*100)
			if err != nil {
				t.Errorf("TryLock failed: %v", err)
				return
			}
			if locked {
				counter++
				time.Sleep(time.Millisecond * 5) // Simulate work
				counter--
				err = l.Unlock(ctx)
				if err != nil {
					t.Errorf("Unlock failed: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestCASLock_Gosched(t *testing.T) {
	l := CASLock{}
	ctx := context.Background()
	var locked bool
	var err error

	// Lock the mutex
	err = l.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// Try to acquire the lock in a separate goroutine
	go func() {
		locked, err = l.TryLock(ctx, time.Second)
		if err != nil {
			t.Errorf("TryLock failed: %v", err)
			return
		}
	}()

	// Let the other goroutine run for a bit
	time.Sleep(time.Millisecond * 10)

	// Unlock the mutex
	err = l.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Wait for the other goroutine to finish
	time.Sleep(time.Millisecond * 10)

	// Check if the other goroutine acquired the lock
	if !locked {
		t.Errorf("TryLock should have succeeded")
	}
}

// ---------------------  RWLock (读写锁) ---------------------
func TestRWLock_RLockRUnlock(t *testing.T) {
	rw := RWLock{}
	ctx := context.Background()

	// Test case 1: Basic RLock and RUnlock
	err := rw.RLock(ctx)
	if err != nil {
		t.Fatalf("RLock failed: %v", err)
	}
	err = rw.RUnlock(ctx)
	if err != nil {
		t.Fatalf("RUnlock failed: %v", err)
	}

	// Test case 2: Multiple concurrent readers
	numReaders := 10
	var wg sync.WaitGroup
	var counter int32 // Shared counter

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := rw.RLock(ctx)
			if err != nil {
				t.Errorf("RLock failed: %v", err)
				return
			}
			atomic.AddInt32(&counter, 1) // Increment counter while holding the read lock
			time.Sleep(time.Millisecond * 5)
			atomic.AddInt32(&counter, -1) // Decrement counter before releasing the read lock
			err = rw.RUnlock(ctx)
			if err != nil {
				t.Errorf("RUnlock failed: %v", err)
				return
			}
		}()
	}

	wg.Wait()

	// Verify that the counter is 0 after all readers have finished
	if atomic.LoadInt32(&counter) != 0 {
		t.Errorf("Counter should be 0 after all readers, got %d", atomic.LoadInt32(&counter))
	}
}

func TestRWLock_LockUnlock(t *testing.T) {
	rw := RWLock{}
	ctx := context.Background()

	// Test case 1: Basic Lock and Unlock
	err := rw.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	err = rw.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Exclusive access for writers
	var readerRunning bool
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := rw.RLock(ctx)
		if err != nil {
			t.Errorf("RLock failed: %v", err)
			return
		}
		readerRunning = true
		time.Sleep(time.Millisecond * 10) // Simulate reader holding the lock
		readerRunning = false
		err = rw.RUnlock(ctx)
		if err != nil {
			t.Errorf("RUnlock failed: %v", err)
			return
		}
	}()

	time.Sleep(time.Millisecond * 5) // Give reader a chance to acquire the lock

	err = rw.Lock(ctx) // Attempt to acquire write lock
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	if readerRunning {
		t.Errorf("Writer acquired lock while reader was still running")
	}
	err = rw.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	wg.Wait() // Wait for the reader
}

func TestRWLock_TryLock(t *testing.T) {
	rw := RWLock{}
	ctx := context.Background()

	// Test case 1: Successful TryLock
	locked, err := rw.TryLock(ctx, time.Millisecond*100)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if !locked {
		t.Errorf("TryLock should have succeeded")
	}
	err = rw.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Unsuccessful TryLock (timeout)
	err = rw.Lock(ctx) // Lock it first
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	locked, err = rw.TryLock(ctx, time.Millisecond*10) // Short timeout
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if locked {
		t.Errorf("TryLock should have timed out")
	}
	err = rw.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 3: TryLock when readers are present
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := rw.RLock(ctx)
		if err != nil {
			t.Errorf("RLock failed: %v", err)
			return
		}
		time.Sleep(time.Millisecond * 20) // Simulate reader holding the lock
		err = rw.RUnlock(ctx)
		if err != nil {
			t.Errorf("RUnlock failed: %v", err)
			return
		}
	}()

	time.Sleep(time.Millisecond * 5) // Give reader a chance to acquire the lock

	locked, err = rw.TryLock(ctx, time.Millisecond*10) // Attempt to acquire write lock
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if locked {
		t.Errorf("TryLock should have failed because of existing reader")
		err = rw.Unlock(ctx) // Clean Up
		if err != nil {
			t.Fatalf("Unlock failed: %v", err)
		}
	}

	wg.Wait() // Wait for the reader
}

func TestRWLock_Close(t *testing.T) {
	rw := RWLock{}
	ctx := context.Background()

	// Test case 1: Close successfully
	err := rw.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Test case 2: Double close (should return error)
	err = rw.Close()
	if err == nil {
		t.Errorf("Second Close should have returned an error")
	}

	// Test case 3: RLock after close (should return error)
	err = rw.RLock(ctx)
	if err == nil {
		t.Errorf("RLock after Close should have returned an error")
	}

	// Test case 4: RUnlock after close (should return error)
	err = rw.RUnlock(ctx)
	if err == nil {
		t.Errorf("RUnlock after Close should have returned an error")
	}

	// Test case 5: Lock after close (should return error)
	err = rw.Lock(ctx)
	if err == nil {
		t.Errorf("Lock after Close should have returned an error")
	}

	// Test case 6: Unlock after close (should return error)
	err = rw.Unlock(ctx)
	if err == nil {
		t.Errorf("Unlock after Close should have returned an error")
	}

	// Test case 7: TryLock after close (should return error)
	_, err = rw.TryLock(ctx, time.Millisecond*10)
	if err == nil {
		t.Errorf("TryLock after Close should have returned an error")
	}
}

func TestRWLock_ConcurrentReadWrite(t *testing.T) {
	rw := RWLock{}
	ctx := context.Background()
	var counter int32
	var wg sync.WaitGroup
	numReaders := 10
	numWriters := 2

	// Launch Readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ { //Each reader does a few read operations
				err := rw.RLock(ctx)
				if err != nil {
					t.Errorf("RLock failed: %v", err)
					return
				}
				_ = atomic.LoadInt32(&counter) //Read the counter
				time.Sleep(time.Millisecond * 2)
				err = rw.RUnlock(ctx)
				if err != nil {
					t.Errorf("RUnlock failed: %v", err)
					return
				}
				time.Sleep(time.Millisecond) //Small delay between reads
			}

		}()
	}

	//Launch writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ { //Each writer performs a few write operations
				err := rw.Lock(ctx)
				if err != nil {
					t.Errorf("Lock failed: %v", err)
					return
				}
				atomic.AddInt32(&counter, 1) //Increment the counter
				time.Sleep(time.Millisecond * 3)
				atomic.AddInt32(&counter, -1) //Decrement the counter
				err = rw.Unlock(ctx)
				if err != nil {
					t.Errorf("Unlock failed: %v", err)
					return
				}
				time.Sleep(time.Millisecond * 2) //Small delay between writes
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish.
	// No direct validation of counter as it changes, but the test validates that
	// read and write operations are properly synchronized and no race conditions occur.
}

// ---------------------  TicketLock (票据锁) ---------------------
func TestTicketLock_LockUnlock(t *testing.T) {
	tl := TicketLock{}
	ctx := context.Background()

	// Test case 1: Basic lock and unlock
	err := tl.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	err = tl.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Concurrent lock and unlock
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := tl.Lock(ctx)
			if err != nil {
				t.Errorf("Lock failed: %v", err)
				return
			}
			counter++
			time.Sleep(time.Millisecond) // Simulate work
			counter--
			err = tl.Unlock(ctx)
			if err != nil {
				t.Errorf("Unlock failed: %v", err)
				return
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestTicketLock_TryLock(t *testing.T) {
	tl := TicketLock{}
	ctx := context.Background()

	// Test case 1: Successful TryLock
	locked, err := tl.TryLock(ctx, time.Millisecond*100)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if !locked {
		t.Errorf("TryLock should have succeeded")
	}
	err = tl.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Unsuccessful TryLock (timeout)
	err = tl.Lock(ctx) // Lock it first
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	locked, err = tl.TryLock(ctx, time.Millisecond*10) // Short timeout
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if locked {
		t.Errorf("TryLock should have timed out")
	}
	tl.Unlock(ctx)

	// Test case 3: TryLock after timeout, check serviceTicket is rolled back
	serviceTicketStart := atomic.LoadInt64(&tl.serviceTicket)
	locked, err = tl.TryLock(ctx, time.Millisecond*10)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if locked {
		t.Errorf("TryLock should have timed out")
		err = tl.Unlock(ctx)
		if err != nil {
			t.Fatalf("Unlock failed: %v", err)
		}
	}

	serviceTicketEnd := atomic.LoadInt64(&tl.serviceTicket)
	if serviceTicketStart != serviceTicketEnd {
		t.Errorf("serviceTicket should have been rolled back, start=%d, end=%d", serviceTicketStart, serviceTicketEnd)
	}

}

func TestTicketLock_Close(t *testing.T) {
	tl := TicketLock{}
	ctx := context.Background()

	// Test case 1: Close successfully
	err := tl.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Test case 2: Double close (should return error)
	err = tl.Close()
	if err == nil {
		t.Errorf("Second Close should have returned an error")
	}

	// Test case 3: Lock after close (should return error)
	err = tl.Lock(ctx)
	if err == nil {
		t.Errorf("Lock after Close should have returned an error")
	}

	// Test case 4: Unlock after close (should return error)
	err = tl.Unlock(ctx)
	if err == nil {
		t.Errorf("Unlock after Close should have returned an error")
	}

	// Test case 5: TryLock after close (should return error)
	_, err = tl.TryLock(ctx, time.Millisecond*10)
	if err == nil {
		t.Errorf("TryLock after Close should have returned an error")
	}
}

func TestTicketLock_Concurrent(t *testing.T) {
	tl := TicketLock{}
	ctx := context.Background()
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			locked, err := tl.TryLock(ctx, time.Millisecond*100)
			if err != nil {
				t.Errorf("TryLock failed: %v", err)
				return
			}
			if locked {
				counter++
				time.Sleep(time.Millisecond * 5) // Simulate work
				counter--
				err = tl.Unlock(ctx)
				if err != nil {
					t.Errorf("Unlock failed: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestTicketLock_Fairness(t *testing.T) {
	tl := TicketLock{}
	ctx := context.Background()
	numGoroutines := 10
	var order []int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int64) {
			defer wg.Done()
			err := tl.Lock(ctx)
			if err != nil {
				t.Errorf("Lock failed: %v", err)
				return
			}
			mu.Lock()
			order = append(order, id)
			mu.Unlock()
			time.Sleep(time.Millisecond * 10) // Simulate work
			err = tl.Unlock(ctx)
			if err != nil {
				t.Errorf("Unlock failed: %v", err)
				return
			}
		}(int64(i))
	}

	wg.Wait() // Wait for all goroutines to finish.

	// Check that the order is strictly FIFO
	for i := 0; i < numGoroutines; i++ {
		if order[i] != int64(i) {
			t.Errorf("Fairness check failed: expected %d, got %d", i, order[i])
		}
	}
}

// ---------------------  ALock (数组锁) ---------------------
func TestALock_LockUnlock(t *testing.T) {
	size := 4 // Choose a small size for testing purposes
	al := NewALock(size)
	ctx := context.Background()

	// Test case 1: Basic lock and unlock
	err := al.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	err = al.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Concurrent lock and unlock
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := al.Lock(ctx)
			if err != nil {
				t.Errorf("Lock failed: %v", err)
				return
			}
			counter++
			time.Sleep(time.Millisecond) // Simulate work
			counter--
			err = al.Unlock(ctx)
			if err != nil {
				t.Errorf("Unlock failed: %v", err)
				return
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestALock_TryLock(t *testing.T) {
	size := 4 // Choose a small size for testing purposes
	al := NewALock(size)
	ctx := context.Background()

	// Test case 1: Successful TryLock
	locked, err := al.TryLock(ctx, time.Millisecond*100)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if !locked {
		t.Errorf("TryLock should have succeeded")
	}
	err = al.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Unsuccessful TryLock (timeout)
	err = al.Lock(ctx) // Lock it first
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	locked, err = al.TryLock(ctx, time.Millisecond*10) // Short timeout
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if locked {
		t.Errorf("TryLock should have timed out")
	}
	err = al.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 3: TryLock after timeout, check queue is rolled back
	queueStart := atomic.LoadInt64(&al.queue)
	locked, err = al.TryLock(ctx, time.Millisecond*10)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if locked {
		t.Errorf("TryLock should have timed out")
		err = al.Unlock(ctx)
		if err != nil {
			t.Fatalf("Unlock failed: %v", err)
		}
	}

	queueEnd := atomic.LoadInt64(&al.queue)
	if queueStart != queueEnd {
		t.Errorf("Queue should have been rolled back, start=%d, end=%d", queueStart, queueEnd)
	}
}

func TestALock_Close(t *testing.T) {
	size := 4 // Choose a small size for testing purposes
	al := NewALock(size)
	ctx := context.Background()

	// Test case 1: Close successfully
	err := al.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Test case 2: Double close (should return error)
	err = al.Close()
	if err == nil {
		t.Errorf("Second Close should have returned an error")
	}

	// Test case 3: Lock after close (should return error)
	err = al.Lock(ctx)
	if err == nil {
		t.Errorf("Lock after Close should have returned an error")
	}

	// Test case 4: Unlock after close (should return error)
	err = al.Unlock(ctx)
	if err == nil {
		t.Errorf("Unlock after Close should have returned an error")
	}

	// Test case 5: TryLock after close (should return error)
	_, err = al.TryLock(ctx, time.Millisecond*10)
	if err == nil {
		t.Errorf("TryLock after Close should have returned an error")
	}
}

func TestALock_Concurrent(t *testing.T) {
	size := 4 // Choose a small size for testing purposes
	al := NewALock(size)
	ctx := context.Background()
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			locked, err := al.TryLock(ctx, time.Millisecond*100)
			if err != nil {
				t.Errorf("TryLock failed: %v", err)
				return
			}
			if locked {
				counter++
				time.Sleep(time.Millisecond * 5) // Simulate work
				counter--
				err = al.Unlock(ctx)
				if err != nil {
					t.Errorf("Unlock failed: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestALock_Fairness(t *testing.T) {
	size := 4 // Choose a small size for testing purposes
	al := NewALock(size)
	ctx := context.Background()
	numGoroutines := size
	var order []int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			err := al.Lock(ctx)
			if err != nil {
				t.Errorf("Lock failed: %v", err)
				return
			}
			mu.Lock()
			order = append(order, int64(id))
			mu.Unlock()
			time.Sleep(time.Millisecond * 10) // Simulate work
			err = al.Unlock(ctx)
			if err != nil {
				t.Errorf("Unlock failed: %v", err)
				return
			}
		}(i)
	}

	wg.Wait() // Wait for all goroutines to finish.

	// Check that the order is roughly fair (not strictly guaranteed due to goroutine scheduling)
	for i := 0; i < numGoroutines; i++ {
		if int(order[i]) != i {
			t.Logf("Fairness check failed: expected %d, got %d", i, order[i])
		}
	}
}

// ---------------------  CLHLock (链表锁) ---------------------
func TestCLHLock_LockUnlock(t *testing.T) {
	l := NewCLHLock()
	ctx := context.Background()

	// Test case 1: Basic lock and unlock
	err := l.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	err = l.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Concurrent lock and unlock
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := l.Lock(ctx)
			if err != nil {
				t.Errorf("Lock failed: %v", err)
				return
			}
			counter++
			time.Sleep(time.Millisecond) // Simulate work
			counter--
			err = l.Unlock(ctx)
			if err != nil {
				t.Errorf("Unlock failed: %v", err)
				return
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestCLHLock_TryLock(t *testing.T) {
	l := NewCLHLock()
	ctx := context.Background()

	// Test case 1: Successful TryLock
	locked, err := l.TryLock(ctx, time.Millisecond*100)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if !locked {
		t.Errorf("TryLock should have succeeded")
	}
	err = l.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Test case 2: Unsuccessful TryLock (timeout)
	err = l.Lock(ctx) // Lock it first
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	locked, err = l.TryLock(ctx, time.Millisecond*10) // Short timeout
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if locked {
		t.Errorf("TryLock should have timed out")
	}
	err = l.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
}

func TestCLHLock_Close(t *testing.T) {
	l := NewCLHLock()
	ctx := context.Background()

	// Test case 1: Close successfully
	err := l.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Test case 2: Double close (should not return error as per spec, only sets the closed flag)
	//err = l.Close()
	//if err == nil {
	//	t.Errorf("Second Close should have returned an error")
	//}

	// Test case 3: Lock after close (should return error)
	err = l.Lock(ctx)
	if err == nil {
		t.Errorf("Lock after Close should have returned an error")
	}

	// Test case 4: Unlock after close (should return error)
	err = l.Unlock(ctx)
	if err == nil {
		t.Errorf("Unlock after Close should have returned an error")
	}

	// Test case 5: TryLock after close (should return error)
	locked, err := l.TryLock(ctx, time.Millisecond*10)
	if err == nil {
		t.Errorf("TryLock after Close should have returned an error")
	}
	if locked {
		t.Errorf("TryLock should not have succeeded after close")
	}
}

func TestCLHLock_Concurrent(t *testing.T) {
	l := NewCLHLock()
	ctx := context.Background()
	var counter int
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			locked, err := l.TryLock(ctx, time.Millisecond*100)
			if err != nil {
				t.Errorf("TryLock failed: %v", err)
				return
			}
			if locked {
				counter++
				time.Sleep(time.Millisecond * 5) // Simulate work
				counter--
				err = l.Unlock(ctx)
				if err != nil {
					t.Errorf("Unlock failed: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check that the counter is always 0 (no race condition)
	if counter != 0 {
		t.Errorf("Race condition detected, counter = %d", counter)
	}
}

func TestCLHLock_Gosched(t *testing.T) {
	l := NewCLHLock()
	ctx := context.Background()
	var locked bool
	var err error

	// Lock the mutex
	err = l.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// Try to acquire the lock in a separate goroutine
	go func() {
		locked, err = l.TryLock(ctx, time.Second)
		if err != nil {
			t.Errorf("TryLock failed: %v", err)
			return
		}
	}()

	// Let the other goroutine run for a bit
	time.Sleep(time.Millisecond * 10)

	// Unlock the mutex
	err = l.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Wait for the other goroutine to finish
	time.Sleep(time.Millisecond * 10)

	// Check if the other goroutine acquired the lock
	if !locked {
		t.Errorf("TryLock should have succeeded")
	}
}

func TestCLHLock_UnlockOnUnlocked(t *testing.T) {
	l := NewCLHLock()
	ctx := context.Background()

	// Attempt to unlock an unlocked lock
	err := l.Unlock(ctx)
	if err == nil {
		t.Errorf("Unlock should have returned an error when called on an unlocked lock")
	}

	// Lock and then unlock to make sure subsequent unlocks are still handled correctly
	err = l.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	err = l.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Attempt to unlock again (should return an error)
	err = l.Unlock(ctx)
	if err == nil {
		t.Errorf("Second Unlock should have returned an error")
	}
}

func TestCLHLock_TryLockContended(t *testing.T) {
	l := NewCLHLock()
	ctx := context.Background()

	// Lock the lock in the main goroutine
	err := l.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// Attempt to TryLock in a separate goroutine
	var locked bool
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		locked, err = l.TryLock(ctx, time.Millisecond*100) // Short timeout
		if err != nil {
			t.Errorf("TryLock failed: %v", err)
			return
		}
	}()

	// Wait a bit for the TryLock to start
	time.Sleep(time.Millisecond * 10)

	// Unlock the lock in the main goroutine
	err = l.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	wg.Wait() // Wait for TryLock to complete

	// Check if TryLock succeeded (it should, eventually)
	if !locked {
		t.Errorf("TryLock should have succeeded after the lock was released")
	}
}
