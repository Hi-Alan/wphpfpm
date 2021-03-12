package phpfpm

import (
	"runtime"
	"sync/atomic"
)

// roundUp takes a uint64 greater than 0 and rounds it up to the next
// power of 2.
// https://github.com/golang-collections/go-datastructures/blob/master/queue/ring.go
// 相当于将低位全部填1
func roundUp(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	// v |= v >> 32 // for uint64
	v++
	return v
}

// NewRingBuffer - Simple RingBuffer pacakge
// Ringbuffer is non blocking for readers and writers, writers will
// overwrite older data in a circular fashion. Readers will read
// from the current position and update it.
// https://github.com/sahmad98/go-ringbuffer/blob/master/ring_buffer.go

// https://github.com/golang-collections/go-datastructures/blob/master/queue/ring.go

// RingBuffer Structure
type RingBuffer struct {
	capacity  uint32     // capacity of the Ringbuffer
	mask      uint32     // capacity - 1
	len       uint32     // Current size of the Ringbuffer
	container []*Process // Array container of objects
	status    []uint32   // Array container of objects operation status
	reader    uint32     // Reader position 0 - max_uint32
	writer    uint32     // Writer Position
}

// NewRingBuffer Create a new RingBuffer of initial size "size"
// Returns a pointer to the new RingBuffer
func NewRingBuffer(capacity uint32) *RingBuffer {
	capacity = roundUp(capacity) // be a power of 2, let index & (size -1) === index % size

	rb := &RingBuffer{
		capacity:  (capacity),
		mask:      (capacity - 1),
		len:       0,
		container: make([]*Process, capacity),
		status:    make([]uint32, capacity),
		reader:    0,
		writer:    0,
	}

	return rb
}

// Put object into the RingBuffer
func (r *RingBuffer) Put(v *Process) bool {

	swapped := false
	i := 0
	for i < 10 && !swapped {

		pos := atomic.LoadUint32(&r.reader)
		current := atomic.LoadUint32(&r.writer)
		if current > pos-1+r.capacity { // full; overflow will be zero
			runtime.Gosched()
			i++
			continue
		}
		swapped = atomic.CompareAndSwapUint32(&r.writer, current, current+1) // 避免并发写入, 累加溢出后归零
		if swapped {
			current = current & r.mask // 代替 (current + 1) % r.len 运算
			i = 0
			for i < 10 { // wait reading
				status := atomic.LoadUint32(&r.status[current])
				if status == 0 {
					r.container[current] = v
					atomic.CompareAndSwapUint32(&r.status[current], 0, 1) // write completed
					return true
				}
				runtime.Gosched()
				i++
			}
			return false // write failured
		}
		runtime.Gosched()
		i++
	}
	return false
}

// Get single object from the RingBuffer
func (r *RingBuffer) Get() *Process {

	swapped := false
	i := 0
	for i < 10 && !swapped {
		pos := atomic.LoadUint32(&r.writer)
		current := atomic.LoadUint32(&r.reader)
		if pos == current { // empty
			return nil
		}
		swapped = atomic.CompareAndSwapUint32(&r.reader, current, current+1) // 避免并发写入, 一直累加溢出后归零
		if swapped {
			current := (current & r.mask) // 代替 (current + 1) % r.len 运算
			i = 0
			for i < 10 { // wait writing
				status := atomic.LoadUint32(&r.status[current])
				if status == 1 {
					val := r.container[current]
					atomic.CompareAndSwapUint32(&r.status[current], 1, 0) // read completed
					return val
				}
				runtime.Gosched()
				i++
			}
			return nil // can not read
		}
		runtime.Gosched()
		i++
	}
	return nil
}

// Len get ringbuffer current size
func (r *RingBuffer) Len() uint32 {
	pos := atomic.LoadUint32(&r.writer)
	current := atomic.LoadUint32(&r.reader)
	return pos - current
}

// Read single object from the RingBuffer
func (r *RingBuffer) _Read(idx *uint32) (uint32, bool) {
	swapped := false
	i := 0
	for i < 10 && !swapped {
		current := atomic.LoadUint32(idx)
		swapped = atomic.CompareAndSwapUint32(idx, current, current+1) // 避免并发写入, 一直累加溢出后归零
		if swapped {
			return current & r.mask, true // 代替 (current + 1) % r.len 运算
		}
		runtime.Gosched()
		i++
	}
	return 0, false
}
