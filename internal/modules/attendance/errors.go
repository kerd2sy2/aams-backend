package attendance

import "errors"

var (
	ErrEmployeeNotFound   = errors.New("الموظف غير موجود")
	ErrInvalidDate        = errors.New("التاريخ غير صالح")
	ErrUnauthorizedAccess = errors.New("غير مصرح بالوصول إلى سجل الحضور")
)
