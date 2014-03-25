package libucl

// #include <ucl.h>
// #include "util.h"
import "C"

// Object represents a single object within a configuration.
type Object struct {
	object *C.ucl_object_t
}

// ObjectIter is an interator for objects.
type ObjectIter struct {
	object *C.ucl_object_t
	iter   C.ucl_object_iter_t
}

// ObjectType is an enum of the type that an Object represents.
type ObjectType int

const (
	ObjectTypeObject ObjectType = iota
	ObjectTypeArray
	ObjectTypeInt
	ObjectTypeFloat
	ObjectTypeString
	ObjectTypeBoolean
	ObjectTypeTime
	ObjectTypeUserData
	ObjectTypeNull
)

type Emitter int

const (
	EmitJSON Emitter = iota
	EmitJSONCompact
	EmitConfig
	EmitYAML
)

// Free the memory associated with the object. This must be called when
// you're done using it.
func (o *Object) Close() {
	C.ucl_object_unref(o.object)
}

// Emit converts this object to another format and returns it.
func (o *Object) Emit(t Emitter) (string, error) {
	result := C.ucl_object_emit(o.object, uint32(t))
	if result == nil {
		return "", nil
	}

	return C.GoString(C._go_uchar_to_char(result)), nil
}

func (o *Object) Get(key string) *Object {
	obj := C.ucl_object_find_keyl(o.object, C.CString(key), C.size_t(len(key)))
	if obj == nil {
		return nil
	}

	return &Object{object: obj}
}

// Iterate over the objects in this object.
//
// The iterator must be closed when it is finished.
//
// The iterator does not need to be fully consumed.
func (o *Object) Iterate() *ObjectIter {
	// Increase the ref count
	C.ucl_object_ref(o.object)

	return &ObjectIter{
		object: o.object,
		iter:   nil,
	}
}

// Returns the key of this value/object as a string, or the empty
// string if the object doesn't have a key.
func (o *Object) Key() string {
	return C.GoString(C.ucl_object_key(o.object))
}

// Len returns the length of the object, or how many elements are part
// of this object.
//
// For objects, this is the number of key/value pairs.
// For arrays, this is the number of elements.
func (o *Object) Len() uint {
	return uint(o.object.len)
}

// Returns the type that this object represents.
func (o *Object) Type() ObjectType {
	return ObjectType(o.object._type)
}

//------------------------------------------------------------------------
// Conversion Functions
//------------------------------------------------------------------------

func (o *Object) ToInt() int64 {
	return int64(C.ucl_object_toint(o.object))
}

func (o *Object) ToString() string {
	return C.GoString(C.ucl_object_tostring(o.object))
}

func (o *ObjectIter) Close() {
	C.ucl_object_unref(o.object)
}

func (o *ObjectIter) Next() *Object {
	obj := C.ucl_iterate_object(o.object, &o.iter, true)
	if obj == nil {
		return nil
	}

	// Increase the ref count so we have to free it
	C.ucl_object_ref(obj)

	return &Object{object: obj}
}
