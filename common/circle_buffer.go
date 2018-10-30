package common

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
	L := len(buf.array) - 1
	if L == 0 {
		return 0
	}

	if n >= 0 {
		L &= n
	}

	// if n == -1 return L
	return L
}

func (buf *CBuffer) Init(alloc int) {
	buf.array = make([]interface{}, 0, alloc)
	buf.rd, buf.wr = 0, 0
}

// extend buffer
func (buf *CBuffer) extend() {
	// it means that buf.wr == buf.rd
	// extend array and move rd, wr cursors
	p1, p2 := buf.array[:buf.rd], buf.array[buf.rd:]
	buf.array = append(buf.array, make([]interface{}, buf.addLen())...)

	switch {
	case len(p1) == 0:
		//panic("case 1")
		buf.wr = buf.cnt
	case len(p1) < len(p2):
		//panic("case 2")
		/*
		 *  1         2       1              |
		 *|====wr===========|>>>>------------|
		 *     rd                wr          |
		 */
		buf.wr = buf.wrap(buf.cnt + len(p1))
		dst := buf.array[buf.cnt:buf.wr]
		copy(dst, p1)
	default:
		//panic("case 3")
		/*
		 *       1        2                2 |
		 *|============wr===|------------->>>|
		 *             rd                rd  |
		 */
		buf.rd = buf.wrap(len(buf.array) - len(p2))
		dst := buf.array[buf.rd:]
		copy(dst, p2)
	}
}

// Enqueue element to buffer and
// extend buffer if needed.
func (buf *CBuffer) Enqueue(v interface{}) {
	//fmt.Printf("rd=%d, wr=%d, cnt=%d, len=%d\n", buf.rd, buf.wr, buf.cnt, len(buf.array))
	if buf.cnt == len(buf.array) {
		buf.extend()
	}
	buf.array[buf.wr] = v
	buf.wr = buf.wrap(buf.wr + 1)
	buf.cnt++
}

// Dequeue element from buffer
func (buf *CBuffer) Dequeue() (interface{}, bool) {
	if buf.cnt == 0 {
		return nil, false
	}

	v := buf.array[buf.rd]
	buf.rd = buf.wrap(buf.rd + 1)
	buf.cnt--
	return v, true
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
	buf.array = buf.array[:0]
}
