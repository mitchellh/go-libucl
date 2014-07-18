package libucl

import (
	"reflect"
	"testing"
)

func TestObjectEmit(t *testing.T) {
	obj := testParseString(t, "foo = bar; bar = baz;")
	defer obj.Close()

	result, err := obj.Emit(EmitJSON)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := "{\n    \"foo\": \"bar\",\n    \"bar\": \"baz\"\n}"
	if result != expected {
		t.Fatalf("bad: %#v", result)
	}
}

func TestObjectEmit_EmitConfig(t *testing.T) {
	obj := testParseString(t, "foo = bar; bar = baz;")
	defer obj.Close()

	result, err := obj.Emit(EmitConfig)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := "foo = \"bar\";\nbar = \"baz\";\n"
	if result != expected {
		t.Fatalf("bad: %#v", result)
	}
}

func TestObjectDelete(t *testing.T) {
	obj := testParseString(t, "bar = baz;")
	defer obj.Close()

	v := obj.Get("bar")
	if v == nil {
		t.Fatal("should find")
	}
	v.Close()

	obj.Delete("bar")
	v = obj.Get("bar")
	if v != nil {
		v.Close()
		t.Fatalf("should not find")
	}
}

func TestObjectDelete_unknown(t *testing.T) {
	obj := testParseString(t, "bar = baz;")
	defer obj.Close()

	obj.Delete("foo")
}

func TestObjectGet(t *testing.T) {
	obj := testParseString(t, "foo = bar; bar = baz;")
	defer obj.Close()

	v := obj.Get("bar")
	defer v.Close()
	if v == nil {
		t.Fatal("should find")
	}

	if v.Key() != "bar" {
		t.Fatalf("bad: %#v", v.Key())
	}
	if v.ToString() != "baz" {
		t.Fatalf("bad: %#v", v.ToString())
	}
}

func TestObjectLen_array(t *testing.T) {
	obj := testParseString(t, "foo = [foo, bar, baz];")
	defer obj.Close()

	v := obj.Get("foo")
	defer v.Close()
	if v == nil {
		t.Fatal("should find")
	}

	if v.Len() != 3 {
		t.Fatalf("bad: %#v", v.Len())
	}
}

func TestObjectLen_object(t *testing.T) {
	obj := testParseString(t, `bundle "foo" {}; bundle "bar" {};`)
	defer obj.Close()

	v := obj.Get("bundle")
	defer v.Close()
	if v == nil {
		t.Fatal("should find")
	}

	if v.Type() != ObjectTypeObject {
		t.Fatalf("bad: %#v", v.Type())
	}
	if v.Len() != 2 {
		t.Fatalf("bad: %#v", v.Len())
	}
}

func TestObjectIterate(t *testing.T) {
	obj := testParseString(t, "foo = bar; bar = baz;")
	defer obj.Close()

	iter := obj.Iterate(true)
	defer iter.Close()

	result := make([]string, 0, 10)
	for elem := iter.Next(); elem != nil; elem = iter.Next() {
		defer elem.Close()
		result = append(result, elem.Key())
		result = append(result, elem.ToString())
	}

	expected := []string{"foo", "bar", "bar", "baz"}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestObjectIterate_array(t *testing.T) {
	obj := testParseString(t, "foo = [foo, bar, baz];")
	defer obj.Close()

	obj = obj.Get("foo")
	if obj == nil {
		t.Fatal("should have object")
	}

	iter := obj.Iterate(true)
	defer iter.Close()

	result := make([]string, 0, 10)
	for elem := iter.Next(); elem != nil; elem = iter.Next() {
		defer elem.Close()
		result = append(result, elem.ToString())
	}

	expected := []string{"foo", "bar", "baz"}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestObjectToBool(t *testing.T) {
	obj := testParseString(t, "foo = true; bar = false;")
	defer obj.Close()

	v := obj.Get("bar")
	defer v.Close()
	if v == nil {
		t.Fatal("should find")
	}
	if v.ToBool() {
		t.Fatalf("bad: %#v", v.ToBool())
	}
}
