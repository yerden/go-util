package redis

import (
	"fmt"
	"strconv"
)

type Unmarshaller interface {
	Unmarshal(string) fmt.Stringer
}

type mystring string

func (x mystring) String() string { return string(x) }

type String struct{}

func (x String) Unmarshal(in string) fmt.Stringer { return mystring(in) }

type myint16 int16
type myint32 int32
type myint64 int64
type myuint16 uint16
type myuint32 uint32
type myuint64 uint64
type myfloat64 float64

func (x myint16) String() string   { return strconv.FormatInt(int64(x), 10) }
func (x myint32) String() string   { return strconv.FormatInt(int64(x), 10) }
func (x myint64) String() string   { return strconv.FormatInt(int64(x), 10) }
func (x myuint16) String() string  { return strconv.FormatUint(uint64(x), 10) }
func (x myuint32) String() string  { return strconv.FormatUint(uint64(x), 10) }
func (x myuint64) String() string  { return strconv.FormatUint(uint64(x), 10) }
func (x myfloat64) String() string { return fmt.Sprintf("%v", float64(x)) }

type Int16 struct{}
type Int32 struct{}
type Int64 struct{}
type Uint16 struct{}
type Uint32 struct{}
type Uint64 struct{}
type Float64 struct{}

func (x Int16) Unmarshal(in string) fmt.Stringer {
	u, _ := strconv.ParseInt(in, 10, 16)
	return myint16(u)
}

func (x Int32) Unmarshal(in string) fmt.Stringer {
	u, _ := strconv.ParseInt(in, 10, 32)
	return myint32(u)
}

func (x Int64) Unmarshal(in string) fmt.Stringer {
	u, _ := strconv.ParseInt(in, 10, 64)
	return myint64(u)
}

func (x Uint16) Unmarshal(in string) fmt.Stringer {
	u, _ := strconv.ParseUint(in, 10, 16)
	return myuint16(u)
}

func (x Uint32) Unmarshal(in string) fmt.Stringer {
	u, _ := strconv.ParseUint(in, 10, 32)
	return myuint32(u)
}

func (x Uint64) Unmarshal(in string) fmt.Stringer {
	u, _ := strconv.ParseUint(in, 10, 64)
	return myuint64(u)
}

func (x Float64) Unmarshal(in string) fmt.Stringer {
	d, _ := strconv.ParseFloat(in, 64)
	return myfloat64(d)
}
