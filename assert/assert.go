package assert

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

type assert struct {
	t *testing.T
}

type Assert interface {
	Nil(interface{}, ...string)
	NotNil(interface{}, ...string)
	Equal(interface{}, interface{}, ...string)
	NotEqual(interface{}, interface{}, ...string)
	True(bool, ...string)
	NotTrue(bool, ...string)
}

var (
	nilStr = func(x interface{}) string {
		return fmt.Sprintf("%v != nil", x)
	}
	notNilStr = func(x interface{}) string {
		return fmt.Sprintf("%v == nil", x)
	}
	equalStr = func(x, y interface{}) string {
		return fmt.Sprintf("%v != %v", x, y)
	}
	notEqualStr = func(x, y interface{}) string {
		return fmt.Sprintf("%v != %v", x, y)
	}
	isTrue = func(x bool) string {
		return fmt.Sprintf("%v != %v", x, true)
	}
	isNotTrue = func(x bool) string {
		return fmt.Sprintf("%v != %v", x, false)
	}
)

func New(t *testing.T) Assert {
	return &assert{t: t}
}

func (a *assert) checkTrue(x bool, dfl string, s []string) {
	_, fileName, fileLine, ok := runtime.Caller(2)
	if !ok || x {
		return
	}
	if len(s) == 0 {
		s = append(s, dfl)
	}
	fmt.Printf("%s:%d: %s\n", fileName, fileLine, s)
	a.t.FailNow()
}

func (a *assert) True(x bool, s ...string) {
	a.checkTrue(x, isTrue(x), s)
}

func (a *assert) NotTrue(x bool, s ...string) {
	a.checkTrue(!x, isNotTrue(x), s)
}

func (a *assert) Equal(x interface{}, y interface{}, s ...string) {
	a.checkTrue(reflect.DeepEqual(x, y), equalStr(x, y), s)
}

func (a *assert) NotEqual(x interface{}, y interface{}, s ...string) {
	a.checkTrue(!reflect.DeepEqual(x, y), notEqualStr(x, y), s)
}

func (a *assert) NotNil(x interface{}, s ...string) {
	a.checkTrue(x != nil, notNilStr(x), s)
}

func (a *assert) Nil(x interface{}, s ...string) {
	a.checkTrue(x == nil, nilStr(x), s)
}
