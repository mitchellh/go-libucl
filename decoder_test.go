package libucl

import (
	"reflect"
	"testing"
)

func TestObjectDecode_basic(t *testing.T) {
	type Basic struct {
		Bool    bool
		BoolStr string
		Str     string
		Num     int
		NumStr  int
	}

	obj := testParseString(t, `
	bool = true; str = bar; num = 7; numstr = "42";
	boolstr = true;
	`)
	defer obj.Close()

	var result Basic
	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := Basic{
		Bool:    true,
		BoolStr: "true",
		Str:     "bar",
		Num:     7,
		NumStr:  42,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestObjectDecode_interface(t *testing.T) {
	obj := testParseString(t, `
	foo {
		f1 = "foo";
		f2 = [1, 2, 3];
		f3 = ["foo", 2, 42];
		f4 = true;
	}
	`)
	defer obj.Close()

	obj = obj.Get("foo")
	defer obj.Close()

	var result map[string]interface{}
	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(result) != 4 {
		t.Fatalf("bad: %#v", result)
	}

	if result["f1"].(string) != "foo" {
		t.Fatalf("bad: %#v", result)
	}

	expected := []interface{}{1, 2, 3}
	if !reflect.DeepEqual(result["f2"], expected) {
		t.Fatalf("bad: %#v", result["f2"])
	}

	expected = []interface{}{"foo", 2, 42}
	if !reflect.DeepEqual(result["f3"], expected) {
		t.Fatalf("bad: %#v", result["f3"])
	}

	if result["f4"].(bool) != true {
		t.Fatalf("bad: %#v", result)
	}
}

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

func TestObjectDecode_mapObject(t *testing.T) {
	obj := testParseString(t, "foo = bar; bar { baz = \"what\" }")
	defer obj.Close()

	var result map[string]interface{}
	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := map[string]interface{}{
		"foo": "bar",
		"bar": []map[string]interface{}{
			map[string]interface{}{
				"baz": "what",
			},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestObjectDecode_mapObjectMultiple(t *testing.T) {
	obj := testParseString(t, `
	foo = bar
	bar { baz = "what" }
	bar { port = 3000 }
`)
	defer obj.Close()

	var result map[string]interface{}
	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := map[string]interface{}{
		"foo": "bar",
		"bar": []map[string]interface{}{
			map[string]interface{}{
				"baz": "what",
			},
			map[string]interface{}{
				"port": 3000,
			},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestObjectDecode_mapReuseVal(t *testing.T) {
	type Struct struct {
		Foo string
		Bar string
	}

	type Result struct {
		Struct map[string]Struct
	}

	obj := testParseString(t, `
		struct "foo" { foo = "bar"; };
		struct "foo" { bar = "baz"; };
		`)
	defer obj.Close()

	var result Result
	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := Result{
		Struct: map[string]Struct{
			"foo": Struct{
				Foo: "bar",
				Bar: "baz",
			},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
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

func TestObjectDecode_sliceRepeatedKey(t *testing.T) {
	obj := testParseString(t, "foo = foo; foo = bar;")
	defer obj.Close()

	obj = obj.Get("foo")
	defer obj.Close()

	var result []string
	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := []string{"foo", "bar"}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestObjectDecode_sliceSingle(t *testing.T) {
	obj := testParseString(t, "foo = bar;")
	defer obj.Close()

	obj = obj.Get("foo")
	defer obj.Close()

	var result []string
	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := []string{"bar"}
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

func TestObjectDecode_structKeys(t *testing.T) {
	type Struct struct {
		Foo  []string
		Bar  string
		Baz  string
		Keys []string `libucl:",decodedFields"`
	}

	var result Struct

	obj := testParseString(t, "foo = [foo, bar, 12]; bar = baz;")
	defer obj.Close()

	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := Struct{
		Foo:  []string{"foo", "bar", "12"},
		Bar:  "baz",
		Keys: []string{"Foo", "Bar"},
	}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestObjectDecode_mapStructNamed(t *testing.T) {
	type Nested struct {
		Name string `libucl:",key"`
		Foo  string
	}

	var result struct {
		Value map[string]Nested
	}

	obj := testParseString(t, `value "foo" { foo = "bar"; };`)
	defer obj.Close()

	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := map[string]Nested{
		"foo": Nested{
			Name: "foo",
			Foo:  "bar",
		},
	}

	if !reflect.DeepEqual(result.Value, expected) {
		t.Fatalf("bad: %#v", result.Value)
	}
}

func TestObjectDecode_mapStructObject(t *testing.T) {
	type Nested struct {
		Foo    string
		Object *Object `libucl:",object"`
	}

	var result struct {
		Value map[string]Nested
	}

	obj := testParseString(t, `value "foo" { foo = "bar"; };`)
	defer obj.Close()

	valueObj := obj.Get("value")
	defer valueObj.Close()

	fooObj := valueObj.Get("foo")
	defer fooObj.Close()

	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer result.Value["foo"].Object.Close()

	expected := map[string]Nested{
		"foo": Nested{
			Foo:    "bar",
			Object: fooObj,
		},
	}

	if !reflect.DeepEqual(result.Value, expected) {
		t.Fatalf("bad: %#v", result.Value)
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

func TestObjectDecode_structUnusedKeys(t *testing.T) {
	type Struct struct {
		Bar  string
		Keys []string `libucl:",unusedKeys"`
	}

	var result Struct

	obj := testParseString(t, "foo = [bar]; bar = baz; baz = what;")
	defer obj.Close()

	if err := obj.Decode(&result); err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := Struct{
		Bar:  "baz",
		Keys: []string{"foo", "baz"},
	}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("bad: %#v", result)
	}
}
