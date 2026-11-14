package system

import "errors"

var (
	ErrSettingNotFound      = errors.New("الإعداد المطلوب غير موجود")
	ErrBroadcastNotFound    = errors.New("التعميم غير موجود")
	ErrNotificationNotFound = errors.New("الإشعار غير موجود")
	ErrArchiveRestoreFailed = errors.New("فشل استعادة العنصر من الأرشيف")
	ErrArchiveDeleteFailed  = errors.New("فشل الحذف النهائي من الأرشيف")
)
