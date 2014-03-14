package libucl

import (
	"fmt"
	"reflect"
	"strconv"
)

// Decode decodes a libucl object into a native Go structure.
func (o *Object) Decode(v interface{}) error {
	return decode("", o, reflect.ValueOf(v).Elem())
}

func decode(name string, o *Object, result reflect.Value) error {
	switch result.Kind() {
	case reflect.Map:
		return decodeIntoMap(name, o, result)
	case reflect.Slice:
		return decodeIntoSlice(name, o, result)
	case reflect.String:
		return decodeIntoString(name, o, result)
	default:
		return fmt.Errorf("%s: unsupported type: %s", name, result.Kind())
	}

	return nil
}

func decodeIntoMap(name string, o *Object, result reflect.Value) error {
	if o.Type() != ObjectTypeObject {
		return fmt.Errorf("%s: not an object type, can't decode to map", name)
	}

	resultType := result.Type()
	resultElemType := resultType.Elem()
	resultKeyType := resultType.Key()
	if resultKeyType.Kind() != reflect.String {
		return fmt.Errorf("%s: map must have string keys", name)
	}

	// Make a map to store our result
	resultMap := reflect.MakeMap(reflect.MapOf(resultKeyType, resultElemType))

	iter := o.Iterate()
	defer iter.Close()
	for elem := iter.Next(); elem != nil; elem = iter.Next() {
		fieldName := fmt.Sprintf("%s[%s]", name, elem.Key())

		// The key is just the key of the object
		key := reflect.ValueOf(elem.Key())

		// The value we have to be decode
		val := reflect.Indirect(reflect.New(resultElemType))
		err := decode(fieldName, elem, val)
		elem.Close()
		if err != nil {
			return err
		}

		resultMap.SetMapIndex(key, val)
	}

	// Set the final result
	result.Set(resultMap)

	return nil
}

func decodeIntoSlice(name string, o *Object, result reflect.Value) error {
	if o.Type() != ObjectTypeArray {
		return fmt.Errorf("%s: is not type array", name)
	}

	// Create the slice
	resultType := result.Type()
	resultElemType := resultType.Elem()
	resultSliceType := reflect.SliceOf(resultElemType)
	resultSlice := reflect.MakeSlice(resultSliceType, int(o.Len()), int(o.Len()))

	i := 0;
	iter := o.Iterate()
	defer iter.Close()
	for elem := iter.Next(); elem != nil; elem = iter.Next() {
		val := resultSlice.Index(i)
		fieldName := fmt.Sprintf("%s[%d]", name, i)
		err := decode(fieldName, elem, val)
		elem.Close()
		if err != nil {
			return err
		}

		i++
	}

	result.Set(resultSlice)

	return nil
}

func decodeIntoString(name string, o *Object, result reflect.Value) error {
	objType := o.Type()
	switch objType {
	case ObjectTypeString:
		result.SetString(o.ToString())
	case ObjectTypeInt:
		result.SetString(strconv.FormatInt(o.ToInt(), 10))
	default:
		return fmt.Errorf("%s: unsupported type to string: %s", name, objType)
	}

	return nil
}
