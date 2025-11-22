package api

import (
	"github.com/gin-gonic/gin"
)

// APIResponse 统一API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(200, APIResponse{
		Success: true,
		Data:    data,
		Code:    200,
	})
}

// SuccessResponseWithCode 成功响应（自定义状态码）
func SuccessResponseWithCode(c *gin.Context, code int, data interface{}) {
	c.JSON(code, APIResponse{
		Success: true,
		Data:    data,
		Code:    code,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, APIResponse{
		Success: false,
		Code:    code,
		Message: message,
	})
}

// ErrorResponseWithData 错误响应（带数据）
func ErrorResponseWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, APIResponse{
		Success: false,
		Data:    data,
		Code:    code,
		Message: message,
	})
}
