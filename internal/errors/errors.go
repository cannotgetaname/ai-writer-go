package errors

import (
	"fmt"
)

// AppError 应用错误类型
type AppError struct {
	Code    string
	Message string
	Err     error
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 支持错误解包
func (e *AppError) Unwrap() error {
	return e.Err
}

// 预定义错误
var (
	ErrBookNotFound     = &AppError{Code: "BOOK_NOT_FOUND", Message: "书籍不存在"}
	ErrBookExists       = &AppError{Code: "BOOK_EXISTS", Message: "书籍已存在"}
	ErrInvalidName      = &AppError{Code: "INVALID_NAME", Message: "名称不合法"}
	ErrChapterNotFound  = &AppError{Code: "CHAPTER_NOT_FOUND", Message: "章节不存在"}
	ErrUnauthorized     = &AppError{Code: "UNAUTHORIZED", Message: "未授权访问"}
	ErrInvalidInput     = &AppError{Code: "INVALID_INPUT", Message: "输入参数无效"}
	ErrInternal         = &AppError{Code: "INTERNAL_ERROR", Message: "内部错误"}
	ErrLLMFailed        = &AppError{Code: "LLM_FAILED", Message: "AI 模型调用失败"}
)

// New 创建新的应用错误
func New(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap 包装底层错误
func Wrap(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// Is 检查错误类型
func Is(err error, target *AppError) bool {
	if err == nil {
		return false
	}
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == target.Code
}

// GetCode 获取错误码
func GetCode(err error) string {
	if err == nil {
		return ""
	}
	appErr, ok := err.(*AppError)
	if !ok {
		return "UNKNOWN"
	}
	return appErr.Code
}

// GetMessage 获取错误消息
func GetMessage(err error) string {
	if err == nil {
		return ""
	}
	appErr, ok := err.(*AppError)
	if !ok {
		return err.Error()
	}
	return appErr.Message
}