package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestNewSnowflake(t *testing.T) {
	// Test case 1: Default options
	options := Options{}
	sf, err := NewSnowflake(options)
	if err != nil {
		t.Fatalf("NewSnowflake failed: %v", err)
	}
	if sf.epoch != time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli() {
		t.Errorf("Epoch not initialized correctly, got %d", sf.epoch)
	}
	if sf.datacenterBits != 5 {
		t.Errorf("DatacenterBits not initialized correctly, got %d", sf.datacenterBits)
	}
	if sf.workerBits != 5 {
		t.Errorf("WorkerBits not initialized correctly, got %d", sf.workerBits)
	}
	if sf.sequenceBits != 12 {
		t.Errorf("SequenceBits not initialized correctly, got %d", sf.sequenceBits)
	}
	if sf.maxWorkerID != 31 {
		t.Errorf("maxWorkerID not initialized correctly, got %d", sf.maxWorkerID)
	}
	if sf.maxDatacenterID != 31 {
		t.Errorf("maxDatacenterID not initialized correctly, got %d", sf.maxDatacenterID)
	}

	// Test case 2: Custom options
	options = Options{
		Epoch:          1609459200000, // 2021-01-01
		WorkerID:       10,
		DatacenterID:   5,
		TimestampBits:  41,
		DatacenterBits: 6,
		WorkerBits:     6,
		SequenceBits:   10,
	}
	sf, err = NewSnowflake(options)
	if err != nil {
		t.Fatalf("NewSnowflake failed: %v", err)
	}
	if sf.epoch != 1609459200000 {
		t.Errorf("Epoch not initialized correctly, got %d", sf.epoch)
	}
	if sf.workerID != 10 {
		t.Errorf("WorkerID not initialized correctly, got %d", sf.workerID)
	}
	if sf.datacenterID != 5 {
		t.Errorf("DatacenterID not initialized correctly, got %d", sf.datacenterID)
	}
	if sf.sequenceMask != 1023 {
		t.Errorf("sequenceMask not initialized correctly, got %d", sf.sequenceMask)
	}
	if sf.datacenterBits != 6 {
		t.Errorf("DatacenterBits not initialized correctly, got %d", sf.datacenterBits)
	}
	if sf.workerBits != 6 {
		t.Errorf("WorkerBits not initialized correctly, got %d", sf.workerBits)
	}
	if sf.sequenceBits != 10 {
		t.Errorf("SequenceBits not initialized correctly, got %d", sf.sequenceBits)
	}
	if sf.maxWorkerID != 63 {
		t.Errorf("maxWorkerID not initialized correctly, got %d", sf.maxWorkerID)
	}
	if sf.maxDatacenterID != 63 {
		t.Errorf("maxDatacenterID not initialized correctly, got %d", sf.maxDatacenterID)
	}

	// Test case 3: Invalid WorkerID
	options = Options{WorkerID: 64, DatacenterBits: 5, WorkerBits: 5, SequenceBits: 12}
	_, err = NewSnowflake(options)
	if err == nil {
		t.Errorf("Expected error for invalid WorkerID")
	}

	// Test case 4: Invalid DatacenterID
	options = Options{WorkerID: 10, DatacenterID: 32, DatacenterBits: 5, WorkerBits: 5, SequenceBits: 12}
	_, err = NewSnowflake(options)
	if err == nil {
		t.Errorf("Expected error for invalid DatacenterID")
	}
}

func TestSnowflake_GenerateID(t *testing.T) {
	options := Options{
		Epoch:          1609459200000, // 2021-01-01
		WorkerID:       1,
		DatacenterID:   1,
		TimestampBits:  41,
		DatacenterBits: 5,
		WorkerBits:     5,
		SequenceBits:   12,
	}
	sf, err := NewSnowflake(options)
	if err != nil {
		t.Fatalf("NewSnowflake failed: %v", err)
	}

	// Test case 1: Generate a single ID
	id1, err := sf.GenerateID()
	if err != nil {
		t.Fatalf("GenerateID failed: %v", err)
	}
	if id1 == 0 {
		t.Errorf("Generated ID is zero")
	}

	// Test case 2: Generate multiple IDs
	idSet := make(map[int64]bool)
	numIDs := 1000
	for i := 0; i < numIDs; i++ {
		id, err := sf.GenerateID()
		if err != nil {
			t.Fatalf("GenerateID failed: %v", err)
		}
		if _, ok := idSet[id]; ok {
			t.Errorf("Duplicate ID generated: %d", id)
		}
		idSet[id] = true
	}
	if len(idSet) != numIDs {
		t.Errorf("Generated %d unique IDs, expected %d", len(idSet), numIDs)
	}

	// Test case 3: Ensure IDs are increasing
	lastID := int64(0)
	for i := 0; i < 100; i++ {
		id, err := sf.GenerateID()
		if err != nil {
			t.Fatalf("GenerateID failed: %v", err)
		}
		if id <= lastID {
			t.Errorf("IDs are not increasing: current %d, last %d", id, lastID)
		}
		lastID = id
	}

	// Test case 4: Sequence rollover within the same millisecond
	sf.lastTimestamp = time.Now().UnixMilli()
	sf.sequence = sf.sequenceMask // Force sequence to max value
	id, err := sf.GenerateID()
	if err != nil {
		t.Fatalf("GenerateID failed: %v", err)
	}
	if sf.lastTimestamp <= time.Now().UnixMilli() {
		// Sequence rolled over, and time advanced.
		// This is expected.
	}

	if id == 0 {
		t.Errorf("Generated ID is zero after sequence rollover")
	}

	// Test case 5: Clock moves backwards (simulated)
	sf.lastTimestamp = time.Now().UnixMilli()
	time.Sleep(2 * time.Millisecond) // Ensure some time has passed
	currentTime := time.Now().UnixMilli()
	sf.lastTimestamp = currentTime + 10

	// Simulate clock moving backwards
	time.Sleep(5 * time.Millisecond)
	_, err = sf.GenerateID()
	if err == nil {
		t.Errorf("Expected error when clock moves backwards")
	} else {
		fmt.Println(err)
	}
}

func TestSnowflake_handleClockBackwards(t *testing.T) {
	options := Options{
		Epoch:          1609459200000, // 2021-01-01
		WorkerID:       1,
		DatacenterID:   1,
		TimestampBits:  41,
		DatacenterBits: 5,
		WorkerBits:     5,
		SequenceBits:   12,
	}
	sf, err := NewSnowflake(options)
	if err != nil {
		t.Fatalf("NewSnowflake failed: %v", err)
	}

	// Set lastTimestamp to a future time
	sf.lastTimestamp = time.Now().UnixMilli() + 10

	// Simulate a small clock rollback (less than 5ms)
	rollbackTime := time.Now().UnixMilli() + 6

	startTime := time.Now()
	correctedTime := sf.handleClockBackwards(rollbackTime)
	duration := time.Since(startTime)

	if correctedTime <= rollbackTime {
		t.Errorf("handleClockBackwards should return a time greater than rollbackTime")
	}

	// Check that the waiting time was at least 1ms (to avoid CPU spinning)
	if duration < time.Millisecond {
		t.Errorf("handleClockBackwards should wait at least 1ms, waited %v", duration)
	}

	// Simulate a large clock rollback (greater than 5ms)
	sf.lastTimestamp = time.Now().UnixMilli() + 10
	rollbackTime = time.Now().UnixMilli()

	// Expect a panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("handleClockBackwards should panic for large clock rollback")
		}
	}()
	sf.handleClockBackwards(rollbackTime)
}

func TestSnowflake_waitUntilNextMillis(t *testing.T) {
	options := Options{
		Epoch:          1609459200000, // 2021-01-01
		WorkerID:       1,
		DatacenterID:   1,
		TimestampBits:  41,
		DatacenterBits: 5,
		WorkerBits:     5,
		SequenceBits:   12,
	}
	sf, err := NewSnowflake(options)
	if err != nil {
		t.Fatalf("NewSnowflake failed: %v", err)
	}

	lastTimestamp := time.Now().UnixMilli()

	startTime := time.Now()
	nextMillis := sf.waitUntilNextMillis(lastTimestamp)
	duration := time.Since(startTime)

	if nextMillis <= lastTimestamp {
		t.Errorf("waitUntilNextMillis should return a time greater than lastTimestamp")
	}

	// Check that the waiting time was at least 1ms (to avoid CPU spinning)
	if duration < time.Millisecond {
		t.Errorf("waitUntilNextMillis should wait at least 1ms, waited %v", duration)
	}
}
