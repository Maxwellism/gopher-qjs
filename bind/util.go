package bind

import (
	"encoding/json"
	"errors"
	quickjs "github.com/Maxwellism/gopher-qjs/wrap"
	"reflect"
)

func GoObjectToJsValue(goValue interface{}, ctx *quickjs.Context) quickjs.Value {
	value := reflect.ValueOf(goValue)
	switch value.Kind() {
	case reflect.Bool:
		res, _ := goValue.(bool)
		return ctx.Bool(res)
	case reflect.String:
		res, _ := goValue.(string)
		return ctx.String(res)
	case reflect.Int8:
		res, _ := goValue.(int8)
		return ctx.Int32(int32(res))
	case reflect.Int16:
		res, _ := goValue.(int16)
		return ctx.Int32(int32(res))
	case reflect.Int:
		res, _ := goValue.(int)
		return ctx.Int32(int32(res))
	case reflect.Uint:
		res, _ := goValue.(uint)
		return ctx.Int32(int32(res))
	case reflect.Int32:
		res, _ := goValue.(int32)
		return ctx.Int32(int32(res))
	case reflect.Uint32:
		res, _ := goValue.(uint32)
		return ctx.Int32(int32(res))
	case reflect.Int64:
		res, _ := goValue.(int64)
		return ctx.Int64(res)
	case reflect.Uint64:
		res, _ := goValue.(uint64)
		return ctx.Int64(int64(res))
	case reflect.Float32:
		res, _ := goValue.(float32)
		return ctx.Float64(float64(res))
	case reflect.Float64:
		res, _ := goValue.(float64)
		return ctx.Float64(float64(res))
	case reflect.Slice:
		data, err := json.Marshal(&goValue)
		if err != nil {
			panic(err)
		}
		return ctx.ParseJSON(string(data))
	case reflect.Struct:
		data, err := json.Marshal(&goValue)
		if err != nil {
			panic(err)
		}
		return ctx.ParseJSON(string(data))
	case reflect.Map:
		data, err := json.Marshal(&goValue)
		if err != nil {
			panic(err)
		}
		return ctx.ParseJSON(string(data))
	}
	return ctx.Undefined()
}

func JsValueToGoObject(expectArgumentType reflect.Type, jsValue quickjs.Value) (interface{}, error) {
	var resArg interface{}
	if goValue, err := jsValue.GetGoClassObject(); err == nil {
		resArg = goValue
	} else {
		switch expectArgumentType.Kind() {
		case reflect.Slice:
			if expectArgumentType.Elem().Kind() == reflect.Uint8 {
				if !jsValue.IsByteArray() {
					return nil, errors.New("this js object is not an ArrayBuffer")
				}
				byteData, err := jsValue.ToByteArray(uint(jsValue.ByteLen()))
				if err != nil {
					return nil, err
				}
				resArg = byteData
			} else {
				resArg = reflect.New(expectArgumentType).Elem().Interface()
				err := json.Unmarshal([]byte(jsValue.JSONStringify()), &resArg)
				if err != nil {
					return nil, err
				}
			}
		case reflect.Bool:
			if !jsValue.IsBool() {
				return nil, errors.New("this js object is not an bool")
			}
			resArg = jsValue.Bool()
		case reflect.String:
			if !jsValue.IsString() {
				return nil, errors.New("this js object is not an string")
			}
			resArg = jsValue.String()
		case reflect.Int8:
			resArg = int8(jsValue.Int32())
		case reflect.Int16:
			resArg = int16(jsValue.Int32())
		case reflect.Int:
			if !jsValue.IsNumber() {
				return nil, errors.New("this js object is not an number")
			}
			resArg = int(jsValue.Int32())
		case reflect.Uint:
			if !jsValue.IsNumber() {
				return nil, errors.New("this js object is not an number")
			}
			resArg = uint(jsValue.Int32())
		case reflect.Int32:
			if !jsValue.IsNumber() {
				return nil, errors.New("this js object is not an number")
			}
			resArg = jsValue.Int32()
		case reflect.Uint32:
			if !jsValue.IsNumber() {
				return nil, errors.New("this js object is not an number")
			}
			resArg = jsValue.Uint32()
		case reflect.Int64:
			if !jsValue.IsNumber() {
				return nil, errors.New("this js object is not an number")
			}
			resArg = jsValue.Int64()
		case reflect.Uint64:
			if !jsValue.IsNumber() {
				return nil, errors.New("this js object is not an number")
			}
			resArg = uint64(jsValue.Int64())
		case reflect.Float32:
			if !jsValue.IsNumber() {
				return nil, errors.New("this js object is not an number")
			}
			resArg = float32(jsValue.Float64())
		case reflect.Float64:
			if !jsValue.IsNumber() {
				return nil, errors.New("this js object is not an number")
			}
			resArg = jsValue.Float64()
		case reflect.Struct:
			resArg = reflect.New(expectArgumentType).Elem().Interface()
			err := json.Unmarshal([]byte(jsValue.JSONStringify()), &resArg)
			if err != nil {
				return nil, err
			}
		case reflect.Map:
			resArg = reflect.New(expectArgumentType).Elem().Interface()
			err := json.Unmarshal([]byte(jsValue.JSONStringify()), &resArg)
			if err != nil {
				return nil, err
			}
		}
	}
	return resArg, nil
	//
	//res := reflect.ValueOf(resArg)
	//return &res, nil
}
