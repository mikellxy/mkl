package breaker

import (
	"fmt"
	"testing"
	"time"
)

func TestBreaker(t *testing.T) {
	conf := Config{CountDuration: 600 * time.Millisecond, F: 1.5}
	brk := NewBreaker(conf)
	win := brk.counter.(*window)
	brk.Succeed()
	brk.Succeed()
	brk.Succeed()
	brk.Failed()
	time.Sleep(200 * time.Millisecond)
	brk.Succeed()
	brk.Succeed()
	brk.Failed()
	brk.Failed()
	fmt.Println(win.buckets)
	time.Sleep(600 * time.Millisecond)
	fmt.Println(win.getCount())
	brk.Succeed()
	brk.Failed()
	fmt.Println(win.buckets)
	fmt.Println(win.getCount())

	fmt.Println(StatusOpen)
	fmt.Println(StatusHalfOpen)
	fmt.Println(StatusClosed)
}