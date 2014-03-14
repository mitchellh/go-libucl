package libucl

import (
	"reflect"
	"testing"
)

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

func TestObjectIterate(t *testing.T) {
	obj := testParseString(t, "foo = bar; bar = baz;")
	defer obj.Close()

	iter := obj.Iterate()
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

	iter := obj.Iterate()
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
