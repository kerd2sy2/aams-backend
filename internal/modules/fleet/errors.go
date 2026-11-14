package fleet

import "errors"

var (
	ErrPlateRequired    = errors.New("رقم اللوحة مطلوب")
	ErrVehicleNotFound  = errors.New("المركبة غير موجودة")
	ErrPlateAlreadyUsed = errors.New("رقم اللوحة مسجل بالفعل")
	ErrInvalidPlate     = errors.New("رقم اللوحة غير صالح")
)
