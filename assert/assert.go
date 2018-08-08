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

func New(t *testing.T) Assert {
	return &assert{t: t}
}

func (a *assert) checkTrue(x bool, s []string) {
	_, fileName, fileLine, ok := runtime.Caller(2)
	if !ok || x {
		return
	}
	fmt.Printf("%s:%d: %s\n", fileName, fileLine, s)
	a.t.FailNow()
}

func (a *assert) True(x bool, s ...string) {
	a.checkTrue(x, s)
}

func (a *assert) NotTrue(x bool, s ...string) {
	a.checkTrue(!x, s)
}

func (a *assert) Equal(x interface{}, y interface{}, s ...string) {
	a.checkTrue(reflect.DeepEqual(x, y), s)
}

func (a *assert) NotEqual(x interface{}, y interface{}, s ...string) {
	a.checkTrue(!reflect.DeepEqual(x, y), s)
}

func (a *assert) NotNil(x interface{}, s ...string) {
	a.checkTrue(x != nil, s)
}

func (a *assert) Nil(x interface{}, s ...string) {
	a.checkTrue(x == nil, s)
}
