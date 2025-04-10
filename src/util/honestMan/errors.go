package honestMan

import (
	"fmt"
	"reflect"

	"nova/src/util/array"
	"nova/src/util/myError"
	"nova/src/util/operation"
)

type (
	ReadError  struct{ myError.MyError }
	WriteError struct{ myError.MyError }
)

var (
	ReadErr  ReadError
	WriteErr WriteError
)

func (*ReadError) New(msg string) myError.IMyError {
	return &ReadError{myError.MyError{Msg: array.NewDestruction("读取配置错误", msg).JoinWithoutEmpty("：")}}
}

func (*ReadError) Wrap(err error) myError.IMyError {
	return &ReadError{myError.MyError{Msg: fmt.Errorf("读取配置错误"+operation.Ternary(err != nil, "：%w", "%w"), err).Error()}}
}

func (*ReadError) Panic() myError.IMyError {
	return &ReadError{myError.MyError{Msg: "读取配置错误"}}
}

func (my *ReadError) Error() string { return my.Msg }

func (my *ReadError) Is(target error) bool { return reflect.DeepEqual(target, &ReadErr) }

func (*WriteError) New(msg string) myError.IMyError {
	return &WriteError{myError.MyError{Msg: array.NewDestruction("写入配置错误", msg).JoinWithoutEmpty("：")}}
}

func (*WriteError) Wrap(err error) myError.IMyError {
	return &WriteError{myError.MyError{Msg: fmt.Errorf("写入配置错误"+operation.Ternary(err != nil, "：%w", "%w"), err).Error()}}
}

func (*WriteError) Panic() myError.IMyError {
	return &WriteError{myError.MyError{Msg: "写入配置错误"}}
}

func (my *WriteError) Error() string { return my.Msg }

func (my *WriteError) Is(target error) bool { return reflect.DeepEqual(target, &WriteErr) }
