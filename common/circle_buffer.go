package common

/*
 * This is an implementation of circular buffer.
 * It implements  Queue and Stack interfaces.
 * The underlying buffer extends upon request.
 */

var _ Queue = (*CBuffer)(nil)
var _ Stack = (*CBuffer)(nil)

type CBuffer struct {
	array       []interface{}
	rd, wr, cnt int
}

// that makes length of buffer a power of 2
func (buf *CBuffer) addLen() int {
	if x := len(buf.array); x != 0 {
		return x
	}
	return 1
}

func (buf *CBuffer) wrap(n int) int {
	return n & (len(buf.array) - 1)
}

func (buf *CBuffer) Init(alloc int) {
	buf.array = make([]interface{}, alloc)
	buf.Clear()
}

// extend buffer
func (buf *CBuffer) extend() {
	// it means that buf.wr == buf.rd
	// extend array and move rd, wr cursors
	p1, p2 := buf.array[:buf.rd], buf.array[buf.rd:]
	buf.array = append(buf.array, make([]interface{}, buf.addLen())...)

	if newoff := buf.cnt + len(p1); len(p1) < len(p2) {
		//panic("case 2")
		/*
		 *  1         2       1               |
		 *|====wr===========|>>>>-------------|
		 *     rd                wr           |
		 */
		buf.wr = newoff
		copy(buf.array[buf.cnt:buf.wr], p1)
	} else {
		//panic("case 3")
		/*
		 *       1        2                 2 |
		 *|============wr===|-------------->>>|
		 *             rd                 rd  |
		 */
		buf.rd = newoff
		copy(buf.array[buf.rd:], p2)
	}
}

func (buf *CBuffer) Push(v interface{}) error {
	return buf.Enqueue(v)
}

// Enqueue element to buffer and
// extend buffer if needed.
func (buf *CBuffer) Enqueue(v interface{}) error {
	//fmt.Printf("rd=%d, wr=%d, cnt=%d, len=%d\n", buf.rd, buf.wr, buf.cnt, len(buf.array))
	if buf.cnt == len(buf.array) {
		buf.extend()
	}
	buf.array[buf.wr] = v
	buf.wr = buf.wrap(buf.wr + 1)
	buf.cnt++
	return nil
}

// Dequeue element from buffer
func (buf *CBuffer) Dequeue() (v interface{}, ok bool) {
	if ok = (buf.cnt > 0); ok {
		v = buf.array[buf.rd]
		buf.rd = buf.wrap(buf.rd + 1)
		buf.cnt--
	}
	return
}

// Pop element from buffer
func (buf *CBuffer) Pop() (v interface{}, ok bool) {
	if ok = (buf.cnt > 0); ok {
		buf.wr = buf.wrap(buf.wr - 1)
		v = buf.array[buf.wr]
		buf.cnt--
	}
	return
}

// Number of elements in buffer
func (buf *CBuffer) Len() int {
	return buf.cnt
}

// Number of free slots in buffer
// which can be utilized without buffer extension
func (buf *CBuffer) Cap() int {
	return len(buf.array)
}

// Flush buffer, removing all elements
func (buf *CBuffer) Clear() {
	buf.rd, buf.wr, buf.cnt = 0, 0, 0
}
