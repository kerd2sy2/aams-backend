package ticket

import "errors"

var (
	ErrInvalidInput       = errors.New("بيانات الإدخال غير صالحة")
	ErrTicketNotFound     = errors.New("التذكرة غير موجودة")
	ErrUnauthorizedTicket = errors.New("غير مصرح بالوصول إلى هذه التذكرة")
)
