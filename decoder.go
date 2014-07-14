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
	case reflect.Bool:
		return decodeIntoBool(name, o, result)
	case reflect.Interface:
		// Interface is a bit weird. When we see an interface, we do
		// our best effort to determine the type, and put it into that.
		return decodeIntoInterface(name, o, result)
	case reflect.Int:
		return decodeIntoInt(name, o, result)
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

func decodeIntoBool(name string, o *Object, result reflect.Value) error {
	switch o.Type() {
	case ObjectTypeString:
		b, err := strconv.ParseBool(o.ToString())
		if err == nil {
			result.SetBool(b)
		} else {
			return fmt.Errorf("cannot parse '%s' as bool: %s", name, err)
		}
	default:
		result.SetBool(o.ToBool())
	}

	return nil
}

func decodeIntoInt(name string, o *Object, result reflect.Value) error {
	switch o.Type() {
	case ObjectTypeString:
		i, err := strconv.ParseInt(o.ToString(), 0, result.Type().Bits())
		if err == nil {
			result.SetInt(i)
		} else {
			return fmt.Errorf("cannot parse '%s' as int: %s", name, err)
		}
	default:
		result.SetInt(o.ToInt())
	}

	return nil
}

func decodeIntoInterface(name string, o *Object, result reflect.Value) error {
	var set reflect.Value
	redecode := true

	switch o.Type() {
	case ObjectTypeArray:
		redecode = false

		result := make([]interface{}, 0, int(o.Len()))

		iter := o.Iterate(true)
		defer iter.Close()
		for o := iter.Next(); o != nil; o = iter.Next() {
			raw := new(interface{})
			err := decode(name, o, reflect.Indirect(reflect.ValueOf(raw)))
			o.Close()

			if err != nil {
				return err
			}

			result = append(result, *raw)
		}

		set = reflect.ValueOf(result)
	case ObjectTypeBoolean:
		set = reflect.Indirect(reflect.New(reflect.TypeOf(o.ToBool())))
	case ObjectTypeInt:
		var result int
		set = reflect.Indirect(reflect.New(reflect.TypeOf(result)))
	case ObjectTypeObject:
		redecode = false

		result := make([]map[string]interface{}, 0, int(o.Len()))

		var err error
		outer := o.Iterate(false)
		defer outer.Close()
		for o := outer.Next(); o != nil; o = outer.Next() {
			m := make(map[string]interface{})
			inner := o.Iterate(true)
			for o2 := inner.Next(); o2 != nil; o2 = inner.Next() {
				var raw interface{}
				err = decode(name, o2, reflect.Indirect(reflect.ValueOf(&raw)))
				o2.Close()
				if err != nil {
					break
				}

				m[o2.Key()] = raw
			}
			inner.Close()
			o.Close()

			if err != nil {
				return err
			}

			result = append(result, m)
		}

		set = reflect.ValueOf(result)
	case ObjectTypeString:
		set = reflect.Indirect(reflect.New(reflect.TypeOf("")))
	default:
		return fmt.Errorf(
			"%s: unsupported type to interface: %s", name, o.Type())
	}

	if redecode {
		if err := decode(name, o, set); err != nil {
			return err
		}
	}

	result.Set(set)
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

			key := reflect.ValueOf(elem.Key())

			// The value we have to be decode
			val := reflect.Indirect(reflect.New(resultElemType))

			// If we have a pre-existing value in the map, use that
			oldVal := resultMap.MapIndex(key)
			if oldVal.IsValid() {
				val.Set(oldVal)
			}

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
	// Create the slice
	resultType := result.Type()
	resultElemType := resultType.Elem()
	resultSliceType := reflect.SliceOf(resultElemType)
	resultSlice := reflect.MakeSlice(
		resultSliceType, 0, int(o.Len()))

	// Determine how we're doing this
	expand := true
	switch o.Type() {
	case ObjectTypeObject:
		expand = false
	default:
		// Array or anything else: we expand values and take it all
	}

	i := 0
	iter := o.Iterate(expand)
	defer iter.Close()
	for elem := iter.Next(); elem != nil; elem = iter.Next() {
		val := reflect.Indirect(reflect.New(resultElemType))
		fieldName := fmt.Sprintf("%s[%d]", name, i)
		err := decode(fieldName, elem, val)
		elem.Close()
		if err != nil {
			return err
		}

		resultSlice = reflect.Append(resultSlice, val)

		i++
	}

	result.Set(resultSlice)

	return nil
}

func decodeIntoString(name string, o *Object, result reflect.Value) error {
	objType := o.Type()
	switch objType {
	case ObjectTypeBoolean:
		result.SetString(strconv.FormatBool(o.ToBool()))
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
					return fmt.Errorf(
						"%s: unsupported type to struct: %s",
						fieldType.Name, fieldKind)
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

	usedKeys := make(map[string]struct{})
	decodedFields := make([]string, 0, len(fields))
	decodedFieldsVal := make([]reflect.Value, 0)
	unusedKeysVal := make([]reflect.Value, 0)
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
		tagParts := strings.SplitN(tagValue, ",", 2)
		if len(tagParts) >= 2 {
			switch tagParts[1] {
			case "decodedFields":
				decodedFieldsVal = append(decodedFieldsVal, field)
				continue
			case "key":
				field.SetString(o.Key())
				continue
			case "object":
				// Increase the ref count
				o.Ref()

				// Sete the object
				field.Set(reflect.ValueOf(o))
				continue
			case "unusedKeys":
				unusedKeysVal = append(unusedKeysVal, field)
				continue
			}
		}

		if tagParts[0] != "" {
			fieldName = tagParts[0]
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
				// No key matching this field.
				continue
			}
		}

		// Track the used key
		usedKeys[elem.Key()] = struct{}{}

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

		decodedFields = append(decodedFields, fieldType.Name)
	}

	for _, v := range decodedFieldsVal {
		v.Set(reflect.ValueOf(decodedFields))
	}

	// If we want to know what keys are unused, compile thta
	if len(unusedKeysVal) > 0 {
		unusedKeys := make([]string, 0, int(o.Len())-len(usedKeys))

		iter := o.Iterate(true)
		defer iter.Close()
		for elem := iter.Next(); elem != nil; elem = iter.Next() {
			k := elem.Key()
			if _, ok := usedKeys[k]; !ok {
				unusedKeys = append(unusedKeys, k)
			}
			elem.Close()
		}

		if len(unusedKeys) == 0 {
			unusedKeys = nil
		}

		for _, v := range unusedKeysVal {
			v.Set(reflect.ValueOf(unusedKeys))
		}
	}

	return nil
}
