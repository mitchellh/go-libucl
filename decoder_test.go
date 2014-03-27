package libucl

import (
	"reflect"
	"testing"
)

func TestObjectDecode_map(t *testing.T) {
	obj := testParseString(t, "foo = bar; bar = 12;")
	defer obj.Close()

	var result map[string]string
	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	if result["foo"] != "bar" {
		t.Fatalf("bad: %#v", result["foo"])
	}
	if result["bar"] != "12" {
		t.Fatalf("bad: %#v", result["bar"])
	}
}

func TestObjectDecode_mapMulti(t *testing.T) {
	obj := testParseString(t, "foo { foo = bar; }; foo { bar = baz; };")
	defer obj.Close()

	inner := obj.Get("foo")
	defer inner.Close()

	var result map[string]string
	if err := inner.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	if result["foo"] != "bar" {
		t.Fatalf("bad: %#v", result["foo"])
	}
	if result["bar"] != "baz" {
		t.Fatalf("bad: %#v", result["bar"])
	}
}

func TestObjectDecode_mapNonNil(t *testing.T) {
	obj := testParseString(t, "foo = bar; bar = 12;")
	defer obj.Close()

	result := map[string]string{"hey": "hello"}
	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	if result["hey"] != "hello" {
		t.Fatalf("bad hey!")
	}
	if result["foo"] != "bar" {
		t.Fatalf("bad: %#v", result["foo"])
	}
	if result["bar"] != "12" {
		t.Fatalf("bad: %#v", result["bar"])
	}
}

func TestObjectDecode_mapNonObject(t *testing.T) {
	obj := testParseString(t, "foo = [bar];")
	defer obj.Close()

	obj = obj.Get("foo")
	defer obj.Close()

	var result map[string]string
	if err := obj.Decode(&result); err == nil {
		t.Fatal("should fail")
	}
}

func TestObjectDecode_nestedStruct(t *testing.T) {
	type Nested struct {
		Foo string
	}

	var result struct {
		Value Nested
	}

	obj := testParseString(t, `value { foo = "bar"; }`)
	defer obj.Close()

	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	if result.Value.Foo != "bar" {
		t.Fatalf("bad: %#v", result.Value.Foo)
	}
}

func TestObjectDecode_nestedStructRepeated(t *testing.T) {
	type Nested struct {
		Foo string
		Bar string
	}

	var result struct {
		Value Nested
	}

	obj := testParseString(t, `value { foo = "bar"; }; value { bar = "baz" };`)
	defer obj.Close()

	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	if result.Value.Foo != "bar" {
		t.Fatalf("bad: %#v", result.Value.Foo)
	}
	if result.Value.Bar != "baz" {
		t.Fatalf("bad: %#v", result.Value.Bar)
	}
}

func TestObjectDecode_slice(t *testing.T) {
	obj := testParseString(t, "foo = [foo, bar, 12];")
	defer obj.Close()

	obj = obj.Get("foo")
	defer obj.Close()

	var result []string
	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := []string{"foo", "bar", "12"}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestObjectDecode_struct(t *testing.T) {
	var result struct {
		Foo []string
		Bar string
	}

	obj := testParseString(t, "foo = [foo, bar, 12]; bar = baz;")
	defer obj.Close()

	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := []string{"foo", "bar", "12"}
	if !reflect.DeepEqual(expected, result.Foo) {
		t.Fatalf("bad: %#v", result.Foo)
	}
	if result.Bar != "baz" {
		t.Fatalf("bad: %#v", result.Bar)
	}
}

func TestObjectDecode_structArray(t *testing.T) {
	type Nested struct {
		Foo string
	}

	var result struct {
		Value []*Nested
	}

	obj := testParseString(t, `value { foo = "bar"; }; value { foo = "baz"; }`)
	defer obj.Close()

	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(result.Value) != 2 {
		t.Fatalf("bad: %#v", result.Value)
	}
	if result.Value[0].Foo != "bar" {
		t.Fatalf("bad: %#v", result.Value[0].Foo)
	}
	if result.Value[1].Foo != "baz" {
		t.Fatalf("bad: %#v", result.Value[1].Foo)
	}
}

func TestObjectDecode_structSquash(t *testing.T) {
	type Foo struct {
		Baz string
	}

	var result struct {
		Bar string
		Foo `libucl:",squash"`
	}

	obj := testParseString(t, "baz = what; bar = baz;")
	defer obj.Close()

	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	if result.Bar != "baz" {
		t.Fatalf("bad: %#v", result.Bar)
	}
	if result.Baz != "what" {
		t.Fatalf("bad: %#v", result.Baz)
	}
}
