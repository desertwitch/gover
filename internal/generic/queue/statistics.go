package queue

import (
	"log/slog"
	"sync"
	"time"
)

type TransferInfo struct {
	sync.RWMutex
	Error            error
	Started          bool
	Finished         bool
	Percentage       float64
	TimeRemaining    time.Duration
	EstimatedFinish  time.Time
	StartTime        time.Time
	EndTime          time.Time
	BytesTotal       uint64
	BytesTransferred uint64
	TransferRate     float64
}

func (t *TransferInfo) Start(bytesTotal uint64) {
	t.Lock()
	defer t.Unlock()

	now := time.Now()

	t.Started = true
	t.StartTime = now
	t.BytesTotal = bytesTotal
	t.BytesTransferred = 0
	t.Percentage = 0
	t.TransferRate = 0

	t.EstimatedFinish = now.Add(1 * time.Second)
	t.TimeRemaining = 1 * time.Second
}

func (t *TransferInfo) End() {
	t.Lock()
	defer t.Unlock()

	t.Finished = true
	t.EndTime = time.Now()

	t.BytesTransferred = t.BytesTotal
	t.Percentage = 100.0
	t.TimeRemaining = 0
	t.EstimatedFinish = t.EndTime
}

func (t *TransferInfo) IsStarted() bool {
	t.RLock()
	defer t.RUnlock()

	return t.Started
}

func (t *TransferInfo) IsDone() bool {
	t.RLock()
	defer t.RUnlock()

	return t.Finished
}

func (t *TransferInfo) IsError() bool {
	t.RLock()
	defer t.RUnlock()

	return t.Error != nil
}

func (t *TransferInfo) SetError(err error) {
	t.Lock()
	defer t.Unlock()

	t.Error = err
}

func (t *TransferInfo) Update(totalBytesRead uint64) {
	t.Lock()
	defer t.Unlock()

	now := time.Now()
	elapsed := now.Sub(t.StartTime)

	additionalBytes := totalBytesRead - t.BytesTransferred
	t.BytesTransferred = totalBytesRead

	if elapsed < time.Second {
		return
	}

	if t.BytesTotal > 0 {
		t.Percentage = float64(t.BytesTransferred) / float64(t.BytesTotal) * 100 //nolint:mnd
	}

	instantRate := float64(t.BytesTransferred) / elapsed.Seconds()

	if t.TransferRate == 0 {
		t.TransferRate = instantRate
	} else {
		t.TransferRate = 0.7*t.TransferRate + 0.3*instantRate //nolint:mnd
	}

	slog.Debug("Transfer calculation",
		"instantRate_MBps", instantRate/1024/1024,
		"weightedRate_MBps", t.TransferRate/1024/1024,
		"bytesTransferred", t.BytesTransferred,
		"newBytes", additionalBytes,
		"elapsed_sec", elapsed.Seconds())

	if t.TransferRate > 0 && t.BytesTransferred < t.BytesTotal {
		bytesRemaining := t.BytesTotal - t.BytesTransferred
		secondsRemaining := float64(bytesRemaining) / t.TransferRate
		t.TimeRemaining = time.Duration(secondsRemaining) * time.Second
		t.EstimatedFinish = now.Add(t.TimeRemaining)
	}
}

func (t *TransferInfo) GetStats() (float64, time.Time, time.Duration, uint64, uint64, float64) {
	t.RLock()
	defer t.RUnlock()

	return t.Percentage, t.EstimatedFinish, t.TimeRemaining, t.BytesTransferred, t.BytesTotal, t.TransferRate
}
