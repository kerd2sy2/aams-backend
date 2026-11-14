package hr_legal

import "errors"

var (
	ErrInvestigationNotFound = errors.New("التحقيق غير موجود")
	ErrDocumentNotFound      = errors.New("المستند غير موجود")
	ErrBankAccountNotFound   = errors.New("الحساب البنكي غير موجود")
	ErrLeaveRequestNotFound  = errors.New("طلب الإجازة غير موجود")
	ErrViolationNotFound     = errors.New("المخالفة المرورية غير موجودة")
	ErrFuelLogNotFound       = errors.New("سجل الوقود غير موجود")
)
