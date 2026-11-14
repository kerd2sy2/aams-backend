package work

import "errors"

var (
	ErrActiveSessionExists  = errors.New("يوجد شفت عمل نشط بالفعل حالياً")
	ErrNoActiveSession      = errors.New("لا يوجد شفت نشط قيد العمل حالياً أو تم إنهاؤه مسبقاً")
	ErrCrossBranchForbidden = errors.New("غير مصرح: الموظف يتبع فرعاً آخر")
)
