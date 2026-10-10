package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const privacyPolicyHTML = `<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>سياسة الخصوصية | AAMS Privacy Policy</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      background-color: #f8fafc;
      color: #334155;
      line-height: 1.7;
      padding: 30px 15px;
    }
    .container {
      max-width: 800px;
      margin: 0 auto;
      background: #ffffff;
      padding: 40px;
      border-radius: 16px;
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.05);
      border: 1px solid #e2e8f0;
    }
    .badge {
      display: inline-block;
      background: #fff7ed;
      color: #ea580c;
      padding: 4px 12px;
      border-radius: 9999px;
      font-size: 0.85rem;
      font-weight: bold;
      margin-bottom: 12px;
    }
    h1 {
      color: #0f172a;
      font-size: 2rem;
      margin-bottom: 8px;
    }
    .date {
      color: #64748b;
      font-size: 0.85rem;
      margin-bottom: 24px;
      padding-bottom: 16px;
      border-bottom: 1px solid #e2e8f0;
    }
    h2 {
      color: #1e293b;
      font-size: 1.25rem;
      margin-top: 24px;
      margin-bottom: 12px;
      display: flex;
      align-items: center;
      gap: 8px;
    }
    h2::before {
      content: "";
      display: inline-block;
      width: 8px;
      height: 8px;
      background: #ea580c;
      border-radius: 50%;
    }
    p { margin-bottom: 14px; font-size: 0.95rem; }
    .card {
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      padding: 16px;
      border-radius: 10px;
      margin-bottom: 12px;
      font-size: 0.95rem;
    }
    .card strong { color: #0f172a; display: block; margin-bottom: 4px; }
    ul { margin: 12px 24px 16px 0; font-size: 0.95rem; }
    li { margin-bottom: 6px; }
    .contact-box {
      margin-top: 24px;
      background: #fff7ed;
      border: 1px solid #ffedd5;
      padding: 18px;
      border-radius: 12px;
    }
    .contact-box a { color: #ea580c; font-weight: bold; text-decoration: none; }
    .contact-box a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <div class="container">
    <div class="badge">تطبيق AAMS لإدارة الأسطول</div>
    <h1>سياسة الخصوصية (Privacy Policy)</h1>
    <div class="date">آخر تحديث: 2026-10-10</div>

    <section>
      <h2>1. مقدمة</h2>
      <p>نلتزم في <strong>AAMS</strong> بحماية خصوصية بيانات مستخدمينا وموظفينا الميدانيين. توضح سياسة الخصوصية هذه كيفية جمع البيانات، استخدامها، ومشاركتها عند استخدام تطبيق الهاتف المحمول <strong>AAMS</strong>.</p>
    </section>

    <section>
      <h2>2. البيانات والأذونات التي يجمعها ويطلبها التطبيق</h2>
      <div class="card">
        <strong>📍 بيانات الموقع الجغرافي (Location Data):</strong>
        يطلب التطبيق إذن الوصول للموقع الجغرافي لتحديد وتوثيق إحداثيات الحوادث الميدانية والبلاغات الطارئة ومتابعة المهام التشغيلية وتسجيل الحضور والانصراف.
      </div>
      <div class="card">
        <strong>📷 الكاميرا ومعرض الصور (Camera & Storage):</strong>
        تُستخدم الكاميرا لمسح لوحات المركبات وقراءة عدادات المسافات وتوثيق صور الصيانة وبلاغات الحوادث.
      </div>
      <div class="card">
        <strong>👤 معلومات الحساب (User & Account Info):</strong>
        نجمع بيانات التعريف الأساسية (الاسم، رقم الهاتف، الرقم الوظيفي) لإدارة الحسابات وتسجيل الدخول وصلاحيات التشغيل.
      </div>
      <div class="card">
        <strong>🔔 الإشعارات وسجلات النظام (Push Notifications & Logs):</strong>
        نستخدم الإشعارات لإرسال التنبيهات الفورية وبلاغات الطوارئ المهمة للمستخدمين.
      </div>
    </section>

    <section>
      <h2>3. الغرض من استخدام البيانات</h2>
      <ul>
        <li>إدارة ومتابعة العمليات الميدانية والأسطول.</li>
        <li>التحقق من لوحات المركبات وحالة الصيانة.</li>
        <li>توثيق ومعالجة بلاغات الحوادث ومواقعها بدقة.</li>
        <li>ضمان أمان النظام والتأكد من هوية السائقين والمشرفين.</li>
      </ul>
    </section>

    <section>
      <h2>4. أمان البيانات والمشاركة مع أطراف ثالثة</h2>
      <p>نحن <strong>لا نبيع أو نؤجر</strong> بيانات المستخدمين لأي أطراف خارجية أو إعلانية. يتم نقل وتشفير جميع البيانات باستخدام بروتوكول HTTPS/TLS الآمن. نستخدم خدمات موثوقة مثل Firebase Cloud Messaging لإرسال الإشعارات الميدانية فقط.</p>
    </section>

    <section>
      <h2>5. حذف البيانات والاحتفاظ بها</h2>
      <p>يمكن للمستخدم طلب حذف حسابه والبيانات المرتبطة به في أي وقت من خلال التواصل مع مسؤول النظام أو عبر مراسلتنا على بريد الدعم الفني.</p>
    </section>

    <section>
      <h2>6. معلومات التواصل</h2>
      <div class="contact-box">
        <p>إذا كان لديك أي استفسار حول سياسة الخصوصية، يرجى التواصل معنا عبر:</p>
        <p style="margin-top: 8px;">📧 البريد الإلكتروني: <a href="mailto:support@kerd2sy.com">support@kerd2sy.com</a></p>
        <p style="margin-top: 4px;">🏢 الجهة المسؤولة: <strong>فريق إدارة وتشغيل منصة AAMS</strong></p>
      </div>
    </section>
  </div>
</body>
</html>
`

func registerPrivacyRoutes(r *gin.Engine) {
	handler := func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, privacyPolicyHTML)
	}

	r.GET("/privacy", handler)
	r.GET("/privacy-policy", handler)
	r.GET("/api/v1/privacy", handler)
	r.GET("/api/v1/privacy-policy", handler)
}
