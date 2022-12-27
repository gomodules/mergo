package mergo_test

import (
	"reflect"
	"testing"

	"gomodules.xyz/mergo"
)

type transformer struct {
	m map[reflect.Type]func(dst, src reflect.Value) error
}

func (s *transformer) Transformer(t reflect.Type) func(dst, src reflect.Value) error {
	if fn, ok := s.m[t]; ok {
		return fn
	}
	return nil
}

type foo struct {
	s   string
	Bar *bar
}

type bar struct {
	i int
	s map[string]string
}

func TestMergeWithTransformerNilStruct(t *testing.T) {
	a := foo{s: "foo"}
	b := foo{Bar: &bar{i: 2, s: map[string]string{"foo": "bar"}}}

	if err := mergo.Merge(&a, &b, mergo.WithOverride, mergo.WithTransformers(&transformer{
		m: map[reflect.Type]func(dst, src reflect.Value) error{
			reflect.TypeOf(&bar{}): func(dst, src reflect.Value) error {
				// Do sthg with Elem
				t.Log(dst.Elem().FieldByName("i"))
				t.Log(src.Elem())
				return nil
			},
		},
	})); err != nil {
		t.Error(err)
	}

	if a.s != "foo" {
		t.Errorf("b not merged in properly: a.s.Value(%s) != expected(%s)", a.s, "foo")
	}

	if a.Bar == nil {
		t.Errorf("b not merged in properly: a.Bar shouldn't be nil")
	}
}

func TestMergeNonPointer(t *testing.T) {
	dst := bar{
		i: 1,
	}
	src := bar{
		i: 2,
		s: map[string]string{
			"a": "1",
		},
	}
	want := mergo.ErrNonPointerAgument

	if got := mergo.Merge(dst, src); got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestMapNonPointer(t *testing.T) {
	dst := make(map[string]bar)
	src := map[string]bar{
		"a": {
			i: 2,
			s: map[string]string{
				"a": "1",
			},
		},
	}
	want := mergo.ErrNonPointerAgument
	if got := mergo.Merge(dst, src); got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

type data struct {
	B *bool
}

func TestMergeBoolPointerWithNullTransformer(t *testing.T) {
	dst := data{
		B: func(v bool) *bool { return &v }(true),
	}

	src := data{
		B:func(v bool) *bool { return &v }(false),
	}

	err := mergo.Merge(&dst, src, mergo.WithOverride, mergo.WithTransformers(mergo.NewNullTransformer()))
	if err != nil {
		t.Error(err)
	}

	if *dst.B != *src.B{
		t.Errorf("src not merged in properly: dst.B.Value(%v) != expected(%v)", *dst.B, *src.B)
	}
}
