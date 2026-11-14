package maintenance

import "errors"

var (
	ErrLogNotFound            = errors.New("سجل الصيانة غير موجود")
	ErrRequestNotFound        = errors.New("طلب الصيانة غير موجود")
	ErrVehiclePlateRequired   = errors.New("رقم لوحة المركبة مطلوب")
	ErrIssueDescRequired      = errors.New("وصف العطل أو المشكلة مطلوب")
	ErrUnauthorizedPermission = errors.New("غير مصرح: لا تملك الصلاحية لتنفيذ هذا الإجراء")
)
