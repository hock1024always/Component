package api

import (
	"GoChatting/conf"
	"GoChatting/serializer"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
)

// 返回错误信息 ErrorResponse
// ErrorResponse函数用于处理错误响应
func ErrorResponse(err error) serializer.Response {
	// 判断错误是否为validator.ValidationErrors类型
	if ve, ok := err.(validator.ValidationErrors); ok {
		// 遍历错误
		for _, e := range ve {
			// 获取错误字段
			field := conf.T(fmt.Sprintf("Field.%s", e.Field))
			// 获取错误标签
			tag := conf.T(fmt.Sprintf("Tag.Valid.%s", e.Tag))
			// 返回错误响应
			return serializer.Response{
				Status: 400,
				Msg:    fmt.Sprintf("%s%s", field, tag),
				Error:  fmt.Sprint(err),
			}
		}
	}
	// 判断错误是否为json.UnmarshalTypeError类型
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		// 返回错误响应
		return serializer.Response{
			Status: 400,
			Msg:    "JSON类型不匹配",
			Error:  fmt.Sprint(err),
		}
	}

	// 返回错误响应
	return serializer.Response{
		Status: 400,
		Msg:    "参数错误",
		Error:  fmt.Sprint(err),
	}
}
