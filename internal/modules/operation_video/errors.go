package operation_video

import "errors"

var (
	ErrVideoNotFound      = errors.New("الفيديو غير موجود أو انتهت صلاحيته")
	ErrVideoFileMissing   = errors.New("ملف الفيديو غير موجود على السيرفر")
	ErrInvalidVideoFile   = errors.New("صيغة الفيديو غير مدعومة، يرجى رفع ملف فيديو صالح (.mp4, .mov, .webm)")
	ErrForbiddenVideoView = errors.New("غير مصرح: مشاهدة فيديوهات التوثيق مقتصرة على الإدارة والإشراف فقط")
)
