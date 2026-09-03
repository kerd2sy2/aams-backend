package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/dto"
	"delivery-backend/internal/repository"
	"delivery-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

func isImageURL(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "/")
}

func (h *EmployeeHandler) saveImageIfBase64(value, category string) string {
	if value == "" {
		return ""
	}
	if isImageURL(value) {
		return value
	}
	imgUrl, err := h.storageService.SaveBase64Image(value, category)
	if err == nil {
		return imgUrl
	}
	return value
}

type AuthHandler struct {
	authService  service.AuthService
	auditService service.AuditService
}

func NewAuthHandler(authService service.AuthService, auditService service.AuditService) *AuthHandler {
	return &AuthHandler{authService: authService, auditService: auditService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "البيانات المدخلة غير صحيحة", "details": err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Audit Log
	_ = h.auditService.LogAction(c.Request.Context(), resp.Admin.Name, "تسجيل دخول", "تم تسجيل الدخول بنجاح", c.ClientIP(), nil)

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req dto.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "البيانات المدخلة غير صحيحة", "details": err.Error()})
		return
	}

	resp, err := h.authService.GoogleLogin(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Audit Log
	_ = h.auditService.LogAction(c.Request.Context(), resp.Admin.Name, "تسجيل دخول Google", "تم تسجيل الدخول بواسطة Google بنجاح", c.ClientIP(), nil)

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) LinkGoogle(c *gin.Context) {
	admin := getAdminContext(c)
	if admin == nil || admin.ID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}

	var req dto.GoogleLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "البريد الإلكتروني لحساب Google مطلوب", "details": err.Error()})
		return
	}

	updatedAdmin, err := h.authService.LinkGoogle(c.Request.Context(), admin.ID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = h.auditService.LogAction(c.Request.Context(), admin.Name, "ربط حساب Google", fmt.Sprintf("تم ربط الحساب بـ %s", req.Email), c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{
		"message":          "تم ربط حساب Google بنجاح",
		"google_email":     updatedAdmin.GoogleEmail,
		"google_avatar":    updatedAdmin.GoogleAvatar,
		"is_google_linked": updatedAdmin.GoogleEmail != "",
	})
}

func (h *AuthHandler) UnlinkGoogle(c *gin.Context) {
	admin := getAdminContext(c)
	if admin == nil || admin.ID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}

	if err := h.authService.UnlinkGoogle(c.Request.Context(), admin.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "تعذر إلغاء ربط الحساب"})
		return
	}

	_ = h.auditService.LogAction(c.Request.Context(), admin.Name, "إلغاء ربط Google", "تم إلغاء ربط حساب Google", c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{
		"message":          "تم إلغاء ربط حساب Google بنجاح",
		"is_google_linked": false,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رمز التحديث مطلوب"})
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Me returns the currently authenticated admin (from the auth context, which
// reads the branch from the database in real time). The frontend uses this to
// avoid relying on stale localStorage branch_id values.
func (h *AuthHandler) Me(c *gin.Context) {
	admin := getAdminContext(c)
	if admin == nil || admin.ID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "المستخدم غير موجود"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":               admin.ID,
		"name":             admin.Name,
		"email":            admin.Email,
		"username":         admin.Username,
		"phone":            admin.Phone,
		"role":             admin.Role,
		"role_id":          admin.RoleID,
		"google_email":     admin.GoogleEmail,
		"google_avatar":    admin.GoogleAvatar,
		"is_google_linked": admin.GoogleEmail != "",
		"permissions":      service.ResolveAdminPermissions(admin),
		"branch_id":        getBranchID(c),
	})
}


type EmployeeHandler struct {
	empService     service.EmployeeService
	storageService service.StorageService
	auditService   service.AuditService
	workRepo       repository.WorkRepository
}

func NewEmployeeHandler(empService service.EmployeeService, storageService service.StorageService, auditService service.AuditService, workRepo repository.WorkRepository) *EmployeeHandler {
	return &EmployeeHandler{empService: empService, storageService: storageService, auditService: auditService, workRepo: workRepo}
}

func (h *EmployeeHandler) Create(c *gin.Context) {
	var req dto.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى ملء جميع الحقول المطلوبة (اسم الموظف، الهوية الوطنية)", "details": err.Error()})
		return
	}

	// If logged-in admin is a supervisor (has branch_id), force employee to same branch
	if branchID, exists := c.Get("branch_id"); exists && branchID != nil {
		if bid, ok := branchID.(*uuid.UUID); ok && bid != nil {
			req.BranchID = bid
		} else if bid, ok := branchID.(uuid.UUID); ok {
			req.BranchID = &bid
		}
	}

	// Process base64 images if passed (skip when value is already a URL)
	req.PersonalImage = h.saveImageIfBase64(req.PersonalImage, "personal")
	req.NationalIDImage = h.saveImageIfBase64(req.NationalIDImage, "national_id")
	req.DrivingLicenseImage = h.saveImageIfBase64(req.DrivingLicenseImage, "license")

	emp, err := h.empService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "إنشاء موظف", "تم إضافة الموظف: "+emp.Name+" (ID: "+emp.ID.String()+")", c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusCreated, emp)
}

func (h *EmployeeHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	empID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	// Check branch access before modifying
	existingEmp, err := h.empService.GetByID(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}
	if !checkEmployeeBranchAccess(c, existingEmp) {
		return
	}

	var req dto.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "البيانات غير صالحة", "details": err.Error()})
		return
	}

	req.PersonalImage = h.saveImageIfBase64(req.PersonalImage, "personal")
	req.NationalIDImage = h.saveImageIfBase64(req.NationalIDImage, "national_id")
	req.DrivingLicenseImage = h.saveImageIfBase64(req.DrivingLicenseImage, "license")

	emp, err := h.empService.Update(c.Request.Context(), empID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "تعديل موظف", "تم تعديل بيانات الموظف: "+emp.Name, c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, emp)
}

func (h *EmployeeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	empID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	emp, err := h.empService.GetByID(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	if !checkEmployeeBranchAccess(c, emp) {
		return
	}

	empName := emp.Name

	if err := h.empService.Delete(c.Request.Context(), empID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حذف الموظف: " + err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "حذف موظف", "تم حذف الموظف: "+empName, c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الموظف بنجاح"})
}

func (h *EmployeeHandler) BatchSetOilChange(c *gin.Context) {
	var req dto.BatchOilSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	// Check branch access for each employee in the batch
	for _, entry := range req.Entries {
		empID, err := uuid.Parse(entry.EmployeeID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "معرف موظف غير صالح: " + entry.EmployeeID})
			return
		}
		emp, err := h.empService.GetByID(c.Request.Context(), empID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود: " + entry.EmployeeID})
			return
		}
		if !checkEmployeeBranchAccess(c, emp) {
			return
		}
	}

	if err := h.empService.BatchSetOilChange(c.Request.Context(), req.Entries); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "إعداد تغيير الزيت", fmt.Sprintf("تم تحديث آخر تغيير زيت لـ %d مندوب", len(req.Entries)), c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث بيانات تغيير الزيت بنجاح", "count": len(req.Entries)})
}

func (h *EmployeeHandler) GetWorking(c *gin.Context) {
	ids, err := h.workRepo.FindAllActiveEmployeeIDs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب المناديب العاملين"})
		return
	}

	if len(ids) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []domain.Employee{}, "total": 0})
		return
	}

	// Get branch filter for supervisors
	var branchID *uuid.UUID
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if b, ok := bid.(*uuid.UUID); ok && b != nil {
			branchID = b
		} else if b, ok := bid.(uuid.UUID); ok {
			branchID = &b
		}
	}

	emps, err := h.empService.GetAll(c.Request.Context(), dto.EmployeeFilter{Limit: 1000})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب بيانات المناديب"})
		return
	}

	// Filter employees to only those with active sessions
	idSet := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}

	var workingEmps []domain.Employee
	allEmps, ok := emps.Data.([]domain.Employee)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في معالجة بيانات الموظفين"})
		return
	}
	for _, emp := range allEmps {
		if !idSet[emp.ID] {
			continue
		}
		// Filter by branch for supervisors; admins (branchID == nil) see all
		if branchID != nil && (emp.BranchID == nil || *emp.BranchID != *branchID) {
			continue
		}
		workingEmps = append(workingEmps, emp)
	}

	c.JSON(http.StatusOK, gin.H{"data": workingEmps, "total": len(workingEmps)})
}

func (h *EmployeeHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	empID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	emp, err := h.empService.GetByID(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	if !checkEmployeeBranchAccess(c, emp) {
		return
	}

	c.JSON(http.StatusOK, emp)
}

func (h *EmployeeHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	var branchID *uuid.UUID
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if b, ok := bid.(*uuid.UUID); ok {
			branchID = b
		} else if b, ok := bid.(uuid.UUID); ok {
			branchID = &b
		}
	}

	results, err := h.empService.Search(c.Request.Context(), q, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *EmployeeHandler) GetAll(c *gin.Context) {
	var filter dto.EmployeeFilter
	filter.Search = c.Query("search")
	filter.ApplicationID = c.Query("application_id")
	filter.SortBy = c.DefaultQuery("sort_by", "created_at")
	filter.Order = c.DefaultQuery("order", "desc")

	// Auto-filter by branch for supervisors
	if branchID, exists := c.Get("branch_id"); exists && branchID != nil {
		if bid, ok := branchID.(*uuid.UUID); ok && bid != nil {
			filter.BranchID = bid
		} else if bid, ok := branchID.(uuid.UUID); ok {
			filter.BranchID = &bid
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filter.Page = page
	filter.Limit = limit

	resp, err := h.empService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *EmployeeHandler) GetBarcode(c *gin.Context) {
	idStr := c.Param("id")
	empID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	emp, err := h.empService.GetByID(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}
	if !checkEmployeeBranchAccess(c, emp) {
		return
	}

	barcodeData, err := h.empService.GetBarcode(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"barcode": barcodeData})
}

func (h *EmployeeHandler) GetQRCode(c *gin.Context) {
	idStr := c.Param("id")
	empID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	emp, err := h.empService.GetByID(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}
	if !checkEmployeeBranchAccess(c, emp) {
		return
	}

	qrData, err := h.empService.GetQRCode(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"qr_code": qrData})
}

func (h *EmployeeHandler) GetPrintCard(c *gin.Context) {
	idStr := c.Param("id")
	empID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	emp, err := h.empService.GetByID(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	if !checkEmployeeBranchAccess(c, emp) {
		return
	}

	barcodeData, _ := h.empService.GetBarcode(c.Request.Context(), empID)
	qrData, _ := h.empService.GetQRCode(c.Request.Context(), empID)

	c.JSON(http.StatusOK, gin.H{
		"employee": emp,
		"barcode":  barcodeData,
		"qr_code":  qrData,
	})
}

func (h *EmployeeHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "لم يتم اختيار ملف"})
		return
	}

	category := c.DefaultPostForm("category", "personal")
	imgUrl, err := h.storageService.SaveImage(file, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل رفع الصورة: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": imgUrl})
}

func (h *EmployeeHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "لم يتم اختيار ملف"})
		return
	}

	// 50 MB limit
	if file.Size > 50<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "حجم الملف كبير جداً. الحد الأقصى 50 ميجابايت"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
		".pdf": true,
		".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true,
		".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "نوع الملف غير مسموح به"})
		return
	}

	_ = os.MkdirAll("uploads/documents", 0755)
	filename := fmt.Sprintf("doc_%d%s", time.Now().UnixNano(), ext)
	dstPath := filepath.Join("uploads", "documents", filename)

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل فتح الملف"})
		return
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حفظ الملف"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل نسخ الملف"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": fmt.Sprintf("/uploads/documents/%s", filename)})
}

type WorkHandler struct {
	workService       service.WorkService
	auditService      service.AuditService
	attendanceService service.AttendanceService
	empRepo           repository.EmployeeRepository
}

func NewWorkHandler(workService service.WorkService, auditService service.AuditService, attendanceService service.AttendanceService, empRepo repository.EmployeeRepository) *WorkHandler {
	return &WorkHandler{workService: workService, auditService: auditService, attendanceService: attendanceService, empRepo: empRepo}
}

func (h *WorkHandler) checkWorkEmployeeBranch(c *gin.Context, empID uuid.UUID) (*domain.Employee, bool) {
	emp, err := h.empRepo.FindByID(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return nil, false
	}
	if !checkEmployeeBranchAccess(c, emp) {
		return nil, false
	}
	return emp, true
}

func (h *WorkHandler) StartWork(c *gin.Context) {
	var req dto.StartWorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة. يرجى اختيار الموظف وإدخال قراءة عداد البداية", "details": err.Error()})
		return
	}

	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}
	if _, ok := h.checkWorkEmployeeBranch(c, empID); !ok {
		return
	}

	session, err := h.workService.StartWork(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	empName := "الموظف"
	if session.Employee != nil {
		empName = session.Employee.Name
	}
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "بدء شفت عمل", "تم بدء شفت العمل للموظف: "+empName+" - عداد البداية: "+strconv.FormatFloat(session.StartKM, 'f', 2, 64)+" كم", c.ClientIP(), getBranchID(c))

	// Auto-mark attendance as present
	today := time.Now().Format("2006-01-02")
	if session.EmployeeID != nil {
		_, _ = h.attendanceService.ToggleAttendance(c.Request.Context(), uuid.Nil, *session.EmployeeID, today, "present", "بداية شفت تلقائي")
	}

	c.JSON(http.StatusCreated, session)
}

func (h *WorkHandler) EndWork(c *gin.Context) {
	var req dto.EndWorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة. يرجى التأكد من إدخال قراءة عداد النهاية بشكل صحيح", "details": err.Error()})
		return
	}

	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}
	if _, ok := h.checkWorkEmployeeBranch(c, empID); !ok {
		return
	}

	session, err := h.workService.EndWork(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	empName := "الموظف"
	if session.Employee != nil {
		empName = session.Employee.Name
	}
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "إنهاء شفت عمل", "تم إنهاء شفت العمل للموظف: "+empName+" - المسافة المقطوعة: "+strconv.FormatFloat(session.Distance, 'f', 2, 64)+" كم - عدد الطلبات: "+strconv.Itoa(session.OrdersCount), c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) UpdateWorkSession(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الشفت غير صالح"})
		return
	}

	var req dto.UpdateWorkSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	// Verify branch access: check against the session's employee
	session, err := h.workService.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الشفت غير موجود"})
		return
	}
	if session.EmployeeID != nil {
		if _, ok := h.checkWorkEmployeeBranch(c, *session.EmployeeID); !ok {
			return
		}
	}

	session, err = h.workService.UpdateWorkSession(c.Request.Context(), sessionID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	empName := "الموظف"
	if session.Employee != nil {
		empName = session.Employee.Name
	}
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "تعديل شفت عمل",
		fmt.Sprintf("تم تعديل شفت العمل للموظف: %s - عداد البداية: %.2f كم - عداد النهاية: %.2f كم - المسافة: %.2f كم - الطلبات: %d - الوقود: %.2f ر.س",
			empName, session.StartKM, session.EndKM, session.Distance, session.OrdersCount, session.FuelCost), c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) ReviewWorkSession(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الشفت غير صالح"})
		return
	}

	var req dto.ReviewWorkSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	var reviewerID *uuid.UUID
	if adminIDStr := c.GetString("admin_id"); adminIDStr != "" {
		if parsed, err := uuid.Parse(adminIDStr); err == nil {
			reviewerID = &parsed
		}
	}

	adminName := c.GetString("admin_name")
	session, err := h.workService.ReviewSession(c.Request.Context(), sessionID, req, reviewerID, adminName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "مراجعة وتدقيق شفت",
		fmt.Sprintf("تم تدقيق عدادات الشفت ID: %s (حالة المراجعة: %v) - ملاحظات: %s", sessionIDStr, req.IsReviewed, req.ReviewNotes), c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) GetSessionByID(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الشفت غير صالح"})
		return
	}

	session, err := h.workService.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "جلسة العمل غير موجودة"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) GetActiveSession(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	if _, ok := h.checkWorkEmployeeBranch(c, empID); !ok {
		return
	}

	session, err := h.workService.GetActiveSession(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "لا يوجد شفت نشط حالياً"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) GetLastKM(c *gin.Context) {
	empIDStr := strings.TrimSpace(c.Query("employee_id"))
	motorcycleNumber := strings.TrimSpace(c.Query("motorcycle_number"))
	if motorcycleNumber == "" {
		motorcycleNumber = strings.TrimSpace(c.Query("plate"))
	}

	var empID uuid.UUID
	if empIDStr != "" {
		var err error
		empID, err = uuid.Parse(empIDStr)
		if err == nil {
			if _, ok := h.checkWorkEmployeeBranch(c, empID); !ok {
				return
			}
		}
	}

	lastEndKM, lastStartKM, err := h.workService.GetLastSessionOrVehicleKM(c.Request.Context(), empID, motorcycleNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "لا توجد قراءة سابقة مسجلة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"last_end_km":   lastEndKM,
		"last_start_km": lastStartKM,
	})
}

func (h *WorkHandler) TodayCount(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	if _, ok := h.checkWorkEmployeeBranch(c, empID); !ok {
		return
	}

	count, err := h.workService.CountTodaySessions(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب البيانات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"today_count": count})
}

type DashboardHandler struct {
	dashService service.DashboardService
}

func NewDashboardHandler(dashService service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashService: dashService}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	// Get branch_id from context for filtering (from DB, not JWT)
	var branchID *uuid.UUID
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if b, ok := bid.(*uuid.UUID); ok {
			branchID = b
		} else if b, ok := bid.(uuid.UUID); ok {
			branchID = &b
		}
	}

	stats, err := h.dashService.GetStats(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب إحصائيات لوحة التحكم: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

type ReportHandler struct {
	reportService service.ReportService
}

func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) GetReports(c *gin.Context) {
	var filter dto.ReportFilter
	filter.StartDate = c.Query("start_date")
	filter.EndDate = c.Query("end_date")
	filter.EmployeeID = c.Query("employee_id")
	filter.ApplicationID = c.Query("application_id")

	// Auto-filter by branch for supervisors
	if branchID, exists := c.Get("branch_id"); exists && branchID != nil {
		if b, ok := branchID.(*uuid.UUID); ok {
			filter.BranchID = b
		} else if b, ok := branchID.(uuid.UUID); ok {
			filter.BranchID = &b
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	filter.Page = page
	filter.Limit = limit

	reports, total, err := h.reportService.GetReports(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := int((total + int64(filter.Limit) - 1) / int64(filter.Limit))

	c.JSON(http.StatusOK, gin.H{
		"data":        reports,
		"total":       total,
		"page":        filter.Page,
		"limit":       filter.Limit,
		"total_pages": totalPages,
	})
}

func (h *ReportHandler) ExportReports(c *gin.Context) {
	var filter dto.ReportFilter
	filter.StartDate = c.Query("start_date")
	filter.EndDate = c.Query("end_date")
	filter.EmployeeID = c.Query("employee_id")
	filter.ApplicationID = c.Query("application_id")

	if branchID, exists := c.Get("branch_id"); exists && branchID != nil {
		if b, ok := branchID.(*uuid.UUID); ok {
			filter.BranchID = b
		} else if b, ok := branchID.(uuid.UUID); ok {
			filter.BranchID = &b
		}
	}

	filter.Page = 1
	filter.Limit = 10000

	reports, _, err := h.reportService.GetReports(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	appMap := map[string]string{
		"ninja":  "نينجا",
		"keeta":  "كيتا",
		"hunger": "هنجر",
		"toyou":  "تو يو",
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "تقارير الشفتات"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")

	// Title
	f.MergeCell(sheetName, "A1", "J1")
	f.SetCellValue(sheetName, "A1", fmt.Sprintf("تقرير الشفتات - %s", time.Now().Format("2006-01-02")))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	// Headers
	headers := []string{"اسم المندوب", "التاريخ", "وقت البداية", "وقت النهاية", "مدة العمل", "المسافة (كم)", "الطلبات", "الوقود (ر.س)", "التطبيق", "الملاحظات"}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E3A5F"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		cell := col + "3"
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Data rows
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})

	for i, r := range reports {
		rowNum := i + 4
		endTimeStr := ""
		if r.EndTime != nil {
			endTimeStr = r.EndTime.Format("15:04")
		}
		appName := appMap[r.ApplicationType]
		if appName == "" {
			appName = r.ApplicationType
		}

		vals := []interface{}{
			r.EmployeeName,
			r.StartTime.Format("2006-01-02"),
			r.StartTime.Format("15:04"),
			endTimeStr,
			r.WorkingDuration,
			r.Distance,
			r.OrdersCount,
			r.FuelCost,
			appName,
			r.Notes,
		}
		for j, v := range vals {
			col, _ := excelize.ColumnNumberToName(j + 1)
			cell := fmt.Sprintf("%s%d", col, rowNum)
			f.SetCellValue(sheetName, cell, v)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}
	}

	// Column widths
	f.SetColWidth(sheetName, "A", "A", 22)
	f.SetColWidth(sheetName, "B", "B", 14)
	f.SetColWidth(sheetName, "C", "C", 12)
	f.SetColWidth(sheetName, "D", "D", 12)
	f.SetColWidth(sheetName, "E", "E", 12)
	f.SetColWidth(sheetName, "F", "F", 14)
	f.SetColWidth(sheetName, "G", "G", 10)
	f.SetColWidth(sheetName, "H", "H", 13)
	f.SetColWidth(sheetName, "I", "I", 12)
	f.SetColWidth(sheetName, "J", "J", 20)

	f.SetSheetView(sheetName, -1, &excelize.ViewOptions{RightToLeft: boolPtr(true)})
	idx, _ := f.GetSheetIndex(sheetName)
	f.SetActiveSheet(idx)

	fileName := "تقرير_الشفتات.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	c.Header("Content-Transfer-Encoding", "binary")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل إنشاء ملف Excel"})
	}
}

func (h *ReportHandler) GetDailyReport(c *gin.Context) {
	var branchID *uuid.UUID
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if b, ok := bid.(*uuid.UUID); ok {
			branchID = b
		} else if b, ok := bid.(uuid.UUID); ok {
			branchID = &b
		}
	}

	// Accept optional date parameter (YYYY-MM-DD), defaults to today
	dateStr := c.Query("date")
	targetDate := time.Now()
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			targetDate = parsed
		}
	}

	report, err := h.reportService.GetDailyReport(c.Request.Context(), branchID, targetDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) ExportDailyReport(c *gin.Context) {
	var branchID *uuid.UUID
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if b, ok := bid.(*uuid.UUID); ok {
			branchID = b
		} else if b, ok := bid.(uuid.UUID); ok {
			branchID = &b
		}
	}

	dateStr := c.Query("date")
	targetDate := time.Now()
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			targetDate = parsed
		}
	}

	report, err := h.reportService.GetDailyReport(c.Request.Context(), branchID, targetDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "التقرير اليومي"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")

	// Title row (merged)
	f.MergeCell(sheetName, "A1", "F1")
	titleCell := "A1"
	f.SetCellValue(sheetName, titleCell, "التقرير اليومي - "+targetDate.Format("2006-01-02"))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheetName, titleCell, titleCell, titleStyle)

	// Headers
	headers := []string{"اسم المندوب", "التطبيق", "الكيلومترات", "الطلبات", "الوقود", "عدد الشفتات"}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E3A5F"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "#FFFFFF"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		cell := col + "3"
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Data rows
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})

	for i, row := range report.Rows {
		rowNum := i + 4
		vals := []interface{}{
			row.EmployeeName,
			row.AppName,
			row.TotalKM,
			row.TotalOrders,
			row.TotalFuel,
			row.SessionsCount,
		}
		for j, v := range vals {
			col, _ := excelize.ColumnNumberToName(j + 1)
			cell := fmt.Sprintf("%s%d", col, rowNum)
			f.SetCellValue(sheetName, cell, v)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}
	}

	// Summary row
	summaryRow := len(report.Rows) + 4
	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E8F0FE"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
	})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "الإجمالي")
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), report.TotalKM)
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", summaryRow), report.TotalOrders)
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", summaryRow), report.TotalFuel)
	f.SetCellValue(sheetName, fmt.Sprintf("F%d", summaryRow), len(report.Rows))
	for col := 0; col < 6; col++ {
		colName, _ := excelize.ColumnNumberToName(col + 1)
		cell := fmt.Sprintf("%s%d", colName, summaryRow)
		f.SetCellStyle(sheetName, cell, cell, summaryStyle)
	}

	// Column widths
	f.SetColWidth(sheetName, "A", "A", 22) // اسم المندوب
	f.SetColWidth(sheetName, "B", "B", 16) // التطبيق
	f.SetColWidth(sheetName, "C", "C", 14) // الكيلومترات
	f.SetColWidth(sheetName, "D", "D", 10) // الطلبات
	f.SetColWidth(sheetName, "E", "E", 12) // الوقود
	f.SetColWidth(sheetName, "F", "F", 12) // عدد الشفتات

	// Set sheet RTL
	f.SetSheetView(sheetName, -1, &excelize.ViewOptions{RightToLeft: boolPtr(true)})

	idx, _ := f.GetSheetIndex(sheetName)
	f.SetActiveSheet(idx)

	fileName := fmt.Sprintf("التقرير_اليومي_%s.xlsx", targetDate.Format("2006-01-02"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	c.Header("Content-Transfer-Encoding", "binary")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل إنشاء ملف Excel"})
	}
}

func boolPtr(b bool) *bool { return &b }

type RoleHandler struct {
	roleService  service.RoleService
	auditService service.AuditService
}

func NewRoleHandler(roleService service.RoleService, auditService service.AuditService) *RoleHandler {
	return &RoleHandler{roleService: roleService, auditService: auditService}
}

func (h *RoleHandler) GetAll(c *gin.Context) {
	roles, err := h.roleService.GetAllRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب قائمة الأدوار"})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الدور غير صالح"})
		return
	}
	role, err := h.roleService.GetRoleByID(c.Request.Context(), roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *RoleHandler) Create(c *gin.Context) {
	if !checkAdminOnly(c) {
		return
	}
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى ملء جميع الحقول المطلوبة", "details": err.Error()})
		return
	}

	role, err := h.roleService.CreateRole(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "إنشاء دور جديد", "تم إنشاء الدور: "+role.DisplayName+" ("+role.Name+")", c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusCreated, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	if !checkAdminOnly(c) {
		return
	}
	idStr := c.Param("id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الدور غير صالح"})
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "البيانات غير صالحة", "details": err.Error()})
		return
	}

	role, err := h.roleService.UpdateRole(c.Request.Context(), roleID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "تعديل دور", "تم تعديل الدور: "+role.DisplayName, c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	if !checkAdminOnly(c) {
		return
	}
	idStr := c.Param("id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الدور غير صالح"})
		return
	}

	if err := h.roleService.DeleteRole(c.Request.Context(), roleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "حذف دور", "تم حذف الدور: "+idStr, c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الدور بنجاح"})
}

func (h *RoleHandler) GetPermissions(c *gin.Context) {
	catalog := h.roleService.GetPermissionsCatalog()
	c.JSON(http.StatusOK, catalog)
}

type AdminHandler struct {
	adminService service.AdminService
	auditService service.AuditService
	adminRepo    repository.AdminRepository
}

func NewAdminHandler(adminService service.AdminService, auditService service.AuditService, adminRepo repository.AdminRepository) *AdminHandler {
	return &AdminHandler{adminService: adminService, auditService: auditService, adminRepo: adminRepo}
}

func (h *AdminHandler) Create(c *gin.Context) {
	if !checkAdminOnly(c) {
		return
	}
	var req dto.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى ملء جميع الحقول المطلوبة", "details": err.Error()})
		return
	}

	admin, err := h.adminService.CreateAdmin(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "إضافة مدير", "تم إضافة المدير: "+admin.Name+" (اليوزر: "+admin.Username+")", c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusCreated, admin)
}

func (h *AdminHandler) Update(c *gin.Context) {
	if !checkAdminOnly(c) {
		return
	}
	idStr := c.Param("id")
	adminID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المدير غير صالح"})
		return
	}

	// Check branch access: supervisors can only edit admins in their own branch
	targetAdmin, err := h.adminRepo.FindByID(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "المدير غير موجود"})
		return
	}
	adminBranchID := getBranchID(c)
	if adminBranchID != nil {
		if targetAdmin.BranchID == nil || *targetAdmin.BranchID != *adminBranchID {
			c.JSON(http.StatusForbidden, gin.H{"error": "ليس لديك صلاحية تعديل هذا المدير - الفرع غير مطابق"})
			return
		}
	}

	var req dto.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "البيانات غير صالحة", "details": err.Error()})
		return
	}

	admin, err := h.adminService.UpdateAdmin(c.Request.Context(), adminID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "تعديل مدير", "تم تعديل بيانات المدير: "+admin.Name, c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, admin)
}

func (h *AdminHandler) Delete(c *gin.Context) {
	if !checkAdminOnly(c) {
		return
	}
	idStr := c.Param("id")
	adminID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المدير غير صالح"})
		return
	}

	// Prevent self-deletion
	currentAdminID, _ := c.Get("admin_id")
	adminUUID, ok := currentAdminID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في التحقق من صلاحيات المستخدم"})
		return
	}
	if currentAdminID != nil && adminUUID.String() == idStr {
		c.JSON(http.StatusBadRequest, gin.H{"error": "لا يمكنك حذف حسابك الحالي"})
		return
	}

	// Check branch access: supervisors can only delete admins in their own branch
	targetAdmin, err := h.adminRepo.FindByID(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "المدير غير موجود"})
		return
	}
	adminBranchID := getBranchID(c)
	if adminBranchID != nil {
		if targetAdmin.BranchID == nil || *targetAdmin.BranchID != *adminBranchID {
			c.JSON(http.StatusForbidden, gin.H{"error": "ليس لديك صلاحية حذف هذا المدير - الفرع غير مطابق"})
			return
		}
	}

	if err := h.adminService.DeleteAdmin(c.Request.Context(), adminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "حذف مدير", "تم حذف مدير", c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف المدير بنجاح"})
}

func (h *AdminHandler) GetAll(c *gin.Context) {
	branchID := getBranchID(c)
	admins, err := h.adminService.GetAllAdmins(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type AdminWithPermissions struct {
		domain.Admin
		PermissionsList []string `json:"permissions_list"`
	}
	var res []AdminWithPermissions
	for _, a := range admins {
		res = append(res, AdminWithPermissions{
			Admin:           a,
			PermissionsList: service.ResolveAdminPermissions(&a),
		})
	}

	c.JSON(http.StatusOK, res)
}


func (h *AdminHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى إدخال كلمة المرور القديمة والجديدة", "details": err.Error()})
		return
	}

	adminIDVal, exists := c.Get("admin_id")
	if !exists || adminIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}

	if err := h.adminService.ChangePassword(c.Request.Context(), adminIDVal.(uuid.UUID), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "تغيير كلمة المرور", "تم تغيير كلمة المرور بنجاح", c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم تغيير كلمة المرور بنجاح"})
}

type AuditHandler struct {
	auditService service.AuditService
}

func NewAuditHandler(auditService service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

func (h *AuditHandler) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	branchID := getBranchID(c)

	logs, total, err := h.auditService.GetLogs(c.Request.Context(), branchID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func (h *AuditHandler) DeleteLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.auditService.DeleteLog(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف سجل العملية بنجاح"})
}

func (h *AuditHandler) BulkDeleteLogs(c *gin.Context) {
	var req struct {
		IDs []uuid.UUID `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الحذف غير صالحة", "details": err.Error()})
		return
	}

	if err := h.auditService.BulkDeleteLogs(c.Request.Context(), req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("تم حذف %d سجل بنجاح", len(req.IDs))})
}

func (h *AuditHandler) ClearLogs(c *gin.Context) {
	branchID := getBranchID(c)

	if err := h.auditService.ClearLogs(c.Request.Context(), branchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تفريغ وحذف سجل العمليات بالكامل من قاعدة البيانات"})
}


// --- Attendance Handler ---

type AttendanceHandler struct {
	attendanceService service.AttendanceService
	auditService      service.AuditService
	empRepo           repository.EmployeeRepository
}

func NewAttendanceHandler(attendanceService service.AttendanceService, auditService service.AuditService, empRepo repository.EmployeeRepository) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceService: attendanceService,
		auditService:      auditService,
		empRepo:           empRepo,
	}
}

func (h *AttendanceHandler) GetAttendance(c *gin.Context) {
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	// Auto-filter by branch for supervisors
	branchID := getBranchID(c)

	attendance, err := h.attendanceService.GetAttendance(c.Request.Context(), date, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("فشل في جلب بيانات الحضور: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date": date,
		"data": attendance,
	})
}

func (h *AttendanceHandler) ToggleAttendance(c *gin.Context) {
	adminName := c.GetString("admin_name")
	adminBranchID := getBranchID(c)

	employeeIDStr := c.Param("employee_id")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	// Verify branch access
	emp, err := h.empRepo.FindByID(c.Request.Context(), employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}
	if !checkEmployeeBranchAccess(c, emp) {
		return
	}

	var req struct {
		Date   string `json:"date" binding:"required"`
		Status string `json:"status" binding:"required"`
		Note   string `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	if req.Status != "present" && req.Status != "absent" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "الحالة يجب أن تكون present أو absent"})
		return
	}

	result, err := h.attendanceService.ToggleAttendance(c.Request.Context(), uuid.Nil, employeeID, req.Date, req.Status, req.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("فشل في تحديث الحضور: %v", err)})
		return
	}

	// Log the action
	action := "تم تسجيل حضور"
	details := result.EmployeeName + " - " + req.Date
	if req.Status == "absent" {
		action = "تم تسجيل غياب"
	}
	_ = h.auditService.LogAction(c.Request.Context(), adminName, action, details, c.ClientIP(), adminBranchID)

	c.JSON(http.StatusOK, gin.H{"data": result})
}

type BranchHandler struct {
	branchService service.BranchService
	auditService  service.AuditService
}

func NewBranchHandler(branchService service.BranchService, auditService service.AuditService) *BranchHandler {
	return &BranchHandler{branchService: branchService, auditService: auditService}
}

func (h *BranchHandler) GetAll(c *gin.Context) {
	branches, err := h.branchService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, branches)
}

func (h *BranchHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	branchID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفرع غير صالح"})
		return
	}

	branch, err := h.branchService.GetByID(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, branch)
}

func (h *BranchHandler) Create(c *gin.Context) {
	if !checkAdminOnly(c) {
		return
	}

	var req dto.CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى إدخال اسم الفرع", "details": err.Error()})
		return
	}

	branch, err := h.branchService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "إنشاء فرع", "تم إنشاء الفرع: "+branch.Name, c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusCreated, branch)
}

func (h *BranchHandler) Update(c *gin.Context) {
	if !checkAdminOnly(c) {
		return
	}

	idStr := c.Param("id")
	branchID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفرع غير صالح"})
		return
	}

	var req dto.CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى إدخال اسم الفرع", "details": err.Error()})
		return
	}

	branch, err := h.branchService.Update(c.Request.Context(), branchID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "تعديل فرع", "تم تعديل الفرع إلى: "+branch.Name, c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, branch)
}

func (h *BranchHandler) Delete(c *gin.Context) {
	if !checkAdminOnly(c) {
		return
	}

	idStr := c.Param("id")
	branchID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفرع غير صالح"})
		return
	}

	if err := h.branchService.Delete(c.Request.Context(), branchID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "حذف فرع", "تم حذف فرع", c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الفرع بنجاح"})
}

// getBranchID extracts branch_id from gin context, returns nil for general manager.
// Correctly handles the Go interface nil-trap: a nil *uuid.UUID wrapped in an interface
// is != nil at the interface level, so we must type-assert first.
func getBranchID(c *gin.Context) *uuid.UUID {
	bid, exists := c.Get("branch_id")
	if !exists {
		return nil
	}
	if b, ok := bid.(*uuid.UUID); ok {
		return b
	}
	if b, ok := bid.(uuid.UUID); ok {
		return &b
	}
	return nil
}

// checkEmployeeBranchAccess verifies that the current admin (from context) can access the given employee.
// General managers (branch_id == nil) can access all employees. Supervisors can only access their branch.
// Returns true if access is allowed; false and sends a 403 JSON response if denied.
func checkEmployeeBranchAccess(c *gin.Context, emp *domain.Employee) bool {
	adminBranchID := getBranchID(c)
	if adminBranchID == nil {
		return true // general admin - full access
	}
	if emp.BranchID == nil || *emp.BranchID != *adminBranchID {
		c.JSON(http.StatusForbidden, gin.H{"error": "ليس لديك صلاحية الوصول لهذا المندوب - الفرع غير مطابق"})
		return false
	}
	return true
}

// checkAdminOnly verifies the current user is a general admin (not supervisor).
// Returns true if allowed; false and sends a 403 JSON response if denied.
func checkAdminOnly(c *gin.Context) bool {
	adminBranchID := getBranchID(c)
	if adminBranchID != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "هذه العملية متاحة للمدير العام فقط"})
		return false
	}
	return true
}

// WorkHandler - CheckOilChange
func (h *WorkHandler) CheckOilChange(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	if _, ok := h.checkWorkEmployeeBranch(c, empID); !ok {
		return
	}

	resp, err := h.workService.CheckOilChange(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// InventoryHandler
type InventoryHandler struct {
	invService   service.InventoryService
	auditService service.AuditService
}

func NewInventoryHandler(invService service.InventoryService, auditService service.AuditService) *InventoryHandler {
	return &InventoryHandler{invService: invService, auditService: auditService}
}

func (h *InventoryHandler) CreateItem(c *gin.Context) {
	var req dto.CreateInventoryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى ملء جميع الحقول المطلوبة", "details": err.Error()})
		return
	}

	item, err := h.invService.CreateItem(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "إضافة صنف للمخزن", "تم إضافة الصنف: "+item.Name+" (الكمية: "+strconv.Itoa(item.Quantity)+")", c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusCreated, item)
}

func (h *InventoryHandler) UpdateItem(c *gin.Context) {
	idStr := c.Param("id")
	itemID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الصنف غير صالح"})
		return
	}

	var req dto.UpdateInventoryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "البيانات غير صالحة", "details": err.Error()})
		return
	}

	item, err := h.invService.UpdateItem(c.Request.Context(), itemID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "تعديل صنف بالمخزن", "تم تعديل الصنف: "+item.Name, c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, item)
}

func (h *InventoryHandler) DeleteItem(c *gin.Context) {
	idStr := c.Param("id")
	itemID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الصنف غير صالح"})
		return
	}

	if err := h.invService.DeleteItem(c.Request.Context(), itemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "حذف صنف من المخزن", "تم حذف صنف من المخزن", c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الصنف بنجاح"})
}

func (h *InventoryHandler) GetItems(c *gin.Context) {
	itemType := c.Query("type")
	branchID := getBranchID(c)
	items, err := h.invService.GetAllItems(c.Request.Context(), itemType, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *InventoryHandler) GetItemByID(c *gin.Context) {
	idStr := c.Param("id")
	itemID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الصنف غير صالح"})
		return
	}

	item, err := h.invService.GetItemByID(c.Request.Context(), itemID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الصنف غير موجود"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *InventoryHandler) FindByBarcode(c *gin.Context) {
	barcode := c.Query("barcode")
	if barcode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى إدخال الباركود"})
		return
	}

	item, err := h.invService.FindByBarcode(c.Request.Context(), barcode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "لا يوجد صنف بهذا الباركود"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *InventoryHandler) AddStock(c *gin.Context) {
	var req dto.InventoryTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى ملء جميع الحقول المطلوبة", "details": err.Error()})
		return
	}

	branchID := getBranchID(c)
	tx, err := h.invService.AddStock(c.Request.Context(), req, branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "إضافة مخزون", "تمت إضافة "+strconv.Itoa(req.Quantity)+" "+tx.Item.Unit+" إلى "+tx.Item.Name, c.ClientIP(), branchID)

	c.JSON(http.StatusCreated, tx)
}

func (h *InventoryHandler) RemoveStock(c *gin.Context) {
	var req dto.InventoryTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى ملء جميع الحقول المطلوبة", "details": err.Error()})
		return
	}

	branchID := getBranchID(c)
	tx, err := h.invService.RemoveStock(c.Request.Context(), req, branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "صرف من المخزن", "تم صرف "+strconv.Itoa(req.Quantity)+" "+tx.Item.Unit+" من "+tx.Item.Name, c.ClientIP(), branchID)

	c.JSON(http.StatusCreated, tx)
}

func (h *InventoryHandler) DispenseOil(c *gin.Context) {
	var req dto.DispenseOilRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى ملء جميع الحقول المطلوبة", "details": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	branchID := getBranchID(c)
	log, err := h.invService.DispenseOil(c.Request.Context(), req, adminName, branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	empName := "مندوب"
	if log.Employee != nil {
		empName = log.Employee.Name
	}
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "صرف زيت لمندوب", "تم صرف "+strconv.Itoa(req.Quantity)+" جركن زيت للمندوب: "+empName, c.ClientIP(), branchID)

	c.JSON(http.StatusCreated, log)
}

func (h *InventoryHandler) GetTransactions(c *gin.Context) {
	var itemID *uuid.UUID
	if idStr := c.Query("item_id"); idStr != "" {
		parsed, err := uuid.Parse(idStr)
		if err == nil {
			itemID = &parsed
		}
	}

	branchID := getBranchID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	txs, total, err := h.invService.GetTransactions(c.Request.Context(), itemID, branchID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"data":        txs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func (h *InventoryHandler) DeleteAllTransactions(c *gin.Context) {
	if !checkAdminOnly(c) {
		return
	}

	if err := h.invService.DeleteAllTransactions(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حذف الحركات: " + err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "مسح حركات المخزون", "تم مسح جميع حركات المخزون", c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم مسح جميع حركات المخزون بنجاح"})
}

func (h *InventoryHandler) CreatePurchaseInvoice(c *gin.Context) {
	var req dto.CreatePurchaseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى ملء جميع الحقول المطلوبة", "details": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	branchID := getBranchID(c)

	invoice, err := h.invService.CreatePurchaseInvoice(c.Request.Context(), req, branchID, adminName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = h.auditService.LogAction(c.Request.Context(), adminName, "إنشاء فاتورة مشتريات", fmt.Sprintf("تم إنشاء فاتورة مشتريات رقم %s للمورد %s بإجمالي %.2f", invoice.InvoiceNumber, invoice.SupplierName, invoice.TotalAmount), c.ClientIP(), branchID)

	c.JSON(http.StatusCreated, invoice)
}

func (h *InventoryHandler) GetPurchaseInvoices(c *gin.Context) {
	branchID := getBranchID(c)
	search := strings.TrimSpace(c.Query("search"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	invoices, total, err := h.invService.GetPurchaseInvoices(c.Request.Context(), branchID, search, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"data":        invoices,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func (h *InventoryHandler) GetPurchaseInvoiceByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفاتورة غير صالح"})
		return
	}

	invoice, err := h.invService.GetPurchaseInvoiceByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "فاتورة المشتريات غير موجودة"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

func (h *InventoryHandler) DeletePurchaseInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفاتورة غير صالح"})
		return
	}

	if err := h.invService.DeletePurchaseInvoice(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حذف فاتورة المشتريات: " + err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "حذف فاتورة مشتريات", "تم حذف فاتورة مشتريات معرف: "+id.String(), c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف فاتورة المشتريات بنجاح"})
}

// MaintenanceHandler
type MaintenanceHandler struct {
	maintService service.MaintenanceService
}

func NewMaintenanceHandler(maintService service.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{maintService: maintService}
}

func (h *MaintenanceHandler) GetEmployeeLogs(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	logs, err := h.maintService.GetEmployeeLogs(c.Request.Context(), empID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *MaintenanceHandler) GetAllLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, total, err := h.maintService.GetAllLogs(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

type SettingHandler struct {
	settingService service.SettingService
	auditService   service.AuditService
}

func NewSettingHandler(settingService service.SettingService, auditService service.AuditService) *SettingHandler {
	return &SettingHandler{settingService: settingService, auditService: auditService}
}

func (h *SettingHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingService.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *SettingHandler) UpdateSettings(c *gin.Context) {
	if !checkAdminOnly(c) {
		return
	}

	var req dto.UpdateAppSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	if err := h.settingService.UpdateSettings(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "تحديث الإعدادات", "تم تحديث إعدادات النظام: "+req.SiteName, c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم حفظ الإعدادات بنجاح"})
}

// Public settings endpoint (no auth needed) for login page
func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.settingService.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// ------------------------------------------------------------------
// Vehicle Handler Implementation (الدبابات والمركبات)
// ------------------------------------------------------------------

type VehicleHandler struct {
	vehicleService service.VehicleService
	auditService   service.AuditService
}

func NewVehicleHandler(vehicleService service.VehicleService, auditService service.AuditService) *VehicleHandler {
	return &VehicleHandler{vehicleService: vehicleService, auditService: auditService}
}

func (h *VehicleHandler) Create(c *gin.Context) {
	var req dto.CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	// For branch supervisors, lock branch_id
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if b, ok := bid.(*uuid.UUID); ok && b != nil {
			req.BranchID = b
		}
	}

	vehicle, err := h.vehicleService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName, _ := c.Get("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName.(string), "إضافة دباب/مركبة", fmt.Sprintf("تمت إضافة الدباب رقم اللوحة %s", vehicle.PlateNumber), c.ClientIP(), vehicle.BranchID)

	c.JSON(http.StatusCreated, vehicle)
}

func (h *VehicleHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المركبة غير صالح"})
		return
	}

	var req dto.UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	vehicle, err := h.vehicleService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName, _ := c.Get("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName.(string), "تعديل دباب/مركبة", fmt.Sprintf("تم تعديل بيانات الدباب رقم اللوحة %s", vehicle.PlateNumber), c.ClientIP(), vehicle.BranchID)

	c.JSON(http.StatusOK, vehicle)
}

func (h *VehicleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المركبة غير صالح"})
		return
	}

	if err := h.vehicleService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حذف المركبة"})
		return
	}

	adminName, _ := c.Get("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName.(string), "حذف دباب/مركبة", fmt.Sprintf("تم حذف المركبة معرف %s", idStr), c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف المركبة بنجاح"})
}

func (h *VehicleHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المركبة غير صالح"})
		return
	}

	vehicle, err := h.vehicleService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "المركبة غير موجودة"})
		return
	}

	c.JSON(http.StatusOK, vehicle)
}

func (h *VehicleHandler) GetAll(c *gin.Context) {
	var filter dto.VehicleFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معاملات بحث غير صالحة"})
		return
	}

	// Filter by branch for supervisors
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if b, ok := bid.(*uuid.UUID); ok && b != nil {
			filter.BranchID = b
		}
	}

	vehicles, total, err := h.vehicleService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب قائمة المركبات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  vehicles,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

func (h *VehicleHandler) CheckKM(c *gin.Context) {
	plate := strings.TrimSpace(c.Query("plate"))
	if plate == "" {
		plate = strings.TrimSpace(c.Query("motorcycle_number"))
	}
	if plate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رقم اللوحة مطلوب"})
		return
	}

	latestKM, err := h.vehicleService.GetLatestKM(c.Request.Context(), plate)
	if err != nil || latestKM == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "لا يوجد عداد مسجل لهذه المركبة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plate_number": plate,
		"current_km":   latestKM,
	})
}

func (h *VehicleHandler) RecordOilChange(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المركبة غير صالح"})
		return
	}

	if err := h.vehicleService.RecordOilChange(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName, _ := c.Get("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName.(string), "تغيير زيت للدباب", fmt.Sprintf("تم تسجيل تغيير زيت للدباب معرف %s", idStr), c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{"message": "تم تسجيل تغيير الزيت وتحديث العداد بنجاح"})
}

// ------------------------------------------------------------------
// 1. FuelLogHandler (سجلات الوقود)
// ------------------------------------------------------------------
type FuelLogHandler struct {
	fuelLogService service.FuelLogService
	auditService   service.AuditService
}

func NewFuelLogHandler(fuelLogService service.FuelLogService, auditService service.AuditService) *FuelLogHandler {
	return &FuelLogHandler{fuelLogService: fuelLogService, auditService: auditService}
}

func (h *FuelLogHandler) Create(c *gin.Context) {
	var req dto.CreateFuelLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		val := bID.(uuid.UUID)
		branchID = &val
	}

	log, err := h.fuelLogService.Create(c.Request.Context(), req, branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, log)
}

func (h *FuelLogHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateFuelLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	log, err := h.fuelLogService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, log)
}

func (h *FuelLogHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.fuelLogService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف سجل الوقود بنجاح"})
}

func (h *FuelLogHandler) GetAll(c *gin.Context) {
	var filter dto.FuelLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		val := bID.(uuid.UUID)
		branchID = &val
	}

	logs, total, err := h.fuelLogService.GetAll(c.Request.Context(), filter, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalCost, totalLiters, totalCount, _ := h.fuelLogService.GetStats(c.Request.Context(), branchID, filter.StartDate, filter.EndDate)

	c.JSON(http.StatusOK, gin.H{
		"data":         logs,
		"total":        total,
		"total_cost":   totalCost,
		"total_liters": totalLiters,
		"total_count":  totalCount,
		"page":         filter.Page,
		"limit":        filter.Limit,
	})
}

// ------------------------------------------------------------------
// 2. TrafficViolationHandler (المخالفات المرورية)
// ------------------------------------------------------------------
type TrafficViolationHandler struct {
	violationService service.TrafficViolationService
	auditService     service.AuditService
}

func NewTrafficViolationHandler(violationService service.TrafficViolationService, auditService service.AuditService) *TrafficViolationHandler {
	return &TrafficViolationHandler{violationService: violationService, auditService: auditService}
}

func (h *TrafficViolationHandler) Create(c *gin.Context) {
	var req dto.CreateTrafficViolationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if ptr, ok := bID.(*uuid.UUID); ok {
			branchID = ptr
		} else if val, ok := bID.(uuid.UUID); ok {
			branchID = &val
		}
	}

	v, err := h.violationService.Create(c.Request.Context(), req, branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, v)
}

func (h *TrafficViolationHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateTrafficViolationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	v, err := h.violationService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v)
}

func (h *TrafficViolationHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.violationService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف المخالفة بنجاح"})
}

func (h *TrafficViolationHandler) GetAll(c *gin.Context) {
	var filter dto.TrafficViolationFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(uuid.UUID); ok {
			branchID = &val
		} else if ptrVal, ok := bID.(*uuid.UUID); ok {
			branchID = ptrVal
		}
	}

	list, total, err := h.violationService.GetAll(c.Request.Context(), filter, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalAmount, deductedAmount, totalCount, _ := h.violationService.GetStats(c.Request.Context(), branchID)

	c.JSON(http.StatusOK, gin.H{
		"data":            list,
		"total":           total,
		"total_amount":    totalAmount,
		"deducted_amount": deductedAmount,
		"total_count":     totalCount,
		"page":            filter.Page,
		"limit":           filter.Limit,
	})
}

// ------------------------------------------------------------------
// 3. MaintenanceRequestHandler (طلبات الصيانة)
// ------------------------------------------------------------------
type MaintenanceRequestHandler struct {
	maintenanceService service.MaintenanceRequestService
	auditService       service.AuditService
}

func NewMaintenanceRequestHandler(maintenanceService service.MaintenanceRequestService, auditService service.AuditService) *MaintenanceRequestHandler {
	return &MaintenanceRequestHandler{maintenanceService: maintenanceService, auditService: auditService}
}

func (h *MaintenanceRequestHandler) Create(c *gin.Context) {
	var req dto.CreateMaintenanceRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		val := bID.(uuid.UUID)
		branchID = &val
	}

	m, err := h.maintenanceService.Create(c.Request.Context(), req, branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, m)
}

func (h *MaintenanceRequestHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateMaintenanceRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	m, err := h.maintenanceService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, m)
}

func (h *MaintenanceRequestHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.maintenanceService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف طلب الصيانة بنجاح"})
}

func (h *MaintenanceRequestHandler) GetAll(c *gin.Context) {
	var filter dto.MaintenanceRequestFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(uuid.UUID); ok {
			branchID = &val
		} else if ptrVal, ok := bID.(*uuid.UUID); ok {
			branchID = ptrVal
		}
	}

	list, total, err := h.maintenanceService.GetAll(c.Request.Context(), filter, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

// ------------------------------------------------------------------
// 4. EmployeeDocumentHandler (المستندات والرخص)
// ------------------------------------------------------------------
type EmployeeDocumentHandler struct {
	docService service.EmployeeDocumentService
}

func NewEmployeeDocumentHandler(docService service.EmployeeDocumentService) *EmployeeDocumentHandler {
	return &EmployeeDocumentHandler{docService: docService}
}

func (h *EmployeeDocumentHandler) Create(c *gin.Context) {
	var req dto.CreateEmployeeDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	doc, err := h.docService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

func (h *EmployeeDocumentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateEmployeeDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	doc, err := h.docService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *EmployeeDocumentHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	doc, err := h.docService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "المستند غير موجود", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *EmployeeDocumentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.docService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف المستند بنجاح"})
}

func (h *EmployeeDocumentHandler) GetAll(c *gin.Context) {
	var filter dto.EmployeeDocumentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	list, total, err := h.docService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

func (h *EmployeeDocumentHandler) GetExpiringSoon(c *gin.Context) {
	list, err := h.docService.GetExpiringSoon(c.Request.Context(), 30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": len(list),
	})
}

// ------------------------------------------------------------------
// 5. EmployeeBankAccountHandler (الحسابات البنكية)
// ------------------------------------------------------------------
type EmployeeBankAccountHandler struct {
	bankService service.EmployeeBankAccountService
}

func NewEmployeeBankAccountHandler(bankService service.EmployeeBankAccountService) *EmployeeBankAccountHandler {
	return &EmployeeBankAccountHandler{bankService: bankService}
}

func (h *EmployeeBankAccountHandler) Create(c *gin.Context) {
	var req dto.CreateEmployeeBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	acc, err := h.bankService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, acc)
}

func (h *EmployeeBankAccountHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateEmployeeBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	acc, err := h.bankService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, acc)
}

func (h *EmployeeBankAccountHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.bankService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الحساب البنكي بنجاح"})
}

func (h *EmployeeBankAccountHandler) GetAll(c *gin.Context) {
	var filter dto.EmployeeBankAccountFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	list, total, err := h.bankService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

// ------------------------------------------------------------------
// 6. LeaveRequestHandler (طلبات الإجازات)
// ------------------------------------------------------------------
type LeaveRequestHandler struct {
	leaveService service.LeaveRequestService
	auditService service.AuditService
}

func NewLeaveRequestHandler(leaveService service.LeaveRequestService, auditService service.AuditService) *LeaveRequestHandler {
	return &LeaveRequestHandler{leaveService: leaveService, auditService: auditService}
}

func (h *LeaveRequestHandler) Create(c *gin.Context) {
	var req dto.CreateLeaveRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	leave, err := h.leaveService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, leave)
}

func (h *LeaveRequestHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateLeaveRequestStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	if req.ApprovedByName == "" {
		if adminName, exists := c.Get("admin_name"); exists && adminName != nil {
			req.ApprovedByName = adminName.(string)
		}
	}

	leave, err := h.leaveService.UpdateStatus(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, leave)
}

func (h *LeaveRequestHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.leaveService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف طلب الإجازة بنجاح"})
}

func (h *LeaveRequestHandler) GetAll(c *gin.Context) {
	var filter dto.LeaveRequestFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	list, total, err := h.leaveService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

// ------------------------------------------------------------------
// 7. SupportTicketHandler (تذاكر الدعم والشكاوى)
// ------------------------------------------------------------------
type SupportTicketHandler struct {
	ticketService service.SupportTicketService
	auditService  service.AuditService
}

func NewSupportTicketHandler(ticketService service.SupportTicketService, auditService service.AuditService) *SupportTicketHandler {
	return &SupportTicketHandler{ticketService: ticketService, auditService: auditService}
}

func (h *SupportTicketHandler) Create(c *gin.Context) {
	var req dto.CreateSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		val := bID.(uuid.UUID)
		branchID = &val
	}

	ticket, err := h.ticketService.Create(c.Request.Context(), req, branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func (h *SupportTicketHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	ticket, err := h.ticketService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *SupportTicketHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.ticketService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف التذكرة بنجاح"})
}

func (h *SupportTicketHandler) GetAll(c *gin.Context) {
	var filter dto.SupportTicketFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	branchID := getBranchID(c)

	list, total, err := h.ticketService.GetAll(c.Request.Context(), filter, branchID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

// ------------------------------------------------------------------
// 8. ArchiveHandler (سجل الأرشيف والمحذوفات)
// ------------------------------------------------------------------
type ArchiveHandler struct {
	archiveService service.ArchiveService
	auditService   service.AuditService
}

func NewArchiveHandler(archiveService service.ArchiveService, auditService service.AuditService) *ArchiveHandler {
	return &ArchiveHandler{
		archiveService: archiveService,
		auditService:   auditService,
	}
}

func (h *ArchiveHandler) GetArchived(c *gin.Context) {
	var filter dto.ArchiveFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	branchID := getBranchID(c)
	if branchID != nil {
		filter.BranchID = branchID
	}

	res, err := h.archiveService.GetArchivedItems(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *ArchiveHandler) Restore(c *gin.Context) {
	var req dto.RestoreArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الاسترجاع غير مكتملة", "details": err.Error()})
		return
	}

	if err := h.archiveService.Restore(c.Request.Context(), req.Type, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "استرجاع من الأرشيف", fmt.Sprintf("تم استرجاع عنصر من نوع %s برقم %s", req.Type, req.ID.String()), c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم استرجاع العنصر بنجاح"})
}

func (h *ArchiveHandler) PermanentDelete(c *gin.Context) {
	var req dto.PermanentDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Fallback to query params if sent as query
		itemType := c.Query("type")
		idStr := c.Query("id")
		parsedID, errParse := uuid.Parse(idStr)
		if itemType == "" || errParse != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الحذف النهائي غير صالحة"})
			return
		}
		req.Type = itemType
		req.ID = parsedID
	}

	if err := h.archiveService.PermanentDelete(c.Request.Context(), req.Type, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "حذف نهائي", fmt.Sprintf("تم حذف نهائي لعنصر من نوع %s برقم %s", req.Type, req.ID.String()), c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف العنصر نهائياً وبنجاح"})
}

func (h *ArchiveHandler) BulkRestore(c *gin.Context) {
	var req dto.BulkArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الاسترجاع الجماعي غير صالحة", "details": err.Error()})
		return
	}

	if err := h.archiveService.BulkRestore(c.Request.Context(), req.Type, req.IDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "استرجاع جماعي", fmt.Sprintf("تم استرجاع %d عناصر من نوع %s", len(req.IDs), req.Type), c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("تم استرجاع %d عناصر بنجاح", len(req.IDs))})
}

func (h *ArchiveHandler) BulkPermanentDelete(c *gin.Context) {
	var req dto.BulkArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الحذف الجماعي غير صالحة", "details": err.Error()})
		return
	}

	if err := h.archiveService.BulkPermanentDelete(c.Request.Context(), req.Type, req.IDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "حذف نهائي جماعي", fmt.Sprintf("تم حذف نهائي لـ %d عناصر من نوع %s", len(req.IDs), req.Type), c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("تم حذف %d عناصر نهائياً بنجاح", len(req.IDs))})
}

// ------------------------------------------------------------------
// 9. OTP Handler (إدارة وتوثيق رموز OTP للمناديب)
// ------------------------------------------------------------------
type OTPHandler struct {
	otpService   service.OTPService
	auditService service.AuditService
}

func NewOTPHandler(otpService service.OTPService, auditService service.AuditService) *OTPHandler {
	return &OTPHandler{
		otpService:   otpService,
		auditService: auditService,
	}
}

// Public: Request OTP by National ID
func (h *OTPHandler) RequestOTP(c *gin.Context) {
	var req dto.RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الطلب غير صالحة: " + err.Error()})
		return
	}

	resp, err := h.otpService.RequestOTP(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Public: Verify 4-digit OTP & Return Login Token
func (h *OTPHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات التحقق غير صالحة: " + err.Error()})
		return
	}

	resp, err := h.otpService.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Protected: Get OTP Requests List for Supervisors/Admins
func (h *OTPHandler) GetOTPList(c *gin.Context) {
	var query dto.OTPListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معاملات البحث غير صالحة"})
		return
	}

	list, total, err := h.otpService.GetOTPList(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب طلبات رموز التحقق: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
	})
}

// Protected: Cancel OTP Request
func (h *OTPHandler) CancelOTP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الرمز غير صالح"})
		return
	}

	if err := h.otpService.CancelOTP(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	_ = h.auditService.LogAction(c.Request.Context(), adminName, "إلغاء رمز OTP", fmt.Sprintf("تم إلغاء رمز التحقق برقم %s", id.String()), c.ClientIP(), getBranchID(c))

	c.JSON(http.StatusOK, gin.H{"message": "تم إلغاء رمز التحقق بنجاح"})
}




