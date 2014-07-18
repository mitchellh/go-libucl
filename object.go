package libucl

import "unsafe"

// #include "go-libucl.h"
import "C"

// Object represents a single object within a configuration.
type Object struct {
	object *C.ucl_object_t
}

// ObjectIter is an interator for objects.
type ObjectIter struct {
	expand bool
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

// Emitter is a type of built-in emitter that can be used to convert
// an object to another config format.
type Emitter int

const (
	EmitJSON Emitter = iota
	EmitJSONCompact
	EmitConfig
	EmitYAML
)

// Free the memory associated with the object. This must be called when
// you're done using it.
func (o *Object) Close() error {
	C.ucl_object_unref(o.object)
	return nil
}

// Emit converts this object to another format and returns it.
func (o *Object) Emit(t Emitter) (string, error) {
	result := C.ucl_object_emit(o.object, uint32(t))
	if result == nil {
		return "", nil
	}

	return C.GoString(C._go_uchar_to_char(result)), nil
}

// Delete removes the given key from the object. The key will automatically
// be dereferenced once when this is called.
func (o *Object) Delete(key string) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	C.ucl_object_delete_key(o.object, ckey)
}

func (o *Object) Get(key string) *Object {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	obj := C.ucl_object_find_keyl(o.object, ckey, C.size_t(len(key)))
	if obj == nil {
		return nil
	}

	result := &Object{object: obj}
	result.Ref()
	return result
}

// Iterate over the objects in this object.
//
// The iterator must be closed when it is finished.
//
// The iterator does not need to be fully consumed.
func (o *Object) Iterate(expand bool) *ObjectIter {
	// Increase the ref count
	C.ucl_object_ref(o.object)

	return &ObjectIter{
		expand: expand,
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
	// This is weird. If the object is an object and it has a "next",
	// then it is actually an array of objects, and to get the count
	// we actually need to iterate and count.
	if o.Type() == ObjectTypeObject && o.object.next != nil {
		iter := o.Iterate(false)
		defer iter.Close()

		var count uint = 0
		for obj := iter.Next(); obj != nil; obj = iter.Next() {
			obj.Close()
			count += 1
		}

		return count
	}

	return uint(o.object.len)
}

// Increments the ref count associated with this. You have to call
// close an additional time to free the memory.
func (o *Object) Ref() error {
	C.ucl_object_ref(o.object)
	return nil
}

// Returns the type that this object represents.
func (o *Object) Type() ObjectType {
	return ObjectType(C.ucl_object_type(o.object))
}

//------------------------------------------------------------------------
// Conversion Functions
//------------------------------------------------------------------------

func (o *Object) ToBool() bool {
	return bool(C.ucl_object_toboolean(o.object))
}

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
	obj := C.ucl_iterate_object(o.object, &o.iter, C._Bool(o.expand))
	if obj == nil {
		return nil
	}

	// Increase the ref count so we have to free it
	C.ucl_object_ref(obj)

	return &Object{object: obj}
}
