package auth

import (
	"errors"
	"regexp"
)

var (
	ErrPasswordTooShort      = errors.New("密码至少需要 8 个字符")
	ErrPasswordMissingUpper  = errors.New("密码必须包含至少一个大写字母")
	ErrPasswordMissingLower  = errors.New("密码必须包含至少一个小写字母")
	ErrPasswordMissingDigit  = errors.New("密码必须包含至少一个数字")
	ErrInvalidUsernameFormat = errors.New("用户名只能包含字母、数字和下划线，3-50 个字符")
	ErrInvalidEmailFormat    = errors.New("邮箱格式无效")
)

// PasswordStrength validates password strength
func PasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasUpper {
		return ErrPasswordMissingUpper
	}
	if !hasLower {
		return ErrPasswordMissingLower
	}
	if !hasDigit {
		return ErrPasswordMissingDigit
	}

	return nil
}

// UsernameFormat validates username format
func UsernameFormat(username string) error {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,50}$`, username)
	if !matched {
		return ErrInvalidUsernameFormat
	}
	return nil
}

// EmailFormat validates email format
func EmailFormat(email string) error {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	if !matched {
		return ErrInvalidEmailFormat
	}
	return nil
}
