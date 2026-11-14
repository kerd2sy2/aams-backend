package inventory

import "errors"

var (
	ErrItemNameRequired      = errors.New("اسم الصنف مطلوب")
	ErrInvalidItemType       = errors.New("نوع الصنف يجب أن يكون 'oil' أو 'spare_part'")
	ErrItemNotFound          = errors.New("الصنف غير موجود")
	ErrVehicleNotFoundInInv  = errors.New("لم يتم العثور على الدباب المطلوب")
	ErrQuantityPositive      = errors.New("الكمية يجب أن تكون أكبر من صفر")
)
