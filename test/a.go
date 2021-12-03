package test

import (
	"go2seq/test/bar"
	"go2seq/test/foo"
)

func start() {
	other()
	_ = 123
	w := what{
		t: test{},
	}
	w.t.Test1("test")
	w.getMeOut()
	foo.FooSomeMore("yay")
}

func other() {
	foo.FooSomething()
	bar.BarSomeMore()
}

type what struct {
	t test
}

func (w what) getMeOut() {
	w.t.Test1("now")
}

type test struct{}

func (t test) Test1(breakMe string) {
	bar.BarSomething()
}
