package rocketmq

import (
	"sync"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
)

var (
	testPool = sync.Pool{
		New: func() interface{} {
			return &Test{}
		},
	}

	testPool = sync.Pool{
		New: func() interface{} {
			return &Test{}
		},
	}
)

var (
	testPool = sync.Pool{
		New: func() interface{} {
			return &Test{}
		},
	}
)

//nolint:unused,structcheck

type Test struct {
	g byte

	c int8

	i bool

	d int16

	f float32

	a int32

	n func(interface{})

	m map[string]interface{}

	b int64

	conf config.Config

	e string

	logger log.Logger

	pkg.Fields

	h []int

	u [3]string

	pkg.Field
}

type TestOption func(*Test)

func WithTestConf(conf config.Config) TestOption {
	return func(t *Test) {
		t.conf = conf
	}
}

func WithTestLogger(logger log.Logger) TestOption {
	return func(t *Test) {
		t.logger = logger
	}
}

func (t *Test) Release() {
	t.g = 0
	t.c = 0
	t.i = false
	t.d = 0
	t.f = 0.00
	t.a = 0
	t.n = nil
	for k, _ := range t.m {
		delete(t.m, k)
	}
	t.b = 0
	t.conf = nil
	t.e = ""
	t.logger = nil
	t.Fields = t.Fields[:0]
	t.h = t.h[:0]
	for k, _ := range t.u {
		t.u[k] = ""
	}
	t.Field.Name = ""
	t.Field.Type = ""
	t.Field.TypeName = ""
	t.Field.Field = ""
	t.Field.Size = 0
	t.Field.Doc = t.Field.Doc[:0]
	t.Field.Tag = ""
	testPool.Put(t)
}

type TestOption func(*Test)

func WithTestConf(conf config.Config) TestOption {
	return func(t *Test) {
		t.conf = conf
	}
}

func WithTestLogger(logger log.Logger) TestOption {
	return func(t *Test) {
		t.logger = logger
	}
}

func (t *Test) Release() {
	t.b = 0
	t.c = 0
	t.i = false
	t.f = 0.00
	t.a = 0
	t.h = t.h[:0]
	for k, _ := range t.m {
		delete(t.m, k)
	}
	t.e = ""
	t.g = 0
	for k, _ := range t.u {
		t.u[k] = ""
	}
	t.d = 0
	t.Fields = t.Fields[:0]
	t.Field.Name = ""
	t.Field.Type = ""
	t.Field.TypeName = ""
	t.Field.Field = ""
	t.Field.Size = 0
	t.Field.Doc = t.Field.Doc[:0]
	t.Field.Tag = ""
	t.n = nil
	t.logger = nil
	t.conf = nil
	testPool.Put(t)
}
