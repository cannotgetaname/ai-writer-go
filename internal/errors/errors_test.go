package errors

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	// 测试无底层错误
	err := &AppError{Code: "TEST", Message: "测试错误"}
	expected := "TEST: 测试错误"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	// 测试带底层错误
	baseErr := errors.New("base error")
	err = &AppError{Code: "TEST", Message: "测试错误", Err: baseErr}
	expected = "TEST: 测试错误 (base error)"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestAppError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := &AppError{Code: "TEST", Message: "测试错误", Err: baseErr}

	unwrapped := err.Unwrap()
	if unwrapped != baseErr {
		t.Error("Unwrap should return base error")
	}

	// 测试无底层错误
	err = &AppError{Code: "TEST", Message: "测试错误"}
	if err.Unwrap() != nil {
		t.Error("Unwrap should return nil when no base error")
	}
}

func TestNew(t *testing.T) {
	err := New("CODE", "message")
	if err.Code != "CODE" {
		t.Errorf("Expected code 'CODE', got '%s'", err.Code)
	}
	if err.Message != "message" {
		t.Errorf("Expected message 'message', got '%s'", err.Message)
	}
}

func TestWrap(t *testing.T) {
	baseErr := errors.New("base")
	err := Wrap("CODE", "message", baseErr)
	if err.Err != baseErr {
		t.Error("Wrap should include base error")
	}
}

func TestIs(t *testing.T) {
	err := ErrBookNotFound
	if !Is(err, ErrBookNotFound) {
		t.Error("Is should return true for same error type")
	}
	if Is(err, ErrInvalidName) {
		t.Error("Is should return false for different error type")
	}
	if Is(nil, ErrBookNotFound) {
		t.Error("Is should return false for nil error")
	}
	if Is(errors.New("test"), ErrBookNotFound) {
		t.Error("Is should return false for non-AppError")
	}
}

func TestGetCode(t *testing.T) {
	if GetCode(ErrBookNotFound) != "BOOK_NOT_FOUND" {
		t.Error("GetCode should return correct code")
	}
	if GetCode(nil) != "" {
		t.Error("GetCode should return empty string for nil")
	}
	if GetCode(errors.New("test")) != "UNKNOWN" {
		t.Error("GetCode should return UNKNOWN for non-AppError")
	}
}

func TestGetMessage(t *testing.T) {
	if GetMessage(ErrBookNotFound) != "书籍不存在" {
		t.Error("GetMessage should return correct message")
	}
	if GetMessage(nil) != "" {
		t.Error("GetMessage should return empty string for nil")
	}
	if GetMessage(errors.New("test")) != "test" {
		t.Error("GetMessage should return error string for non-AppError")
	}
}