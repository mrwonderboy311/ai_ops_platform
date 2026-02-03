// Package errors provides error types and utilities
package errors

import (
	"fmt"
)

// AppError represents an application error
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewError creates a new application error
func NewError(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap wraps an error with code and message
func Wrap(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// Predefined errors
var (
	ErrUsernameExists = NewError("USERNAME_EXISTS", "用户名已存在")
	ErrEmailExists    = NewError("EMAIL_EXISTS", "邮箱已被注册")
	ErrWeakPassword   = NewError("WEAK_PASSWORD", "密码强度不足")
	ErrInvalidEmail   = NewError("INVALID_EMAIL", "邮箱格式无效")
	ErrInvalidCredentials = NewError("INVALID_CREDENTIALS", "用户名或密码错误")
	ErrUnauthorized   = NewError("UNAUTHENTICATED", "未认证")
)
