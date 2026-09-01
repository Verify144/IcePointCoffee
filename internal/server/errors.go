package server

import (
	"encoding/json"
	"net/http"
)

// 错误码定义
const (
	// 通用错误
	ErrCodeSuccess         = 0
	ErrCodeBadRequest      = 400
	ErrCodeUnauthorized    = 401
	ErrCodeForbidden       = 403
	ErrCodeNotFound        = 404
	ErrCodeMethodNotAllowed = 405
	ErrCodeInternal        = 500
	ErrCodeServiceUnavailable = 503

	// 业务错误
	ErrCodeAIUnavailable   = 1001
	ErrCodeBuildFailed     = 1002
	ErrCodeCommandFailed   = 1003
	ErrCodePluginNotFound  = 1004
	ErrCodeToolNotFound    = 1005
	ErrCodeConfigInvalid   = 1006
	ErrCodeServerDisconnected = 1007
)

// APIError API 错误响应
type APIError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Hint    string      `json:"hint,omitempty"`
}

// ErrorResponse 标准错误响应包装
type ErrorResponse struct {
	Success bool     `json:"success"`
	Error   APIError `json:"error"`
}

// SuccessResponse 标准成功响应包装
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// NewError 创建错误响应
func NewError(code int, message, hint string) APIError {
	return APIError{
		Code:    code,
		Message: message,
		Hint:    hint,
	}
}

// SendError 发送错误响应
func SendError(w http.ResponseWriter, status int, code int, message, hint string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Error: APIError{
			Code:    code,
			Message: message,
			Hint:    hint,
		},
	})
}

// SendSuccess 发送成功响应
func SendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessResponse{
		Success: true,
		Data:    data,
	})
}
