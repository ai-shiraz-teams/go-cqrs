package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type BusMetrics interface {
	RecordExecution(commandName string, duration time.Duration)

	RecordFailure(commandName string, duration time.Duration)

	GetStats(commandName string) Stats

	GetAllStats() map[string]Stats

	Reset()
}

type Stats struct {
	TotalExecutions uint64

	SuccessfulExecutions uint64

	FailedExecutions uint64

	AverageDuration time.Duration

	MinDuration time.Duration

	MaxDuration time.Duration

	LastExecution time.Time
}

func (s Stats) SuccessRate() float64 {
	if s.TotalExecutions == 0 {
		return 0
	}
	return float64(s.SuccessfulExecutions) / float64(s.TotalExecutions) * 100
}

func (s Stats) FailureRate() float64 {
	if s.TotalExecutions == 0 {
		return 0
	}
	return float64(s.FailedExecutions) / float64(s.TotalExecutions) * 100
}

type DefaultMetrics struct {
	mu    sync.RWMutex
	stats map[string]*commandStats
}

type commandStats struct {
	totalExecutions      uint64
	successfulExecutions uint64
	failedExecutions     uint64
	totalDuration        int64
	minDuration          int64
	maxDuration          int64
	lastExecution        int64
}

func NewDefaultMetrics() *DefaultMetrics {
	return &DefaultMetrics{
		stats: make(map[string]*commandStats),
	}
}

func (m *DefaultMetrics) RecordExecution(commandName string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := m.getOrCreateStats(commandName)

	durationNs := duration.Nanoseconds()
	now := time.Now().UnixNano()

	atomic.AddUint64(&stats.totalExecutions, 1)
	atomic.AddUint64(&stats.successfulExecutions, 1)
	atomic.AddInt64(&stats.totalDuration, durationNs)
	atomic.StoreInt64(&stats.lastExecution, now)

	for {
		currentMin := atomic.LoadInt64(&stats.minDuration)
		if currentMin == 0 || durationNs < currentMin {
			if atomic.CompareAndSwapInt64(&stats.minDuration, currentMin, durationNs) {
				break
			}
		} else {
			break
		}
	}

	for {
		currentMax := atomic.LoadInt64(&stats.maxDuration)
		if durationNs > currentMax {
			if atomic.CompareAndSwapInt64(&stats.maxDuration, currentMax, durationNs) {
				break
			}
		} else {
			break
		}
	}
}

func (m *DefaultMetrics) RecordFailure(commandName string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := m.getOrCreateStats(commandName)

	durationNs := duration.Nanoseconds()
	now := time.Now().UnixNano()

	atomic.AddUint64(&stats.totalExecutions, 1)
	atomic.AddUint64(&stats.failedExecutions, 1)
	atomic.AddInt64(&stats.totalDuration, durationNs)
	atomic.StoreInt64(&stats.lastExecution, now)

	for {
		currentMin := atomic.LoadInt64(&stats.minDuration)
		if currentMin == 0 || durationNs < currentMin {
			if atomic.CompareAndSwapInt64(&stats.minDuration, currentMin, durationNs) {
				break
			}
		} else {
			break
		}
	}

	for {
		currentMax := atomic.LoadInt64(&stats.maxDuration)
		if durationNs > currentMax {
			if atomic.CompareAndSwapInt64(&stats.maxDuration, currentMax, durationNs) {
				break
			}
		} else {
			break
		}
	}
}

func (m *DefaultMetrics) GetStats(commandName string) Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats, exists := m.stats[commandName]
	if !exists {
		return Stats{}
	}

	return m.convertToStats(stats)
}

func (m *DefaultMetrics) GetAllStats() map[string]Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Stats, len(m.stats))
	for commandName, stats := range m.stats {
		result[commandName] = m.convertToStats(stats)
	}

	return result
}

func (m *DefaultMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats = make(map[string]*commandStats)
}

func (m *DefaultMetrics) getOrCreateStats(commandName string) *commandStats {
	stats, exists := m.stats[commandName]
	if !exists {
		stats = &commandStats{}
		m.stats[commandName] = stats
	}
	return stats
}

func (m *DefaultMetrics) convertToStats(stats *commandStats) Stats {
	totalExecutions := atomic.LoadUint64(&stats.totalExecutions)
	successfulExecutions := atomic.LoadUint64(&stats.successfulExecutions)
	failedExecutions := atomic.LoadUint64(&stats.failedExecutions)
	totalDuration := atomic.LoadInt64(&stats.totalDuration)
	minDuration := atomic.LoadInt64(&stats.minDuration)
	maxDuration := atomic.LoadInt64(&stats.maxDuration)
	lastExecution := atomic.LoadInt64(&stats.lastExecution)

	var avgDuration time.Duration
	if totalExecutions > 0 {
		avgDuration = time.Duration(totalDuration / int64(totalExecutions))
	}

	return Stats{
		TotalExecutions:      totalExecutions,
		SuccessfulExecutions: successfulExecutions,
		FailedExecutions:     failedExecutions,
		AverageDuration:      avgDuration,
		MinDuration:          time.Duration(minDuration),
		MaxDuration:          time.Duration(maxDuration),
		LastExecution:        time.Unix(0, lastExecution),
	}
}

type NoOpMetrics struct{}

func NewNoOpMetrics() *NoOpMetrics {
	return &NoOpMetrics{}
}

func (m *NoOpMetrics) RecordExecution(commandName string, duration time.Duration) {}

func (m *NoOpMetrics) RecordFailure(commandName string, duration time.Duration) {}

func (m *NoOpMetrics) GetStats(commandName string) Stats {
	return Stats{}
}

func (m *NoOpMetrics) GetAllStats() map[string]Stats {
	return make(map[string]Stats)
}

func (m *NoOpMetrics) Reset() {}
