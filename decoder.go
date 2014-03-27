package libucl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const tagName = "libucl"

// Decode decodes a libucl object into a native Go structure.
func (o *Object) Decode(v interface{}) error {
	return decode("", o, reflect.ValueOf(v).Elem())
}

func decode(name string, o *Object, result reflect.Value) error {
	switch result.Kind() {
	case reflect.Map:
		return decodeIntoMap(name, o, result)
	case reflect.Ptr:
		return decodeIntoPtr(name, o, result)
	case reflect.Slice:
		return decodeIntoSlice(name, o, result)
	case reflect.String:
		return decodeIntoString(name, o, result)
	case reflect.Struct:
		return decodeIntoStruct(name, o, result)
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
	resultMap := result
	if result.IsNil() {
		resultMap = reflect.MakeMap(
			reflect.MapOf(resultKeyType, resultElemType))
	}

	outerIter := o.Iterate(false)
	defer outerIter.Close()
	for outer := outerIter.Next(); outer != nil; outer = outerIter.Next() {
		iter := outer.Iterate(true)
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
	}

	// Set the final result
	result.Set(resultMap)

	return nil
}

func decodeIntoPtr(name string, o *Object, result reflect.Value) error {
	// Create an element of the concrete (non pointer) type and decode
	// into that. Then set the value of the pointer to this type.
	resultType := result.Type()
	resultElemType := resultType.Elem()
	val := reflect.New(resultElemType)
	if err := decode(name, o, reflect.Indirect(val)); err != nil {
		return err
	}

	result.Set(val)
	return nil
}

func decodeIntoSlice(name string, o *Object, result reflect.Value) error {
	expand := true
	switch o.Type() {
	case ObjectTypeArray:
		// Okay.
	case ObjectTypeObject:
		expand = false
	default:
		return fmt.Errorf("%s: is not type array", name)
	}

	// Create the slice
	resultType := result.Type()
	resultElemType := resultType.Elem()
	resultSliceType := reflect.SliceOf(resultElemType)
	resultSlice := reflect.MakeSlice(resultSliceType, int(o.Len()), int(o.Len()))

	i := 0
	iter := o.Iterate(expand)
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

func decodeIntoStruct(name string, o *Object, result reflect.Value) error {
	// This slice will keep track of all the structs we'll be decoding.
	// There can be more than one struct if there are embedded structs
	// that are squashed.
	structs := make([]reflect.Value, 1, 5)
	structs[0] = result

	// Compile the list of all the fields that we're going to be decoding
	// from all the structs.
	fields := make(map[*reflect.StructField]reflect.Value)
	for len(structs) > 0 {
		structVal := structs[0]
		structs = structs[1:]

		structType := structVal.Type()
		for i := 0; i < structType.NumField(); i++ {
			fieldType := structType.Field(i)

			if fieldType.Anonymous {
				fieldKind := fieldType.Type.Kind()
				if fieldKind != reflect.Struct {
					return fmt.Errorf("%s: unsupported type: %s", fieldType.Name, fieldKind)
				}

				// We have an embedded field. We "squash" the fields down
				// if specified in the tag.
				squash := false
				tagParts := strings.Split(fieldType.Tag.Get(tagName), ",")
				for _, tag := range tagParts[1:] {
					if tag == "squash" {
						squash = true
						break
					}
				}

				if squash {
					structs = append(structs, result.FieldByName(fieldType.Name))
					continue
				}
			}

			// Normal struct field, store it away
			fields[&fieldType] = structVal.Field(i)
		}
	}

	for fieldType, field := range fields {
		if !field.IsValid() {
			// This should never happen
			panic("field is not valid")
		}

		// If we can't set the field, then it is unexported or something,
		// and we just continue onwards.
		if !field.CanSet() {
			continue
		}

		fieldName := fieldType.Name

		tagValue := fieldType.Tag.Get(tagName)
		tagValue = strings.SplitN(tagValue, ",", 2)[0]
		if tagValue != "" {
			fieldName = tagValue
		}

		elem := o.Get(fieldName)
		if elem == nil {
			// Do a slower search by iterating over each key and
			// doing case-insensitive search.
			iter := o.Iterate(true)
			for elem = iter.Next(); elem != nil; elem = iter.Next() {
				if strings.EqualFold(elem.Key(), fieldName) {
					break
				}

				elem.Close()
			}
			iter.Close()

			if elem == nil {
				// No key matching this field
				continue
			}
		}

		// If the name is empty string, then we're at the root, and we
		// don't dot-join the fields.
		if name != "" {
			fieldName = fmt.Sprintf("%s.%s", name, fieldName)
		}

		var err error
		if field.Kind() == reflect.Slice {
			err = decode(fieldName, elem, field)
		} else {
			iter := elem.Iterate(false)
			for {
				obj := iter.Next()
				if obj == nil {
					break
				}

				err = decode(fieldName, obj, field)
				obj.Close()
				if err != nil {
					break
				}
			}
			iter.Close()
		}
		elem.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
