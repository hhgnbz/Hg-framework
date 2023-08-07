package hintmq

import (
	"sync"
	"testing"
	"unsafe"
)

var ch = make(chan string)

func intToBytes(i int) []byte {
	gc := unsafe.Pointer(&i)
	start := uintptr(gc)
	sizeOfByte := unsafe.Sizeof(byte(' '))

	result := make([]byte, 4)
	for index := 0; index < 4; index++ {
		b := *(*byte)(unsafe.Pointer(start + sizeOfByte*uintptr(index)))
		result[index] = b
	}

	return result
}

func TestQueue(t *testing.T) {
	t.Run("test queue", func(t *testing.T) {
		ch := NewHintMQ()
		for i := 0; i < 10086; i++ {
			ch.Write(intToBytes(i))
		}

		index := 0
		for i := 0; i < 10090; i++ {
			select {
			case data := <-ch.Read():
				want := intToBytes(i)
				for j := 0; j < 4; j++ {
					if data[j] != want[j] {
						t.Errorf("got %+d, want %+d", data, want)
					}
				}
			default:
				index++
			}
		}

		if index != 4 {
			t.Errorf("got %d, want %d", index, 4)
		}
	})
}

func BenchmarkNodes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ch := NewHintMQ()
		wg := &sync.WaitGroup{}
		wg.Add(100 * 10000)
		for j := 0; j < 100; j++ {
			go func() {
				for k := 0; k < 10000; k++ {
					ch.Write(intToBytes(k))
					wg.Done()
				}
			}()
		}

		wg.Wait()
		wg.Add(100 * 10000)

		for j := 0; j < 100; j++ {
			for k := 0; k < 10000; k++ {
				select {
				case <-ch.Read():
				default:
					b.Errorf("got unexpected")
				}
				wg.Done()
			}
		}
		wg.Wait()
	}
}
