package mempool

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/33cn/chain33/queue"
)

func TestStep(t *testing.T) {
	done := make(chan struct{})
	in := make(chan queue.Message)
	msg := queue.Message{Id: 0}
	cb := func(in queue.Message) queue.Message {
		in.Id++
		time.Sleep(time.Microsecond)
		return in
	}
	out := step(done, in, cb)
	in <- msg
	msg2 := <-out
	assert.Equal(t, msg2.Id, int64(1))
	close(done)
}

func TestMutiStep(t *testing.T) {
	done := make(chan struct{})
	in := make(chan queue.Message)
	msg := queue.Message{Id: 0}
	step1 := func(in queue.Message) queue.Message {
		in.Id++
		time.Sleep(time.Microsecond)
		return in
	}
	out1 := step(done, in, step1)
	step2 := func(in queue.Message) queue.Message {
		in.Id++
		time.Sleep(time.Microsecond)
		return in
	}
	out21 := step(done, out1, step2)
	out22 := step(done, out1, step2)

	out3 := mergeList(done, out21, out22)
	in <- msg
	msg2 := <-out3
	assert.Equal(t, msg2.Id, int64(2))
	close(done)
}

func BenchmarkStep(b *testing.B) {
	done := make(chan struct{})
	in := make(chan queue.Message)
	msg := queue.Message{Id: 0}
	cb := func(in queue.Message) queue.Message {
		in.Id++
		time.Sleep(100 * time.Microsecond)
		return in
	}
	out := step(done, in, cb)
	go func() {
		for i := 0; i < b.N; i++ {
			in <- msg
		}
	}()
	for i := 0; i < b.N; i++ {
		msg2 := <-out
		assert.Equal(b, msg2.Id, int64(1))
	}
	close(done)
}

func BenchmarkStepMerge(b *testing.B) {
	done := make(chan struct{})
	in := make(chan queue.Message)
	msg := queue.Message{Id: 0}
	cb := func(in queue.Message) queue.Message {
		in.Id++
		time.Sleep(100 * time.Microsecond)
		return in
	}
	chs := make([]<-chan queue.Message, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		chs[i] = step(done, in, cb)
	}
	out := merge(done, chs)
	go func() {
		for i := 0; i < b.N; i++ {
			in <- msg
		}
	}()
	for i := 0; i < b.N; i++ {
		msg2 := <-out
		assert.Equal(b, msg2.Id, int64(1))
	}
	close(done)
}
