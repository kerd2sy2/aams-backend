package target

import "errors"

var (
	ErrIdentifierNotFound = errors.New("المعرف غير موجود")
	ErrBatchNotFound      = errors.New("دفعة الاستيراد غير موجودة")
	ErrAlertNotFound      = errors.New("التنبيه غير موجود")
	ErrInvalidExcelFile   = errors.New("ملف إكسيل غير صالح")
)
