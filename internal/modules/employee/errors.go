package employee

import "errors"

var (
	ErrEmployeeNotFound       = errors.New("الموظف غير موجود")
	ErrNationalIDRegistered   = errors.New("رقم الهوية الوطنية أو الإقامة مسجل مسبقاً")
	ErrNationalIDRequired     = errors.New("رقم الهوية الوطنية أو الإقامة مطلوب")
	ErrEmployeeNameRequired   = errors.New("اسم الموظف مطلوب")
	ErrInvalidCredentials     = errors.New("بيانات الدخول غير صحيحة")
	ErrOldPasswordIncorrect   = errors.New("كلمة المرور الحالية غير صحيحة")
	ErrBranchAccessDenied     = errors.New("لا تملك صلاحية الوصول لموظفي فرع آخر")
)
