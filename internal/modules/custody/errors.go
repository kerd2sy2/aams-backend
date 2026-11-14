package custody

import "errors"

var (
	ErrCustodyNotFound       = errors.New("يومية العهدة غير موجودة")
	ErrCustodyClosed         = errors.New("يومية العهدة مغلقة ولا يمكن تعديلها")
	ErrCustodyExpenseInvalid = errors.New("بيانات مصروف العهدة غير صالحة")
	ErrCustodyAmountPositive = errors.New("المبلغ المضاف للعهدة يجب أن يكون أكبر من صفر")
)
