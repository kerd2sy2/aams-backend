package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("بيانات الاعتماد غير صالحة")
	ErrUserNotFound       = errors.New("المستخدم غير موجود")
	ErrRoleNotFound       = errors.New("الدور غير موجود")
	ErrBranchNotFound     = errors.New("الفرع غير موجود")
	ErrOTPInvalid         = errors.New("رمز التحقق غير صالح أو منتهي الصلاحية")
	ErrUnauthorized       = errors.New("غير مصرح لك بالوصول")
)
