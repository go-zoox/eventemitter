package eventemitter

import (
	"sync"
	"testing"
)

func TestEventEmitter(t *testing.T) {
	e := New()
	count := 0
	e.On("test", func(payload any) {
		count++
		t.Log("test", payload)
	})

	e.Start()

	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		index := i
		wg.Add(1)
		go func() {
			e.Emit("test", index)
			wg.Done()
		}()
	}

	wg.Wait()

	if count != 10 {
		t.Error("count should be 10")
	}
}
