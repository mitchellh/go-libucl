package libucl

import (
	"io/ioutil"
	"testing"
)

func testParseString(t *testing.T, data string) *Object {
	obj, err := ParseString(data)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return obj
}

func TestParser(t *testing.T) {
	p := NewParser(0)
	defer p.Close()

	if err := p.AddString(`foo = bar;`); err != nil {
		t.Fatalf("err: %s", err)
	}

	obj := p.Object()
	if obj == nil {
		t.Fatal("obj should not be nil")
	}
	defer obj.Close()

	if obj.Type() != ObjectTypeObject {
		t.Fatalf("bad: %#v", obj.Type())
	}

	value := obj.Get("foo")
	if value == nil {
		t.Fatal("should have value")
	}
	defer value.Close()

	if value.Type() != ObjectTypeString {
		t.Fatalf("bad: %#v", obj.Type())
	}

	if value.Key() != "foo" {
		t.Fatalf("bad: %#v", value.Key())
	}

	if value.ToString() != "bar" {
		t.Fatalf("bad: %#v", value.ToString())
	}
}

func TestParserAddFile(t *testing.T) {
	tf, err := ioutil.TempFile("", "libucl")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Write([]byte("foo = bar;"))
	tf.Close()

	p := NewParser(0)
	defer p.Close()

	if err := p.AddFile(tf.Name()); err != nil {
		t.Fatalf("err: %s", err)
	}

	obj := p.Object()
	if obj == nil {
		t.Fatal("obj should not be nil")
	}
	defer obj.Close()

	if obj.Type() != ObjectTypeObject {
		t.Fatalf("bad: %#v", obj.Type())
	}

	value := obj.Get("foo")
	if value == nil {
		t.Fatal("should have value")
	}
	defer value.Close()

	if value.Type() != ObjectTypeString {
		t.Fatalf("bad: %#v", obj.Type())
	}

	if value.Key() != "foo" {
		t.Fatalf("bad: %#v", value.Key())
	}

	if value.ToString() != "bar" {
		t.Fatalf("bad: %#v", value.ToString())
	}
}

func TestParserRegisterMacro(t *testing.T) {
	value := ""
	macro := func(data string) {
		value = data
	}

	config := `.foo "bar";`

	p := NewParser(0)
	defer p.Close()

	p.RegisterMacro("foo", macro)

	if err := p.AddString(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	if value != "bar" {
		t.Fatalf("bad: %#v", value)
	}
}

func TestParseString(t *testing.T) {
	obj, err := ParseString("foo = bar; baz = boo;")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if obj == nil {
		t.Fatal("should have object")
	}
	defer obj.Close()

	if obj.Len() != 2 {
		t.Fatalf("bad: %d", obj.Len())
	}
}
