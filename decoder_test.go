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
