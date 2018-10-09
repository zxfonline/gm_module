// Copyright 2016 zxfonline@sina.com. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gm_module

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"

	"github.com/zxfonline/gerror"
	"github.com/zxfonline/golog"
)

var (
	h      reflect.Value
	logger *golog.Logger = golog.New("GMModule")
)

func CastParam(kind reflect.Kind, param string) (reflect.Value, error) {
	switch kind {
	case reflect.Int:
		val, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid int param:%s", param)
		}
		return reflect.ValueOf(int(val)), nil
	case reflect.Uint:
		val, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid uint param:%s", param)
		}
		return reflect.ValueOf(uint(val)), nil
	case reflect.Int8:
		val, err := strconv.ParseInt(param, 10, 8)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid int8 param:%s", param)
		}
		return reflect.ValueOf(int8(val)), nil
	case reflect.Uint8:
		val, err := strconv.ParseUint(param, 10, 8)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid uint8 param:%s", param)
		}
		return reflect.ValueOf(uint8(val)), nil
	case reflect.Int16:
		val, err := strconv.ParseInt(param, 10, 16)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid int16 param:%s", param)
		}
		return reflect.ValueOf(int16(val)), nil
	case reflect.Uint16:
		val, err := strconv.ParseUint(param, 10, 16)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid uint16 param:%s", param)
		}
		return reflect.ValueOf(uint16(val)), nil
	case reflect.Int32:
		val, err := strconv.ParseInt(param, 10, 32)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid int32 param:%s", param)
		}
		return reflect.ValueOf(int32(val)), nil
	case reflect.Uint32:
		val, err := strconv.ParseUint(param, 10, 32)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid uint32 param:%s", param)
		}
		return reflect.ValueOf(uint32(val)), nil
	case reflect.Int64:
		val, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid int64 param:%s", param)
		}
		return reflect.ValueOf(int64(val)), nil
	case reflect.Uint64:
		val, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid uint64 param:%s", param)
		}
		return reflect.ValueOf(uint64(val)), nil
	case reflect.Float32:
		val, err := strconv.ParseFloat(param, 32)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid float32 param:%s", param)
		}
		return reflect.ValueOf(float32(val)), nil
	case reflect.Float64:
		val, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid float64 param:%s", param)
		}
		return reflect.ValueOf(val), nil
	case reflect.Bool:
		val, err := strconv.ParseBool(param)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid bool param:%s", param)
		}
		return reflect.ValueOf(val), nil
	case reflect.String:
		return reflect.ValueOf(param), nil
	default:
		return reflect.Value{}, fmt.Errorf("invalid kind param:%s", param)
	}
}

func call(f ast.Expr) (string, interface{}, error) {
	funcname := ""
	params := make([]string, 0)
	unaryExprs := make(map[token.Pos]token.Token)
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BasicLit:
			val := x.Value
			if nVal, err := strconv.Unquote(val); err == nil {
				val = nVal
			}
			if op, present := unaryExprs[x.ValuePos]; present {
				switch op {
				case token.SUB:
					val = "-" + val
				}
			}
			params = append(params, val)
		case *ast.Ident:
			funcname = x.Name
		case *ast.UnaryExpr:
			v := n.(*ast.UnaryExpr)
			unaryExprs[v.X.Pos()] = v.Op
		}
		return true
	})
	fn := h.MethodByName(funcname)
	if fn.Kind() != reflect.Func {
		return funcname, nil, errors.New("cmd no found")
	}
	if fn.Type().NumIn() > len(params) {
		return funcname, nil, errors.New("cmd param not enough")
	}
	in := make([]reflect.Value, fn.Type().NumIn())
	for i := 0; i < fn.Type().NumIn(); i++ {
		val, err := CastParam(fn.Type().In(i).Kind(), params[i])
		if err == nil {
			in[i] = val
		} else {
			return funcname, nil, err
		}
	}
	ret := fn.Call(in)
	if len(ret) == 0 {
		return funcname, nil, nil
	}
	//默认支持两个返回参数(interface{},error)
	if len(ret) > 1 { //默认判定最后一个返回值为error类型
		if err, ok := ret[len(ret)-1].Interface().(error); ok {
			return funcname, err, nil
		}
	}
	//默认第一个返回值为结果
	return funcname, ret[0].Interface(), nil
}

//类似gerror.SysError 结构
type Response struct {
	Code    gerror.ErrorType `json:"ret"`
	Result  interface{}      `json:"result,omitempty"`
	Content string           `json:"msg"`
}

//注册gm工具类
func RegistHander(r reflect.Value) {
	h = r
}

func HandleCMD(exp string) (rs interface{}) {
	defer func() {
		if e := recover(); e != nil {
			logger.Warnf("\nHandleGM(exp=%s),error:%+v", exp, e)
			rs = gerror.NewError(gerror.SERVER_CMSG_ERROR, fmt.Sprintf("%v", e))
		}
	}()
	if f, perr := parser.ParseExpr(exp); perr == nil {
		funcn, resp, err := call(f)
		if err != nil {
			logger.Warnf("\n\tHandleGM[exp=%s]\n\tfunc:%s\n\terror:%+v", exp, funcn, err)
			switch err.(type) {
			case *gerror.SysError:
				rs = err.(*gerror.SysError)
			case error:
				rs = gerror.New(gerror.SERVER_CMSG_ERROR, err.(error))
			default:
				rs = gerror.NewError(gerror.SERVER_CMSG_ERROR, fmt.Sprintf("%v", err))
			}
		} else {
			logger.Infof("\n\tHandleGM[exp=%s]\n\tfunc:%s\n\trespond:%+v", exp, funcn, resp)
			switch resp.(type) {
			case *gerror.SysError:
				rs = resp
			case error:
				rs = gerror.New(gerror.SERVER_CMSG_ERROR, resp.(error))
			default:
				rs = Response{Code: gerror.OK, Result: resp}
			}
		}
	} else {
		logger.Warnf("\nHandleGM(exp=%s),error:%+v", exp, perr)
		rs = gerror.NewError(gerror.SERVER_CMSG_ERROR, "invalid cmd exp")
	}
	return
}
