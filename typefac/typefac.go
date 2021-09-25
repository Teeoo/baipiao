package typefac

import (
	"errors"
	"reflect"
)

var funcMap = make(map[string]interface{})

// RegFunc 注册一个函数
func RegFunc(name string, fc interface{}) {
	funcMap[name] = fc
}

// Run 运行被注册的函数
func Run(name string, params ...interface{}) (result []reflect.Value, err error) {
	if _, ok := funcMap[name]; !ok {
		err = errors.New(name + " does not exist.")
		return
	}
	f := reflect.ValueOf(funcMap[name])
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is not adapted")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}

var typeRegistry = make(map[string]reflect.Type)

// RegisterType 注册类型
func RegisterType(typedNil interface{}) {
	t := reflect.TypeOf(typedNil)
	//typeRegistry[t.PkgPath()+"."+t.Name()] = t
	typeRegistry[t.String()] = t
}

// CreateInstance 传入被注册的类型和初值，创建一个对象
func CreateInstance(name string, input map[string]interface{}) interface{} {
	structObj := reflect.New(typeRegistry[name])
	structObjValue := structObj.Elem()
	for attrname, attrvalue := range input {
		// structObjValue.FieldByName(attrname).Set(reflect.ValueOf(attrvalue))
		assign(structObjValue.FieldByName(attrname), attrvalue)

	}
	return structObj.Interface()
}

func assign(v reflect.Value, input interface{}) {
	iv := reflect.ValueOf(input)
	if v.CanSet() {
		switch v.Kind() {
		case reflect.Bool:
			v.SetBool(iv.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v.SetInt(iv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v.SetUint(iv.Uint())
		case reflect.String:
			v.SetString(iv.String())
		case reflect.Slice: //TODO...
		case reflect.Map: //TODO...
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				structAttr := v.Field(i)
				if structAttr.IsValid() == false || structAttr.CanSet() == false {
					continue
				}
				assign(structAttr, iv.Field(i).Interface())
			}
		case reflect.Ptr:
			v.Set(iv)
		}
	}
}
