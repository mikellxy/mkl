package breaker

import (
	"math/rand"
	"time"
)

const (
	StatusClosed = iota
	StatusHalfOpen
	StatusOpen
)

type counter interface {
	addSucceed()
	addFailed()
	getCount() (float64, float64)
	setDuration(dur time.Duration)
}

type window struct {
	buckets        [][]float64
	size           int           // 窗口内令牌桶的数量
	duration       time.Duration // 滑动窗口的时间长度
	bucketDuration time.Duration // 每个令牌桶的时间长度
	lastAddTime    time.Time     // 上次操作令牌桶的时间
	bucketOffset   int           // 上次操作的令牌桶在令牌桶数组中的下标
}

func newWindow() *window {
	return &window{
		bucketDuration: 100 * time.Millisecond,
		lastAddTime:    time.Now(),
		bucketOffset:   0,
	}
}

func (w *window) setDuration(dur time.Duration) {
	// 一个桶对应100ms
	q, r := dur/(100*time.Millisecond), dur%(100*time.Millisecond)
	if q < 1 || r != 0 {
		panic("breaker's count duration must be positive integer multiple of 100 millisecond")
	}
	w.size = int(q)
	w.buckets = make([][]float64, w.size)
	for i := range w.buckets {
		w.buckets[i] = []float64{0, 0}
	}
	w.duration = dur
}

func (w *window) getBucketOffset() int {
	span := w.bucketSpan()
	w.lastAddTime = w.lastAddTime.Add(time.Duration(span) * w.bucketDuration)
	if span > w.size {
		// span大于w.size时把span设置为w.size，可以保证把所有令牌桶重置
		span = w.size
	}
	end := w.bucketOffset + span
	e1, e2 := end, -1
	// 这次和前一次操作的两个令牌桶之间的令牌桶需要重置，应为中间这段时间没有请求经过熔断器
	if end >= w.size {
		e1, e2 = w.size-1, end%w.size
	}
	offset := w.bucketOffset
	for i := w.bucketOffset + 1; i <= e1; i++ {
		w.reset(i)
		offset = i
	}
	for i := 0; i <= e2; i++ {
		w.reset(i)
		offset = i
	}
	w.bucketOffset = offset
	return offset
}

func (w *window) bucketSpan() int {
	// 当前与上次插入令牌的时间间隔对应几个令牌桶
	return int(time.Since(w.lastAddTime) / w.bucketDuration)
}

func (w *window) reset(offset int) {
	w.buckets[offset][0] = 0
	w.buckets[offset][1] = 0
}

func (w *window) addSucceed() {
	offset := w.getBucketOffset()
	w.buckets[offset][0] += 1
}

func (w *window) addFailed() {
	offset := w.getBucketOffset()
	w.buckets[offset][1] += 1
}

func (w *window) getCount() (float64, float64) {
	var succeed, failed float64
	span := w.bucketSpan()
	if span > w.size {
		span = w.size
	}
	numBuckets := w.size - span
	offset := (w.bucketOffset + span + 1) % w.size
	for numBuckets > 0 {
		succeed += w.buckets[offset][0]
		failed += w.buckets[offset][1]
		offset = (offset + 1) % w.size
		numBuckets--
	}
	return succeed, failed
}

type breaker struct {
	countDuration time.Duration
	counter       counter
	f             float64
	status        int
	halfOpenCount int
}

func (b *breaker) Succeed() {
	b.counter.addSucceed()
	if b.status == StatusHalfOpen {
		b.halfOpenCount++
		if b.halfOpenCount >= 10 {
			b.status = StatusClosed
		}
	}
}

func (b *breaker) Failed() {
	b.counter.addFailed()
	if b.status == StatusHalfOpen {
		go b.OpenCircuit()
	}
}

func (b *breaker) Allow() bool {
	if b.status == StatusOpen {
		return false
	}
	succeed, failed := b.counter.getCount()
	if succeed*b.f < succeed+failed {
		go b.OpenCircuit()
		return false
	}
	if b.status == StatusHalfOpen {
		if rand.Float64() > 1/b.f {
			return false
		}
	}
	return true
}

func (b *breaker) OpenCircuit() {
	b.status = StatusOpen
	t := time.NewTimer(500 * time.Millisecond)
	select {
	case <-t.C:
		b.status = StatusHalfOpen
		b.halfOpenCount = 0
		rand.Seed(time.Now().Unix())
	}
}

type Config struct {
	CountDuration time.Duration
	F             float64
}

func NewBreaker(conf Config) *breaker {
	win := newWindow()
	win.setDuration(conf.CountDuration)
	return &breaker{
		countDuration: conf.CountDuration,
		counter:       win,
		f:             conf.F,
		status:        StatusClosed,
	}
}
