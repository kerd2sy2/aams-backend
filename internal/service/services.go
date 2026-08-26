package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/dto"
	"delivery-backend/internal/repository"
	"delivery-backend/pkg/barcode"
	"delivery-backend/pkg/config"
	"delivery-backend/pkg/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Permission Catalog & Helper
var SystemPermissionsCatalog = []dto.PermissionGroupDTO{
	{
		Group: "employees",
		Label: "إدارة المناديب",
		Permissions: []dto.PermissionItemDTO{
			{Key: "employees.view", Label: "عرض المناديب", Description: "عرض قائمة وبيانات وبطاقات المناديب"},
			{Key: "employees.create", Label: "إضافة مندوب", Description: "تسجيل وإضافة مناديب جدد في النظام"},
			{Key: "employees.edit", Label: "تعديل مندوب", Description: "تعديل بيانات المناديب والمستندات"},
			{Key: "employees.delete", Label: "حذف مندوب", Description: "حذف سجل المندوب من النظام"},
			{Key: "employees.cards", Label: "طباعة البطاقات والباركود", Description: "طباعة بطاقة العمل والباركود و QR للمندوب"},
		},
	},
	{
		Group: "work",
		Label: "الدوام وتسجيل الشفتات",
		Permissions: []dto.PermissionItemDTO{
			{Key: "work.view", Label: "متابعة الدوام", Description: "عرض شاشات العاملين الآن والمنتهي دوامهم وشفتات اليوم"},
			{Key: "work.start", Label: "تسجيل بدء الدوام", Description: "تسجيل خروج وبدء عمل المندوب والمركبة والكيلومترات"},
			{Key: "work.end", Label: "تسجيل إنهاء الدوام", Description: "تسجيل عودة المندوب وإقفال الشفت والطلبات"},
		},
	},
	{
		Group: "custody",
		Label: "العهدة والمصروفات",
		Permissions: []dto.PermissionItemDTO{
			{Key: "custody.view", Label: "عرض العهدة", Description: "عرض رصيد العهدة وسجلات المصروفات اليومية"},
			{Key: "custody.add", Label: "إضافة رصيد ومصروف", Description: "إضافة مبالغ للعهدة وتسجيل بنود المصروفات"},
			{Key: "custody.delete", Label: "حذف المصروفات", Description: "حذف سجلات وحركات العهدة والمصروفات"},
		},
	},
	{
		Group: "fleet",
		Label: "إدارة الأسطول والمركبات",
		Permissions: []dto.PermissionItemDTO{
			{Key: "vehicles.view", Label: "عرض المركبات والدبابات", Description: "استعراض قائمة أسطول المركبات والدبابات وحالتها"},
			{Key: "vehicles.manage", Label: "إدارة المركبات", Description: "إضافة وتعديل وحذف المركبات والدبابات"},
			{Key: "vehicles.oil", Label: "سجلات غيار الزيت", Description: "تسجيل وتصفير غيار الزيت ومتابعة الكيلومترات"},
			{Key: "fuel.view", Label: "عرض سجلات الوقود", Description: "الاطلاع على فواتير وسجلات تعبئة الوقود"},
			{Key: "fuel.manage", Label: "إدارة سجلات الوقود", Description: "إضافة وتعديل وحذف فواتير الوقود"},
			{Key: "violations.view", Label: "عرض المخالفات", Description: "الاطلاع على المخالفات المرورية المسجلة"},
			{Key: "violations.manage", Label: "إدارة المخالفات", Description: "تسجيل وتعديل وحذف المخالفات المرورية"},
			{Key: "maintenance.view", Label: "عرض طلبات الصيانة", Description: "متابعة طلبات صيانة وإصلاح المركبات"},
			{Key: "maintenance.manage", Label: "إدارة طلبات الصيانة", Description: "إنشاء وتحديث وإغلاق طلبات الصيانة"},
		},
	},
	{
		Group: "hr_documents",
		Label: "الموارد البشرية والمستندات",
		Permissions: []dto.PermissionItemDTO{
			{Key: "documents.view", Label: "عرض المستندات والرخص", Description: "الاطلاع على مستندات الموظفين وتنبيهات الانتهاء"},
			{Key: "documents.manage", Label: "إدارة المستندات", Description: "رفع وتجديد وتعديل وثائق ورخص الموظفين"},
			{Key: "bank_accounts.view", Label: "عرض الحسابات البنكية", Description: "الاطلاع على الآيبان والحسابات البنكية للمناديب"},
			{Key: "bank_accounts.manage", Label: "إدارة الحسابات البنكية", Description: "إضافة وتعديل الحسابات البنكية للمناديب"},
			{Key: "leaves.view", Label: "عرض الإجازات", Description: "الاطلاع على طلبات إجازات الموظفين"},
			{Key: "leaves.manage", Label: "إدارة واعتماد الإجازات", Description: "تقديم وقبول ورفض طلبات الإجازات"},
			{Key: "attendance.view", Label: "الحضور والانصراف", Description: "تسجيل ومتابعة كشف حضور وغياب الموظفين"},
			{Key: "tickets.view", Label: "عرض تذاكر الدعم والشكاوى", Description: "الاطلاع على الشكاوى وطلبات الدعم للمناديب"},
			{Key: "tickets.manage", Label: "إدارة التذاكر", Description: "فتح وتحديث ومتابعة وإغلاق تذاكر الدعم"},
		},
	},
	{
		Group: "reports",
		Label: "التقارير والإحصائيات",
		Permissions: []dto.PermissionItemDTO{
			{Key: "reports.view", Label: "عرض التقارير", Description: "عرض تقارير الشفتات والتقرير اليومي وإحصائيات العمل"},
			{Key: "reports.export", Label: "تصدير التقارير Excel", Description: "تصدير التقارير اليومية إلى ملفات إكسل"},
		},
	},
	{
		Group: "investigations",
		Label: "الاستجوابات والتحقيقات",
		Permissions: []dto.PermissionItemDTO{
			{Key: "investigations.view", Label: "عرض التحقيقات", Description: "الاطلاع على التحقيقات والاستجوابات والسلف والغياب"},
			{Key: "investigations.create", Label: "إنشاء طلب / استجواب", Description: "إنشاء استجواب جديد أو طلب سلفة أو إثبات غياب"},
			{Key: "investigations.approve", Label: "اعتماد وقبول التحقيقات", Description: "الموافقة على أو رفض الإجراءات والخصومات"},
		},
	},
	{
		Group: "inventory",
		Label: "المخزون والقطع",
		Permissions: []dto.PermissionItemDTO{
			{Key: "inventory.view", Label: "عرض المخزون", Description: "استعراض أصناف المخزون والكميات الحالية"},
			{Key: "inventory.manage", Label: "إدارة المخزون والأصناف", Description: "إضافة وتعديل أصناف وتوريد كميات جديدة"},
			{Key: "inventory.dispense", Label: "صرف الزيوت والقطع", Description: "تسجيل عمليات صرف الزيوت وقطع الغيار"},
		},
	},
	{
		Group: "partners",
		Label: "منصات الشركاء",
		Permissions: []dto.PermissionItemDTO{
			{Key: "partners.view", Label: "عرض الشركاء", Description: "الاطلاع على منصات وتطبيقات التوصيل الشريكة"},
			{Key: "partners.manage", Label: "إدارة الشركاء", Description: "إضافة وتعديل وحذف منصات التوصيل"},
		},
	},
	{
		Group: "admin",
		Label: "الإدارة والنظام",
		Permissions: []dto.PermissionItemDTO{
			{Key: "users.manage", Label: "إدارة المستخدمين", Description: "إضافة وتعديل وحذف مستخدمي لوحة التحكم"},
			{Key: "roles.manage", Label: "إدارة الأدوار والصلاحيات", Description: "إنشاء وتعديل الأدوار والصلاحيات وتوزيعها"},
			{Key: "settings.manage", Label: "إعدادات النظام العامة", Description: "تعديل اسم النظام والشعار وإعدادات التطبيق"},
			{Key: "audit_logs.view", Label: "سجل العمليات والنشاط", Description: "الاطلاع على سجلات العمليات والتدقيق في النظام"},
			{Key: "error_logs.view", Label: "سجل أخطاء النظام", Description: "الاطلاع على سجل الأخطاء التقنية"},
			{Key: "archive.view", Label: "عرض سجل الأرشيف والمحذوفات", Description: "الاطلاع على العناصر المؤرشفة والمحذوفة في النظام"},
			{Key: "archive.restore", Label: "استرجاع المحذوفات", Description: "استعادة العناصر المؤرشفة إلى حالتها النشطة"},
			{Key: "archive.delete_permanent", Label: "الحذف النهائي من الأرشيف", Description: "حذف العناصر من قاعدة البيانات نهائياً وبلا رجعة"},
		},
	},
}

func ResolveAdminPermissions(admin *domain.Admin) []string {
	if admin == nil {
		return []string{}
	}
	if strings.ToUpper(admin.Role) == "ADMIN" || strings.ToUpper(admin.Role) == "SUPER_ADMIN" {
		return []string{"*"}
	}
	permMap := make(map[string]bool)
	if admin.RoleObj != nil && admin.RoleObj.Permissions != "" {
		var rolePerms []string
		if err := json.Unmarshal([]byte(admin.RoleObj.Permissions), &rolePerms); err == nil {
			for _, p := range rolePerms {
				if p == "*" {
					return []string{"*"}
				}
				permMap[p] = true
			}
		}
	}
	if admin.Permissions != "" {
		var userPerms []string
		if err := json.Unmarshal([]byte(admin.Permissions), &userPerms); err == nil {
			for _, p := range userPerms {
				if p == "*" {
					return []string{"*"}
				}
				permMap[p] = true
			}
		}
	}
	// Fallback for standard roles if no RoleObj linked yet
	if len(permMap) == 0 && admin.RoleID == nil {
		roleUpper := strings.ToUpper(admin.Role)
		if roleUpper == "SUPERVISOR" {
			return []string{
				"employees.view", "employees.create", "employees.edit", "employees.cards",
				"work.view", "work.start", "work.end", "custody.view", "custody.add",
				"vehicles.view", "vehicles.oil", "fuel.view", "fuel.manage",
				"maintenance.view", "maintenance.manage", "attendance.view",
				"investigations.view", "investigations.create", "inventory.view", "inventory.dispense",
				"tickets.view", "tickets.manage",
			}
		}
	}

	result := make([]string, 0, len(permMap))
	for p := range permMap {
		result = append(result, p)
	}
	return result
}

// AuthService interface & impl
type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.LoginResponse, error)
	GoogleLogin(ctx context.Context, req dto.GoogleLoginRequest) (*dto.LoginResponse, error)
	LinkGoogle(ctx context.Context, adminID uuid.UUID, req dto.GoogleLinkRequest) (*domain.Admin, error)
	UnlinkGoogle(ctx context.Context, adminID uuid.UUID) error
}

type authService struct {
	adminRepo  repository.AdminRepository
	branchRepo repository.BranchRepository
	cfg        *config.Config
}

func NewAuthService(adminRepo repository.AdminRepository, branchRepo repository.BranchRepository, cfg *config.Config) AuthService {
	return &authService{adminRepo: adminRepo, branchRepo: branchRepo, cfg: cfg}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	admin, err := s.adminRepo.FindByLogin(ctx, req.Login)
	if err != nil {
		return nil, errors.New("بيانات الدخول غير صحيحة")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("بيانات الدخول غير صحيحة")
	}

	accessToken, refreshToken, err := jwt.GenerateTokens(admin.ID, admin.Email, admin.Name, admin.Role, admin.BranchID, s.cfg.JWTSecret, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, fmt.Errorf("فشل في إنشاء رمز الجلسة: %w", err)
	}

	resp := &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	resp.Admin.ID = admin.ID
	resp.Admin.Name = admin.Name
	resp.Admin.Email = admin.Email
	resp.Admin.Username = admin.Username
	resp.Admin.Phone = admin.Phone
	resp.Admin.Role = admin.Role
	resp.Admin.RoleID = admin.RoleID
	resp.Admin.Permissions = ResolveAdminPermissions(admin)
	resp.Admin.GoogleEmail = admin.GoogleEmail
	resp.Admin.GoogleAvatar = admin.GoogleAvatar
	resp.Admin.IsGoogleLinked = admin.GoogleEmail != ""
	resp.Admin.BranchID = admin.BranchID

	// Load branch info if admin has a branch
	if admin.BranchID != nil {
		branch, err := s.branchRepo.FindByID(ctx, *admin.BranchID)
		if err == nil {
			resp.Admin.Branch = &struct {
				ID   uuid.UUID `json:"id"`
				Name string    `json:"name"`
			}{ID: branch.ID, Name: branch.Name}
		}
	}

	return resp, nil
}

func (s *authService) GoogleLogin(ctx context.Context, req dto.GoogleLoginRequest) (*dto.LoginResponse, error) {
	if req.Email == "" && req.GoogleID == "" {
		return nil, errors.New("البريد الإلكتروني لحساب Google مطلوب")
	}

	var admin *domain.Admin
	var err error

	// 1. Check by GoogleEmail
	if req.Email != "" {
		admin, err = s.adminRepo.FindByGoogleEmail(ctx, req.Email)
	}

	// 2. If not found, check by GoogleID
	if (err != nil || admin == nil) && req.GoogleID != "" {
		admin, err = s.adminRepo.FindByGoogleID(ctx, req.GoogleID)
	}

	// 3. If still not found, return explicit authorization/link error
	if err != nil || admin == nil || admin.GoogleEmail == "" {
		return nil, errors.New("حساب Google هذا غير مرتبط بأي حساب إداري مصرح به. يرجى تسجيل الدخول باسم المستخدم وكلمة المرور أولاً وربط حسابك من الملف الشخصي")
	}

	accessToken, refreshToken, err := jwt.GenerateTokens(admin.ID, admin.Email, admin.Name, admin.Role, admin.BranchID, s.cfg.JWTSecret, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, fmt.Errorf("فشل في إنشاء رمز الجلسة: %w", err)
	}

	resp := &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	resp.Admin.ID = admin.ID
	resp.Admin.Name = admin.Name
	resp.Admin.Email = admin.Email
	resp.Admin.Username = admin.Username
	resp.Admin.Phone = admin.Phone
	resp.Admin.Role = admin.Role
	resp.Admin.RoleID = admin.RoleID
	resp.Admin.Permissions = ResolveAdminPermissions(admin)
	resp.Admin.GoogleEmail = admin.GoogleEmail
	resp.Admin.GoogleAvatar = admin.GoogleAvatar
	resp.Admin.IsGoogleLinked = admin.GoogleEmail != ""
	resp.Admin.BranchID = admin.BranchID

	if admin.BranchID != nil {
		branch, err := s.branchRepo.FindByID(ctx, *admin.BranchID)
		if err == nil {
			resp.Admin.Branch = &struct {
				ID   uuid.UUID `json:"id"`
				Name string    `json:"name"`
			}{ID: branch.ID, Name: branch.Name}
		}
	}

	return resp, nil
}

func (s *authService) LinkGoogle(ctx context.Context, adminID uuid.UUID, req dto.GoogleLinkRequest) (*domain.Admin, error) {
	if req.Email == "" {
		return nil, errors.New("البريد الإلكتروني لحساب Google مطلوب")
	}

	// Check if this Google email is already linked to another admin
	existing, err := s.adminRepo.FindByGoogleEmail(ctx, req.Email)
	if err == nil && existing != nil && existing.ID != adminID {
		return nil, errors.New("حساب Google هذا مرتبط بالفعل بمستخدم آخر في النظام")
	}

	if err := s.adminRepo.UpdateGoogleLink(ctx, adminID, req.Email, req.GoogleID, req.Avatar); err != nil {
		return nil, fmt.Errorf("فشل في ربط حساب Google: %w", err)
	}

	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return admin, nil
}

func (s *authService) UnlinkGoogle(ctx context.Context, adminID uuid.UUID) error {
	return s.adminRepo.UnlinkGoogle(ctx, adminID)
}

func (s *authService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	claims, err := jwt.ValidateToken(req.RefreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, errors.New("رمز التحديث غير صالحة")
	}

	admin, err := s.adminRepo.FindByID(ctx, claims.AdminID)
	if err != nil {
		return nil, errors.New("المستخدم غير موجود")
	}

	accessToken, refreshToken, err := jwt.GenerateTokens(admin.ID, admin.Email, admin.Name, admin.Role, admin.BranchID, s.cfg.JWTSecret, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, err
	}

	resp := &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	resp.Admin.ID = admin.ID
	resp.Admin.Name = admin.Name
	resp.Admin.Email = admin.Email
	resp.Admin.Username = admin.Username
	resp.Admin.Phone = admin.Phone
	resp.Admin.Role = admin.Role
	resp.Admin.RoleID = admin.RoleID
	resp.Admin.Permissions = ResolveAdminPermissions(admin)
	resp.Admin.BranchID = admin.BranchID

	if admin.BranchID != nil {
		branch, err := s.branchRepo.FindByID(ctx, *admin.BranchID)
		if err == nil {
			resp.Admin.Branch = &struct {
				ID   uuid.UUID `json:"id"`
				Name string    `json:"name"`
			}{ID: branch.ID, Name: branch.Name}
		}
	}

	return resp, nil
}

// RoleService interface & impl
type RoleService interface {
	GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*dto.RoleResponse, error)
	CreateRole(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error)
	UpdateRole(ctx context.Context, id uuid.UUID, req dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	GetPermissionsCatalog() []dto.PermissionGroupDTO
}

type roleService struct {
	roleRepo repository.RoleRepository
}

func NewRoleService(roleRepo repository.RoleRepository) RoleService {
	return &roleService{roleRepo: roleRepo}
}

func (s *roleService) GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := s.roleRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var res []dto.RoleResponse
	for _, r := range roles {
		var perms []string
		if r.Permissions != "" {
			_ = json.Unmarshal([]byte(r.Permissions), &perms)
		}
		usersCount, _ := s.roleRepo.CountUsersByRoleID(ctx, r.ID)
		res = append(res, dto.RoleResponse{
			ID:          r.ID,
			Name:        r.Name,
			DisplayName: r.DisplayName,
			Description: r.Description,
			Permissions: perms,
			IsSystem:    r.IsSystem,
			UsersCount:  usersCount,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		})
	}
	return res, nil
}

func (s *roleService) GetRoleByID(ctx context.Context, id uuid.UUID) (*dto.RoleResponse, error) {
	r, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("الدور غير موجود")
	}
	var perms []string
	if r.Permissions != "" {
		_ = json.Unmarshal([]byte(r.Permissions), &perms)
	}
	usersCount, _ := s.roleRepo.CountUsersByRoleID(ctx, r.ID)
	return &dto.RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		Permissions: perms,
		IsSystem:    r.IsSystem,
		UsersCount:  usersCount,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}, nil
}

func (s *roleService) CreateRole(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.DisplayName) == "" {
		return nil, errors.New("اسم الدور مطلوب")
	}

	cleanName := strings.ToUpper(strings.TrimSpace(req.Name))
	cleanName = strings.ReplaceAll(cleanName, " ", "_")

	if existing, _ := s.roleRepo.FindByName(ctx, cleanName); existing != nil {
		return nil, errors.New("يوجد دور مسجل بهذا الاسم بالفعل")
	}

	permsBytes, err := json.Marshal(req.Permissions)
	if err != nil {
		return nil, fmt.Errorf("خطأ في معالجة الصلاحيات: %w", err)
	}

	role := &domain.Role{
		Name:        cleanName,
		DisplayName: strings.TrimSpace(req.DisplayName),
		Description: strings.TrimSpace(req.Description),
		Permissions: string(permsBytes),
		IsSystem:    false,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("فشل في إنشاء الدور: %w", err)
	}

	return &dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		Permissions: req.Permissions,
		IsSystem:    role.IsSystem,
		UsersCount:  0,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

func (s *roleService) UpdateRole(ctx context.Context, id uuid.UUID, req dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("الدور غير موجود")
	}

	if req.DisplayName != "" {
		role.DisplayName = strings.TrimSpace(req.DisplayName)
	}
	if req.Description != "" {
		role.Description = strings.TrimSpace(req.Description)
	}
	if req.Permissions != nil {
		permsBytes, err := json.Marshal(req.Permissions)
		if err != nil {
			return nil, fmt.Errorf("خطأ في معالجة الصلاحيات: %w", err)
		}
		role.Permissions = string(permsBytes)
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("فشل في تحديث الدور: %w", err)
	}

	var perms []string
	if role.Permissions != "" {
		_ = json.Unmarshal([]byte(role.Permissions), &perms)
	}
	usersCount, _ := s.roleRepo.CountUsersByRoleID(ctx, role.ID)

	return &dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		Permissions: perms,
		IsSystem:    role.IsSystem,
		UsersCount:  usersCount,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

func (s *roleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("الدور غير موجود")
	}

	if role.IsSystem {
		return errors.New("لا يمكن حذف الأدوار الأساسية الخاصة بالنظام")
	}

	usersCount, _ := s.roleRepo.CountUsersByRoleID(ctx, id)
	if usersCount > 0 {
		return fmt.Errorf("لا يمكن حذف هذا الدور لوجود %d مستخدمين مرتبطين به حالياً", usersCount)
	}

	return s.roleRepo.Delete(ctx, id)
}

func (s *roleService) GetPermissionsCatalog() []dto.PermissionGroupDTO {
	return SystemPermissionsCatalog
}

// AdminService interface & impl
type AdminService interface {
	CreateAdmin(ctx context.Context, req dto.CreateAdminRequest) (*domain.Admin, error)
	UpdateAdmin(ctx context.Context, id uuid.UUID, req dto.UpdateAdminRequest) (*domain.Admin, error)
	DeleteAdmin(ctx context.Context, id uuid.UUID) error
	GetAllAdmins(ctx context.Context, branchID *uuid.UUID) ([]domain.Admin, error)
	ChangePassword(ctx context.Context, id uuid.UUID, req dto.ChangePasswordRequest) error
}

type adminService struct {
	adminRepo repository.AdminRepository
}

func NewAdminService(adminRepo repository.AdminRepository) AdminService {
	return &adminService{adminRepo: adminRepo}
}

func (s *adminService) CreateAdmin(ctx context.Context, req dto.CreateAdminRequest) (*domain.Admin, error) {
	// Validate password strength
	if len(req.Password) < 8 {
		return nil, errors.New("كلمة المرور يجب أن لا تقل عن 8 أحرف")
	}
	if !strings.ContainsAny(req.Password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return nil, errors.New("كلمة المرور يجب أن تحتوي على حرف كبير واحد على الأقل (A-Z)")
	}
	if !strings.ContainsAny(req.Password, "abcdefghijklmnopqrstuvwxyz") {
		return nil, errors.New("كلمة المرور يجب أن تحتوي على حرف صغير واحد على الأقل (a-z)")
	}
	if !strings.ContainsAny(req.Password, "0123456789") {
		return nil, errors.New("كلمة المرور يجب أن تحتوي على رقم واحد على الأقل (0-9)")
	}
	if !strings.ContainsAny(req.Password, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
		return nil, errors.New("كلمة المرور يجب أن تحتوي على رمز خاص واحد على الأقل مثل !@#$%")
	}

	// Check if email already exists
	if _, err := s.adminRepo.FindByEmail(ctx, req.Email); err == nil {
		return nil, errors.New("البريد الإلكتروني مستخدم بالفعل")
	}
	// Check if username already exists
	if _, err := s.adminRepo.FindByUsername(ctx, req.Username); err == nil {
		return nil, errors.New("اسم المستخدم مستخدم بالفعل")
	}
	// Check if phone already exists (if provided)
	if req.Phone != "" {
		if _, err := s.adminRepo.FindByPhone(ctx, req.Phone); err == nil {
			return nil, errors.New("رقم الهاتف مستخدم بالفعل")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("فشل في تشفير كلمة المرور: %w", err)
	}

	role := req.Role
	if role == "" {
		role = "SUPERVISOR"
	}

	var permsStr string
	if len(req.Permissions) > 0 {
		permsBytes, _ := json.Marshal(req.Permissions)
		permsStr = string(permsBytes)
	}

	admin := &domain.Admin{
		Email:       req.Email,
		Username:    req.Username,
		Phone:       req.Phone,
		Password:    string(hashedPassword),
		Name:        req.Name,
		Role:        role,
		RoleID:      req.RoleID,
		Permissions: permsStr,
		BranchID:    req.BranchID,
	}

	if err := s.adminRepo.Create(ctx, admin); err != nil {
		return nil, fmt.Errorf("فشل في إنشاء المدير: %w", err)
	}

	return admin, nil
}

func (s *adminService) UpdateAdmin(ctx context.Context, id uuid.UUID, req dto.UpdateAdminRequest) (*domain.Admin, error) {
	admin, err := s.adminRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("المستخدم غير موجود")
	}

	if req.Name != "" {
		admin.Name = req.Name
	}
	if req.Email != "" && req.Email != admin.Email {
		// Check uniqueness if changing email
		if existing, _ := s.adminRepo.FindByEmail(ctx, req.Email); existing != nil && existing.ID != id {
			return nil, errors.New("البريد الإلكتروني مستخدم بالفعل")
		}
		admin.Email = req.Email
	}
	if req.Username != "" && req.Username != admin.Username {
		if existing, _ := s.adminRepo.FindByUsername(ctx, req.Username); existing != nil && existing.ID != id {
			return nil, errors.New("اسم المستخدم مستخدم بالفعل")
		}
		admin.Username = req.Username
	}
	if req.Phone != "" {
		if req.Phone != admin.Phone {
			if existing, _ := s.adminRepo.FindByPhone(ctx, req.Phone); existing != nil && existing.ID != id {
				return nil, errors.New("رقم الهاتف مستخدم بالفعل")
			}
		}
		admin.Phone = req.Phone
	}
	if req.Role != "" {
		admin.Role = req.Role
	}
	if req.RoleID != nil {
		admin.RoleID = req.RoleID
	}
	if req.Permissions != nil {
		permsBytes, _ := json.Marshal(req.Permissions)
		admin.Permissions = string(permsBytes)
	}
	if req.BranchID != nil {
		admin.BranchID = req.BranchID
	}
	if req.Password != "" {
		if len(req.Password) < 8 {
			return nil, errors.New("كلمة المرور يجب أن لا تقل عن 8 أحرف")
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("فشل في تشفير كلمة المرور: %w", err)
		}
		admin.Password = string(hashedPassword)
	}

	if err := s.adminRepo.Update(ctx, admin); err != nil {
		return nil, fmt.Errorf("فشل في تحديث المستخدم: %w", err)
	}

	return admin, nil
}


func (s *adminService) DeleteAdmin(ctx context.Context, id uuid.UUID) error {
	_, err := s.adminRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("المدير غير موجود")
	}
	return s.adminRepo.Delete(ctx, id)
}

func (s *adminService) GetAllAdmins(ctx context.Context, branchID *uuid.UUID) ([]domain.Admin, error) {
	admins, err := s.adminRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	// For supervisors, filter to show only admins from same branch
	if branchID != nil {
		var filtered []domain.Admin
		for _, a := range admins {
			if a.BranchID != nil && *a.BranchID == *branchID {
				filtered = append(filtered, a)
			}
		}
		return filtered, nil
	}
	return admins, nil
}

func (s *adminService) ChangePassword(ctx context.Context, id uuid.UUID, req dto.ChangePasswordRequest) error {
	// Validate new password strength
	if len(req.NewPassword) < 8 {
		return errors.New("كلمة المرور الجديدة يجب أن لا تقل عن 8 أحرف")
	}

	admin, err := s.adminRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("المدير غير موجود")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("كلمة المرور الحالية غير صحيحة")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("فشل في تشفير كلمة المرور: %w", err)
	}

	admin.Password = string(hashedPassword)
	return s.adminRepo.Update(ctx, admin)
}

// EmployeeService interface & impl
type EmployeeService interface {
	Create(ctx context.Context, req dto.CreateEmployeeRequest) (*domain.Employee, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeRequest) (*domain.Employee, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error)
	Search(ctx context.Context, term string, branchID *uuid.UUID) ([]domain.Employee, error)
	GetAll(ctx context.Context, filter dto.EmployeeFilter) (*dto.PaginatedEmployeeResponse, error)
	GetBarcode(ctx context.Context, id uuid.UUID) (string, error)
	GetQRCode(ctx context.Context, id uuid.UUID) (string, error)
	BatchSetOilChange(ctx context.Context, entries []dto.OilSetupEntry) error
}

type employeeService struct {
	empRepo repository.EmployeeRepository
}

func NewEmployeeService(empRepo repository.EmployeeRepository) EmployeeService {
	return &employeeService{empRepo: empRepo}
}

func (s *employeeService) Create(ctx context.Context, req dto.CreateEmployeeRequest) (*domain.Employee, error) {
	existing, err := s.empRepo.FindByNationalID(ctx, req.NationalID)
	if err == nil && existing != nil {
		return nil, errors.New("رقم الهوية الوطنية مسجل بالفعل لموظف آخر")
	}

	employeeID := uuid.New()

	// Automatically generate Barcode (Code128) & QR Code from Employee UUID
	barcodeData, err := barcode.GenerateCode128Base64(employeeID.String())
	if err != nil {
		return nil, fmt.Errorf("فشل توليد الباركود: %w", err)
	}

	qrData, err := barcode.GenerateQRCodeBase64(employeeID.String())
	if err != nil {
		return nil, fmt.Errorf("فشل توليد QR Code: %w", err)
	}

	emp := &domain.Employee{
		ID:                  employeeID,
		Name:                req.Name,
		JobRole:             req.JobRole,
		EmployeeNumber:      req.EmployeeNumber,
		PersonalImage:       req.PersonalImage,
		NationalID:          req.NationalID,
		IqamaExpirationDate: req.IqamaExpirationDate,
		NationalIDImage:     req.NationalIDImage,
		DrivingLicenseImage: req.DrivingLicenseImage,
		KeyNumber:           req.KeyNumber,
		MotorcycleNumber:    req.MotorcycleNumber,
		ApplicationID:       req.ApplicationID,
		ApplicationType:     req.ApplicationType,
		VehicleType:         req.VehicleType,
		Shift:               req.Shift,
		BranchID:            req.BranchID,
		Barcode:             barcodeData,
		QRCode:              qrData,
	}
	// Default to motorcycle if not specified
	if emp.VehicleType == "" {
		emp.VehicleType = "motorcycle"
	}
	// Default to morning shift if not specified
	if emp.Shift == "" {
		emp.Shift = "morning"
	}

	if err := s.empRepo.Create(ctx, emp); err != nil {
		return nil, err
	}

	return emp, nil
}

func (s *employeeService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeRequest) (*domain.Employee, error) {
	emp, err := s.empRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("الموظف غير موجود")
	}

	if req.Name != "" {
		emp.Name = req.Name
	}
	if req.JobRole != "" {
		emp.JobRole = req.JobRole
	}
	if req.EmployeeNumber != "" {
		emp.EmployeeNumber = req.EmployeeNumber
	}
	if req.PersonalImage != "" {
		emp.PersonalImage = req.PersonalImage
	}
	if req.NationalID != "" {
		emp.NationalID = req.NationalID
	}
	if req.IqamaExpirationDate != nil {
		emp.IqamaExpirationDate = req.IqamaExpirationDate
	}
	if req.NationalIDImage != "" {
		emp.NationalIDImage = req.NationalIDImage
	}
	if req.DrivingLicenseImage != "" {
		emp.DrivingLicenseImage = req.DrivingLicenseImage
	}
	if req.KeyNumber != "" {
		emp.KeyNumber = req.KeyNumber
	}
	if req.MotorcycleNumber != "" {
		emp.MotorcycleNumber = req.MotorcycleNumber
	}
	if req.ApplicationID != "" {
		emp.ApplicationID = req.ApplicationID
	}
	if req.ApplicationType != "" {
		emp.ApplicationType = req.ApplicationType
	}
	if req.VehicleType != "" {
		emp.VehicleType = req.VehicleType
	}
	if req.Shift != "" {
		emp.Shift = req.Shift
	}
	if req.BranchID != nil {
		emp.BranchID = req.BranchID
	}

	if err := s.empRepo.Update(ctx, emp); err != nil {
		return nil, err
	}

	return emp, nil
}

func (s *employeeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.empRepo.Delete(ctx, id)
}

func (s *employeeService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	return s.empRepo.FindByID(ctx, id)
}

func (s *employeeService) Search(ctx context.Context, term string, branchID *uuid.UUID) ([]domain.Employee, error) {
	return s.empRepo.FindBySearchTerm(ctx, term, branchID)
}

func (s *employeeService) BatchSetOilChange(ctx context.Context, entries []dto.OilSetupEntry) error {
	for _, entry := range entries {
		empID, err := uuid.Parse(entry.EmployeeID)
		if err != nil {
			return fmt.Errorf("معرف موظف غير صالح: %s", entry.EmployeeID)
		}

		emp, err := s.empRepo.FindByID(ctx, empID)
		if err != nil {
			return fmt.Errorf("الموظف غير موجود: %s", empID)
		}

		emp.LastOilChangeDistance = entry.LastOilChangeDistance
		if err := s.empRepo.Update(ctx, emp); err != nil {
			return fmt.Errorf("فشل تحديث الموظف %s: %w", emp.Name, err)
		}
	}
	return nil
}

func (s *employeeService) GetAll(ctx context.Context, filter dto.EmployeeFilter) (*dto.PaginatedEmployeeResponse, error) {
	employees, total, err := s.empRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &dto.PaginatedEmployeeResponse{
		Data:       employees,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *employeeService) GetBarcode(ctx context.Context, id uuid.UUID) (string, error) {
	emp, err := s.empRepo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if emp.Barcode != "" {
		return emp.Barcode, nil
	}
	return barcode.GenerateCode128Base64(id.String())
}

func (s *employeeService) GetQRCode(ctx context.Context, id uuid.UUID) (string, error) {
	emp, err := s.empRepo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if emp.QRCode != "" {
		return emp.QRCode, nil
	}
	return barcode.GenerateQRCodeBase64(id.String())
}

// WorkService interface & impl
type WorkService interface {
	StartWork(ctx context.Context, req dto.StartWorkRequest) (*domain.WorkSession, error)
	EndWork(ctx context.Context, req dto.EndWorkRequest) (*domain.WorkSession, error)
	UpdateWorkSession(ctx context.Context, sessionID uuid.UUID, req dto.UpdateWorkSessionRequest) (*domain.WorkSession, error)
	GetActiveSession(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error)
	GetLastCompletedSession(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error)
	GetLastSessionOrVehicleKM(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (float64, float64, error)
	CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error)
	CheckOilChange(ctx context.Context, empID uuid.UUID) (*dto.OilChangeCheckResponse, error)
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*domain.WorkSession, error)
}

type workService struct {
	workRepo        repository.WorkRepository
	empRepo         repository.EmployeeRepository
	maintenanceRepo repository.MaintenanceRepository
	vehicleRepo     repository.VehicleRepository
}

func NewWorkService(workRepo repository.WorkRepository, empRepo repository.EmployeeRepository, maintenanceRepo repository.MaintenanceRepository, vehicleRepo repository.VehicleRepository) WorkService {
	return &workService{workRepo: workRepo, empRepo: empRepo, maintenanceRepo: maintenanceRepo, vehicleRepo: vehicleRepo}
}

// oilChangeInterval returns the oil change interval in km for a given vehicle type.
func oilChangeInterval(vehicleType string) float64 {
	if strings.EqualFold(vehicleType, "car") {
		return 10000
	}
	return 950 // motorcycle (default)
}

func (s *workService) StartWork(ctx context.Context, req dto.StartWorkRequest) (*domain.WorkSession, error) {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("معرف الموظف غير صالح")
	}

	emp, err := s.empRepo.FindByID(ctx, empID)
	if err != nil {
		return nil, errors.New("الموظف غير موجود")
	}

	// Check if already active
	activeSession, _ := s.workRepo.FindActiveSessionByEmployeeID(ctx, empID)
	if activeSession != nil {
		return nil, errors.New("الموظف لديه شفت عمل نشط بالفعل حالياً")
	}

	// Vehicle type: use request override if provided, otherwise employee's default
	vehicleType := emp.VehicleType
	if req.VehicleType != "" {
		vehicleType = req.VehicleType
	}

	// Check oil change requirement based on vehicle type (car: 10000km, motorcycle: 950km)
	distanceSinceOil := emp.TotalDistance - emp.LastOilChangeDistance
	interval := oilChangeInterval(vehicleType)
	if distanceSinceOil >= interval {
		return nil, fmt.Errorf("يجب تغيير الزيت أولاً! المسافة المقطوعة منذ آخر تغيير زيت: %.0f كم (الحد الأقصى %.0f كم)", distanceSinceOil, interval)
	}

	// Use the application_id from request if provided, otherwise fallback to employee's default
	appID := emp.ApplicationID
	if req.ApplicationID != "" {
		appID = req.ApplicationID
	}

	// Use the application_type from request if provided, otherwise fallback to employee's default
	appType := emp.ApplicationType
	if req.ApplicationType != "" {
		appType = req.ApplicationType
	}

	// Motorcycle number: use request override if provided, otherwise employee's default
	motorcycleNumber := emp.MotorcycleNumber
	if req.MotorcycleNumber != "" {
		motorcycleNumber = req.MotorcycleNumber
	}

	session := &domain.WorkSession{
		ID:               uuid.New(),
		EmployeeID:       &empID,
		StartTime:        time.Now(),
		StartKM:          req.StartKM,
		ApplicationID:    appID,
		ApplicationType:  appType,
		VehicleType:      vehicleType,
		MotorcycleNumber: motorcycleNumber,
		Notes:            req.Notes,
		Status:           domain.StatusActive,
	}

	if err := s.workRepo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	// Update vehicle status to in-use
	if motorcycleNumber != "" && s.vehicleRepo != nil {
		_ = s.vehicleRepo.SetVehicleStatus(ctx, motorcycleNumber, domain.VehicleStatusInUse)
	}

	session.Employee = emp
	return session, nil
}

func (s *workService) EndWork(ctx context.Context, req dto.EndWorkRequest) (*domain.WorkSession, error) {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("معرف الموظف غير صالح")
	}

	activeSession, err := s.workRepo.FindActiveSessionByEmployeeID(ctx, empID)
	if err != nil || activeSession == nil {
		return nil, errors.New("لا يوجد شفت عمل قائم لهذا الموظف لبدء إنهائه (يجب بدء العمل أولاً)")
	}

	if req.EndKM <= activeSession.StartKM {
		return nil, fmt.Errorf("قراءة عداد النهاية (%.2f) يجب أن تكون أكبر من قراءة البداية (%.2f)", req.EndKM, activeSession.StartKM)
	}

	now := time.Now()
	distance := req.EndKM - activeSession.StartKM

	activeSession.EndTime = &now
	activeSession.EndKM = req.EndKM
	activeSession.Distance = distance
	activeSession.OrdersCount = req.OrdersCount
	activeSession.FuelCost = req.FuelCost
	if req.ApplicationID != "" {
		activeSession.ApplicationID = req.ApplicationID
	}
	if req.ApplicationType != "" {
		activeSession.ApplicationType = req.ApplicationType
	}
	activeSession.Notes = req.Notes
	activeSession.Status = domain.StatusCompleted

	if err := s.workRepo.UpdateSession(ctx, activeSession); err != nil {
		return nil, err
	}

	// Update employee's total distance
	emp, err := s.empRepo.FindByID(ctx, empID)
	if err == nil {
		emp.TotalDistance += distance
		if updateErr := s.empRepo.Update(ctx, emp); updateErr != nil {
			return nil, fmt.Errorf("فشل في تحديث بيانات الموظف: %w", updateErr)
		}
	}

	// Update vehicle odometer and return status to available
	if activeSession.MotorcycleNumber != "" && s.vehicleRepo != nil {
		_ = s.vehicleRepo.UpdateOdometer(ctx, activeSession.MotorcycleNumber, req.EndKM, distance)
	}

	return activeSession, nil
}

func (s *workService) GetLastSessionOrVehicleKM(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (float64, float64, error) {
	// If motorcycle number is specified, prioritize vehicle's latest odometer
	if motorcycleNumber != "" && s.vehicleRepo != nil {
		latestKM, err := s.vehicleRepo.FindLatestVehicleKM(ctx, motorcycleNumber)
		if err == nil && latestKM > 0 {
			return latestKM, 0, nil
		}
	}

	// Fallback to employee's last completed session
	if empID != uuid.Nil {
		session, err := s.workRepo.FindLastCompletedSession(ctx, empID)
		if err == nil && session != nil {
			return session.EndKM, session.StartKM, nil
		}
	}

	return 0, 0, errors.New("لا توجد قراءة سابقة")
}

func (s *workService) UpdateWorkSession(ctx context.Context, sessionID uuid.UUID, req dto.UpdateWorkSessionRequest) (*domain.WorkSession, error) {
	session, err := s.workRepo.FindSessionByID(ctx, sessionID)
	if err != nil {
		return nil, errors.New("الشفت غير موجود")
	}

	// Update employee if provided
	if req.EmployeeID != "" {
		empID, err := uuid.Parse(req.EmployeeID)
		if err != nil {
			return nil, errors.New("معرف الموظف غير صالح")
		}
		_, err = s.empRepo.FindByID(ctx, empID)
		if err != nil {
			return nil, errors.New("الموظف غير موجود")
		}
		session.Employee = nil // clear preloaded old employee
		session.EmployeeID = &empID
	}

	// Update start_km if provided
	if req.StartKM > 0 {
		session.StartKM = req.StartKM
	}

	if req.EndKM > 0 {
		if req.EndKM <= session.StartKM {
			return nil, fmt.Errorf("قراءة عداد النهاية (%.2f) يجب أن تكون أكبر من قراءة البداية (%.2f)", req.EndKM, session.StartKM)
		}
		session.EndKM = req.EndKM
		session.Distance = req.EndKM - session.StartKM
	}
	// Update start_time and end_time if provided
	if req.StartTime != nil {
		session.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		session.EndTime = req.EndTime
		// Recalculate distance if end_km is already set
		if session.EndKM > session.StartKM {
			session.Distance = session.EndKM - session.StartKM
		}
	}

	session.OrdersCount = req.OrdersCount
	session.FuelCost = req.FuelCost
	if req.ApplicationType != "" {
		session.ApplicationType = req.ApplicationType
	}
	if req.Notes != "" {
		session.Notes = req.Notes
	}

	if err := s.workRepo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *workService) GetActiveSession(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error) {
	return s.workRepo.FindActiveSessionByEmployeeID(ctx, empID)
}

func (s *workService) GetLastCompletedSession(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error) {
	return s.workRepo.FindLastCompletedSession(ctx, empID)
}

func (s *workService) CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error) {
	return s.workRepo.CountTodaySessions(ctx, empID)
}

func (s *workService) CheckOilChange(ctx context.Context, empID uuid.UUID) (*dto.OilChangeCheckResponse, error) {
	emp, err := s.empRepo.FindByID(ctx, empID)
	if err != nil {
		return nil, errors.New("الموظف غير موجود")
	}

	distanceSinceOil := emp.TotalDistance - emp.LastOilChangeDistance
	interval := oilChangeInterval(emp.VehicleType)

	return &dto.OilChangeCheckResponse{
		NeedsOilChange:    distanceSinceOil >= interval,
		TotalDistance:     emp.TotalDistance,
		DistanceSinceOil:  distanceSinceOil,
		OilChangeInterval: interval,
		VehicleType:       emp.VehicleType,
	}, nil
}

func (s *workService) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*domain.WorkSession, error) {
	return s.workRepo.FindSessionByID(ctx, sessionID)
}

// Dashboard & Report Services
type DashboardService interface {
	GetStats(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardResponse, error)
}

type dashboardService struct {
	workRepo  repository.WorkRepository
	auditRepo repository.AuditRepository
	empRepo   repository.EmployeeRepository
}

func NewDashboardService(workRepo repository.WorkRepository, auditRepo repository.AuditRepository, empRepo repository.EmployeeRepository) DashboardService {
	return &dashboardService{workRepo: workRepo, auditRepo: auditRepo, empRepo: empRepo}
}

func (s *dashboardService) GetStats(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardResponse, error) {
	resp, err := s.workRepo.GetDashboardStats(ctx, branchID)
	if err != nil {
		return nil, err
	}

	// Count total employees (filtered by branch for supervisors)
	totalEmp, err := s.empRepo.CountAll(ctx, branchID)
	if err == nil {
		resp.TotalEmployees = totalEmp
	}

	// Fetch latest activities - filtered by branch for supervisors
	var latestLogs []domain.AuditLog
	if branchID != nil {
		latestLogs, _ = s.auditRepo.GetLatestLogsByBranch(ctx, *branchID, 10)
	} else {
		latestLogs, _ = s.auditRepo.GetLatestLogs(ctx, 10)
	}
	resp.LatestActivities = make([]dto.AuditLogResponse, len(latestLogs))
	for i, l := range latestLogs {
		resp.LatestActivities[i] = dto.AuditLogResponse{
			ID:        l.ID,
			AdminName: l.AdminName,
			Action:    l.Action,
			Details:   l.Details,
			IPAddress: l.IPAddress,
			CreatedAt: l.CreatedAt,
		}
	}

	return resp, nil
}

type ReportService interface {
	GetReports(ctx context.Context, filter dto.ReportFilter) ([]dto.WorkSessionDetailResponse, int64, error)
	GetDailyReport(ctx context.Context, branchID *uuid.UUID, date time.Time) (*dto.DailyReportResponse, error)
}

type reportService struct {
	workRepo repository.WorkRepository
}

func NewReportService(workRepo repository.WorkRepository) ReportService {
	return &reportService{workRepo: workRepo}
}

func (s *reportService) GetReports(ctx context.Context, filter dto.ReportFilter) ([]dto.WorkSessionDetailResponse, int64, error) {
	sessions, total, err := s.workRepo.GetReports(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	details := make([]dto.WorkSessionDetailResponse, len(sessions))
	for i, s := range sessions {
		durationStr := "جارٍ العمل"
		if s.EndTime != nil {
			dur := s.EndTime.Sub(s.StartTime)
			hours := int(dur.Hours())
			minutes := int(dur.Minutes()) % 60
			durationStr = fmt.Sprintf("%d ساعة و %d دقيقة", hours, minutes)
		}

		empName := "غير محدد"
		personalImg := ""
		natID := ""
		keyNumber := ""
		branchName := ""
		if s.Employee != nil {
			empName = s.Employee.Name
			personalImg = s.Employee.PersonalImage
			natID = s.Employee.NationalID
			keyNumber = s.Employee.KeyNumber
			if s.Employee.Branch != nil {
				branchName = s.Employee.Branch.Name
			}
		}

		empID := uuid.Nil
		if s.EmployeeID != nil {
			empID = *s.EmployeeID
		}

		details[i] = dto.WorkSessionDetailResponse{
			ID:              s.ID,
			EmployeeID:      empID,
			EmployeeName:    empName,
			PersonalImage:   personalImg,
			NationalID:      natID,
			KeyNumber:       keyNumber,
			BranchName:      branchName,
			StartTime:       s.StartTime,
			EndTime:         s.EndTime,
			WorkingDuration: durationStr,
			StartKM:         s.StartKM,
			EndKM:           s.EndKM,
			Distance:        s.Distance,
			OrdersCount:     s.OrdersCount,
			FuelCost:        s.FuelCost,
			ApplicationID:   s.ApplicationID,
			ApplicationType: s.ApplicationType,
			Notes:           s.Notes,
			Status:          s.Status,
		}
	}

	return details, total, nil
}

func (s *reportService) GetDailyReport(ctx context.Context, branchID *uuid.UUID, date time.Time) (*dto.DailyReportResponse, error) {
	rows, err := s.workRepo.GetDailyReport(ctx, branchID, date)
	if err != nil {
		return nil, err
	}

	resp := &dto.DailyReportResponse{
		Rows: rows,
	}

	// Calculate totals
	for _, r := range rows {
		resp.TotalOrders += r.TotalOrders
		resp.TotalKM += r.TotalKM
		resp.TotalFuel += r.TotalFuel
	}

	// Group by app type for summaries
	summaryMap := make(map[string]*dto.DailyAppSummary)
	for _, r := range rows {
		if _, ok := summaryMap[r.AppType]; !ok {
			summaryMap[r.AppType] = &dto.DailyAppSummary{
				AppType: r.AppType,
				AppName: r.AppName,
			}
		}
		s := summaryMap[r.AppType]
		s.Count++
		s.TotalOrders += r.TotalOrders
		s.TotalKM += r.TotalKM
		s.TotalFuel += r.TotalFuel
	}

	for _, s := range summaryMap {
		resp.AppSummaries = append(resp.AppSummaries, *s)
	}

	return resp, nil
}

// BranchService interface & impl
type BranchService interface {
	GetAll(ctx context.Context) ([]dto.BranchResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.BranchResponse, error)
	Create(ctx context.Context, req dto.CreateBranchRequest) (*dto.BranchResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.CreateBranchRequest) (*dto.BranchResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type branchService struct {
	branchRepo   repository.BranchRepository
	employeeRepo repository.EmployeeRepository
}

func NewBranchService(branchRepo repository.BranchRepository, employeeRepo repository.EmployeeRepository) BranchService {
	return &branchService{branchRepo: branchRepo, employeeRepo: employeeRepo}
}

func (s *branchService) countEmployeesForBranch(ctx context.Context, branchID uuid.UUID) (int64, error) {
	if s.employeeRepo == nil {
		return 0, nil
	}
	return s.employeeRepo.CountAll(ctx, &branchID)
}

func (s *branchService) GetAll(ctx context.Context) ([]dto.BranchResponse, error) {
	branches, err := s.branchRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.BranchResponse, len(branches))
	for i, b := range branches {
		var cnt int64
		if count, countErr := s.countEmployeesForBranch(ctx, b.ID); countErr == nil {
			cnt = count
		}
		resp[i] = dto.BranchResponse{ID: b.ID, Name: b.Name, EmployeeCount: cnt}
	}
	return resp, nil
}

func (s *branchService) GetByID(ctx context.Context, id uuid.UUID) (*dto.BranchResponse, error) {
	b, err := s.branchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("الفرع غير موجود")
	}
	var cnt int64
	if count, countErr := s.countEmployeesForBranch(ctx, id); countErr == nil {
		cnt = count
	}
	return &dto.BranchResponse{ID: b.ID, Name: b.Name, EmployeeCount: cnt}, nil
}

func (s *branchService) Create(ctx context.Context, req dto.CreateBranchRequest) (*dto.BranchResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("اسم الفرع مطلوب")
	}
	branch := &domain.Branch{Name: strings.TrimSpace(req.Name)}
	if err := s.branchRepo.Create(ctx, branch); err != nil {
		return nil, fmt.Errorf("فشل في إنشاء الفرع: %w", err)
	}
	return &dto.BranchResponse{ID: branch.ID, Name: branch.Name, EmployeeCount: 0}, nil
}

func (s *branchService) Update(ctx context.Context, id uuid.UUID, req dto.CreateBranchRequest) (*dto.BranchResponse, error) {
	branch, err := s.branchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("الفرع غير موجود")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("اسم الفرع مطلوب")
	}
	branch.Name = strings.TrimSpace(req.Name)
	if err := s.branchRepo.Update(ctx, branch); err != nil {
		return nil, fmt.Errorf("فشل في تحديث الفرع: %w", err)
	}
	var cnt int64
	if count, countErr := s.countEmployeesForBranch(ctx, id); countErr == nil {
		cnt = count
	}
	return &dto.BranchResponse{ID: branch.ID, Name: branch.Name, EmployeeCount: cnt}, nil
}

func (s *branchService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.branchRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("الفرع غير موجود")
	}
	return s.branchRepo.Delete(ctx, id)
}

// InventoryService interface & impl
type InventoryService interface {
	CreateItem(ctx context.Context, req dto.CreateInventoryItemRequest) (*domain.InventoryItem, error)
	UpdateItem(ctx context.Context, id uuid.UUID, req dto.UpdateInventoryItemRequest) (*domain.InventoryItem, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	GetAllItems(ctx context.Context, itemType string, branchID *uuid.UUID) ([]dto.InventoryItemWithStock, error)
	GetItemByID(ctx context.Context, id uuid.UUID) (*domain.InventoryItem, error)
	FindByBarcode(ctx context.Context, barcode string) (*domain.InventoryItem, error)
	AddStock(ctx context.Context, req dto.InventoryTransactionRequest, branchID *uuid.UUID) (*domain.InventoryTransaction, error)
	RemoveStock(ctx context.Context, req dto.InventoryTransactionRequest, branchID *uuid.UUID) (*domain.InventoryTransaction, error)
	DispenseOil(ctx context.Context, req dto.DispenseOilRequest, adminName string, branchID *uuid.UUID) (*domain.MaintenanceLog, error)
	GetTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]domain.InventoryTransaction, int64, error)
	DeleteAllTransactions(ctx context.Context) error
	CreatePurchaseInvoice(ctx context.Context, req dto.CreatePurchaseInvoiceRequest, branchID *uuid.UUID, adminName string) (*domain.PurchaseInvoice, error)
	GetPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]domain.PurchaseInvoice, int64, error)
	GetPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseInvoice, error)
	DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error
}

type inventoryService struct {
	invRepo         repository.InventoryRepository
	empRepo         repository.EmployeeRepository
	maintenanceRepo repository.MaintenanceRepository
}

func NewInventoryService(invRepo repository.InventoryRepository, empRepo repository.EmployeeRepository, maintenanceRepo repository.MaintenanceRepository) InventoryService {
	return &inventoryService{invRepo: invRepo, empRepo: empRepo, maintenanceRepo: maintenanceRepo}
}

func (s *inventoryService) CreateItem(ctx context.Context, req dto.CreateInventoryItemRequest) (*domain.InventoryItem, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("اسم الصنف مطلوب")
	}
	if req.Type != "oil" && req.Type != "spare_part" {
		return nil, errors.New("نوع الصنف يجب أن يكون 'oil' أو 'spare_part'")
	}

	item := &domain.InventoryItem{
		Name:     strings.TrimSpace(req.Name),
		Type:     req.Type,
		Unit:     req.Unit,
		Barcode:  strings.TrimSpace(req.Barcode),
		Quantity: req.Quantity,
		MinAlert: req.MinAlert,
		Notes:    req.Notes,
	}
	if item.MinAlert == 0 {
		item.MinAlert = 5
	}

	if err := s.invRepo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("فشل في إضافة الصنف: %w", err)
	}
	return item, nil
}

func (s *inventoryService) UpdateItem(ctx context.Context, id uuid.UUID, req dto.UpdateInventoryItemRequest) (*domain.InventoryItem, error) {
	item, err := s.invRepo.FindItemByID(ctx, id)
	if err != nil {
		return nil, errors.New("الصنف غير موجود")
	}

	if req.Name != "" {
		item.Name = strings.TrimSpace(req.Name)
	}
	if req.Type != "" {
		if req.Type != "oil" && req.Type != "spare_part" {
			return nil, errors.New("نوع الصنف يجب أن يكون 'oil' أو 'spare_part'")
		}
		item.Type = req.Type
	}
	if req.Unit != "" {
		item.Unit = req.Unit
	}
	if req.Barcode != "" {
		item.Barcode = strings.TrimSpace(req.Barcode)
	}
	item.Quantity = req.Quantity
	if req.MinAlert > 0 {
		item.MinAlert = req.MinAlert
	}
	if req.Notes != "" {
		item.Notes = req.Notes
	}

	if err := s.invRepo.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("فشل في تحديث الصنف: %w", err)
	}
	return item, nil
}

func (s *inventoryService) DeleteItem(ctx context.Context, id uuid.UUID) error {
	_, err := s.invRepo.FindItemByID(ctx, id)
	if err != nil {
		return errors.New("الصنف غير موجود")
	}
	return s.invRepo.DeleteItem(ctx, id)
}

func (s *inventoryService) GetAllItems(ctx context.Context, itemType string, branchID *uuid.UUID) ([]dto.InventoryItemWithStock, error) {
	items, err := s.invRepo.FindAllItems(ctx, itemType)
	if err != nil {
		return nil, err
	}

	// Get branch-specific stock
	stock, err := s.invRepo.GetStockByBranch(ctx, branchID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.InventoryItemWithStock, len(items))
	for i := range items {
		result[i] = dto.InventoryItemWithStock{
			InventoryItem:  &items[i],
			BranchQuantity: stock[items[i].ID],
		}
	}
	return result, nil
}

// AttendanceService interface & impl
type AttendanceService interface {
	GetAttendance(ctx context.Context, date string, branchID *uuid.UUID) ([]domain.AttendanceInfo, error)
	ToggleAttendance(ctx context.Context, adminID uuid.UUID, employeeID uuid.UUID, date string, status string, note string) (*domain.AttendanceInfo, error)
}

type attendanceService struct {
	attendanceRepo repository.AttendanceRepository
	employeeRepo   repository.EmployeeRepository
	workRepo       repository.WorkRepository
}

func NewAttendanceService(attendanceRepo repository.AttendanceRepository, employeeRepo repository.EmployeeRepository, workRepo repository.WorkRepository) AttendanceService {
	return &attendanceService{
		attendanceRepo: attendanceRepo,
		employeeRepo:   employeeRepo,
		workRepo:       workRepo,
	}
}

func (s *attendanceService) GetAttendance(ctx context.Context, date string, branchID *uuid.UUID) ([]domain.AttendanceInfo, error) {
	// Get all employees (active only, not deleted)
	emps, _, err := s.employeeRepo.FindAll(ctx, dto.EmployeeFilter{BranchID: branchID, Limit: 500})
	if err != nil {
		return nil, err
	}

	// Get existing attendance records for this date
	existingRecords, err := s.attendanceRepo.FindByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	recordMap := make(map[uuid.UUID]*domain.Attendance)
	for i := range existingRecords {
		recordMap[existingRecords[i].EmployeeID] = &existingRecords[i]
	}

	result := make([]domain.AttendanceInfo, 0, len(emps))
	for _, emp := range emps {
		branchName := ""
		if emp.Branch != nil {
			branchName = emp.Branch.Name
		}

		info := domain.AttendanceInfo{
			EmployeeID:   emp.ID,
			EmployeeName: emp.Name,
			NationalID:   emp.NationalID,
			BranchName:   branchName,
			VehicleType:  emp.VehicleType,
			Status:       "absent",
		}

		// Check existing record
		if rec, ok := recordMap[emp.ID]; ok {
			info.Status = rec.Status
			info.Note = rec.Note
		}

		result = append(result, info)
	}

	return result, nil
}

func (s *attendanceService) ToggleAttendance(ctx context.Context, adminID uuid.UUID, employeeID uuid.UUID, date string, status string, note string) (*domain.AttendanceInfo, error) {
	attendance := &domain.Attendance{
		EmployeeID: employeeID,
		Date:       date,
		Status:     status,
		Note:       note,
	}

	if err := s.attendanceRepo.Upsert(ctx, attendance); err != nil {
		return nil, err
	}

	// Fetch employee info
	emp, err := s.employeeRepo.FindByID(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	branchName := ""
	if emp.Branch != nil {
		branchName = emp.Branch.Name
	}

	return &domain.AttendanceInfo{
		EmployeeID:   emp.ID,
		EmployeeName: emp.Name,
		NationalID:   emp.NationalID,
		BranchName:   branchName,
		VehicleType:  emp.VehicleType,
		Status:       status,
		Note:         note,
	}, nil
}

func (s *inventoryService) GetItemByID(ctx context.Context, id uuid.UUID) (*domain.InventoryItem, error) {
	return s.invRepo.FindItemByID(ctx, id)
}

func (s *inventoryService) FindByBarcode(ctx context.Context, barcode string) (*domain.InventoryItem, error) {
	return s.invRepo.FindByBarcode(ctx, barcode)
}

func (s *inventoryService) AddStock(ctx context.Context, req dto.InventoryTransactionRequest, branchID *uuid.UUID) (*domain.InventoryTransaction, error) {
	itemID, err := uuid.Parse(req.ItemID)
	if err != nil {
		return nil, errors.New("معرف الصنف غير صالح")
	}

	item, err := s.invRepo.FindItemByID(ctx, itemID)
	if err != nil {
		return nil, errors.New("الصنف غير موجود")
	}

	// Stock is now calculated from transactions per branch — no need to update item.Quantity

	tx := &domain.InventoryTransaction{
		ItemID:   itemID,
		Type:     "in",
		Quantity: req.Quantity,
		BranchID: branchID,
		Notes:    req.Notes,
	}
	if req.EmployeeID != nil {
		if empID, err := uuid.Parse(*req.EmployeeID); err == nil {
			tx.EmployeeID = &empID
		}
	}

	if err := s.invRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("فشل في تسجيل الحركة: %w", err)
	}

	tx.Item = item
	return tx, nil
}

func (s *inventoryService) RemoveStock(ctx context.Context, req dto.InventoryTransactionRequest, branchID *uuid.UUID) (*domain.InventoryTransaction, error) {
	itemID, err := uuid.Parse(req.ItemID)
	if err != nil {
		return nil, errors.New("معرف الصنف غير صالح")
	}

	item, err := s.invRepo.FindItemByID(ctx, itemID)
	if err != nil {
		return nil, errors.New("الصنف غير موجود")
	}

	// Check branch-specific stock
	currentStock, err := s.invRepo.GetItemStock(ctx, itemID, branchID)
	if err != nil {
		return nil, fmt.Errorf("فشل في التحقق من المخزون: %w", err)
	}
	if currentStock < req.Quantity {
		return nil, fmt.Errorf("الكمية غير كافية في الفرع. المتاح: %d %s", currentStock, item.Unit)
	}

	// Stock is now calculated from transactions per branch — no need to update item.Quantity

	tx := &domain.InventoryTransaction{
		ItemID:   itemID,
		Type:     "out",
		Quantity: req.Quantity,
		BranchID: branchID,
		Notes:    req.Notes,
	}
	if req.EmployeeID != nil {
		if empID, err := uuid.Parse(*req.EmployeeID); err == nil {
			tx.EmployeeID = &empID
		}
	}

	if err := s.invRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("فشل في تسجيل الحركة: %w", err)
	}

	tx.Item = item
	return tx, nil
}

func (s *inventoryService) DispenseOil(ctx context.Context, req dto.DispenseOilRequest, adminName string, branchID *uuid.UUID) (*domain.MaintenanceLog, error) {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("معرف الموظف غير صالح")
	}

	emp, err := s.empRepo.FindByID(ctx, empID)
	if err != nil {
		return nil, errors.New("الموظف غير موجود")
	}

	// Find available oil items and their branch stock
	oilItems, err := s.invRepo.FindOilItemsWithStock(ctx, branchID)
	if err != nil || len(oilItems) == 0 {
		return nil, errors.New("لا يوجد زيت متاح في المخزن. يرجى إضافة زيت أولاً")
	}

	// Calculate total available oil for this branch
	totalOil := 0
	for _, item := range oilItems {
		stock, _ := s.invRepo.GetItemStock(ctx, item.ID, branchID)
		totalOil += stock
	}
	if totalOil < req.Quantity {
		return nil, fmt.Errorf("الكمية غير كافية. المتاح: %d جركن زيت", totalOil)
	}

	// Deduct from oil items (per-branch, recorded as transactions)
	remaining := req.Quantity
	for _, item := range oilItems {
		if remaining <= 0 {
			break
		}
		stock, _ := s.invRepo.GetItemStock(ctx, item.ID, branchID)
		if stock > 0 {
			deduct := stock
			if deduct > remaining {
				deduct = remaining
			}
			// No need to update item.Quantity — stock is calculated from transactions

			tx := &domain.InventoryTransaction{
				ItemID:     item.ID,
				Type:       "out",
				Quantity:   deduct,
				EmployeeID: &empID,
				BranchID:   branchID,
				Notes:      fmt.Sprintf("صرف زيت للمندوب: %s - %d جركن", emp.Name, deduct),
			}
			if txErr := s.invRepo.CreateTransaction(ctx, tx); txErr != nil {
				return nil, fmt.Errorf("فشل في تسجيل حركة صرف الزيت: %w", txErr)
			}

			remaining -= deduct
		}
	}

	// Create maintenance log
	maintenanceLog := &domain.MaintenanceLog{
		EmployeeID: &empID,
		Type:       "oil_change",
		Details:    fmt.Sprintf("تغيير زيت - %d جركن", req.Quantity),
		DistanceAt: emp.TotalDistance,
		AdminName:  adminName,
	}

	if err := s.maintenanceRepo.CreateLog(ctx, maintenanceLog); err != nil {
		return nil, fmt.Errorf("فشل في تسجيل الصيانة: %w", err)
	}

	// Reset oil change counter
	emp.LastOilChangeDistance = emp.TotalDistance
	if updateErr := s.empRepo.Update(ctx, emp); updateErr != nil {
		return nil, fmt.Errorf("فشل في تحديث عداد تغيير الزيت للموظف: %w", updateErr)
	}

	maintenanceLog.Employee = emp
	return maintenanceLog, nil
}

func (s *inventoryService) GetTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]domain.InventoryTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.invRepo.FindTransactions(ctx, itemID, branchID, page, limit)
}

func (s *inventoryService) DeleteAllTransactions(ctx context.Context) error {
	return s.invRepo.DeleteAllTransactions(ctx)
}

func (s *inventoryService) CreatePurchaseInvoice(ctx context.Context, req dto.CreatePurchaseInvoiceRequest, branchID *uuid.UUID, adminName string) (*domain.PurchaseInvoice, error) {
	if strings.TrimSpace(req.SupplierName) == "" {
		return nil, errors.New("اسم المورد مطلوب")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("يجب إضافة صنف واحد على الأقل في الفاتورة")
	}

	invoiceNumber := strings.TrimSpace(req.InvoiceNumber)
	if invoiceNumber == "" {
		invoiceNumber = fmt.Sprintf("INV-%s-%d", time.Now().Format("20060102"), time.Now().Unix()%10000)
	}

	invoiceDate := time.Now()
	if req.InvoiceDate != nil && !req.InvoiceDate.IsZero() {
		invoiceDate = *req.InvoiceDate
	}

	var totalAmount float64
	invoiceItems := make([]domain.PurchaseInvoiceItem, 0, len(req.Items))
	stockTxs := make([]*domain.InventoryTransaction, 0, len(req.Items))

	for _, itemReq := range req.Items {
		itemID, err := uuid.Parse(itemReq.ItemID)
		if err != nil {
			return nil, errors.New("معرف الصنف غير صالح")
		}

		item, err := s.invRepo.FindItemByID(ctx, itemID)
		if err != nil {
			return nil, fmt.Errorf("الصنف المحدد غير موجود: %s", itemReq.ItemID)
		}

		if itemReq.Quantity <= 0 {
			return nil, fmt.Errorf("كمية الصنف %s يجب أن تكون أكبر من الصفر", item.Name)
		}

		rowTotal := float64(itemReq.Quantity) * itemReq.UnitPrice
		totalAmount += rowTotal

		invoiceItems = append(invoiceItems, domain.PurchaseInvoiceItem{
			ItemID:     itemID,
			Quantity:   itemReq.Quantity,
			UnitPrice:  itemReq.UnitPrice,
			TotalPrice: rowTotal,
			Notes:      itemReq.Notes,
		})

		// Auto stock-in transaction
		stockTxs = append(stockTxs, &domain.InventoryTransaction{
			ItemID:   itemID,
			Type:     "in",
			Quantity: itemReq.Quantity,
			BranchID: branchID,
			Notes:    fmt.Sprintf("توريد مشتريات (فاتورة رقم: %s - مورد: %s)", invoiceNumber, req.SupplierName),
		})
	}

	subtotal := totalAmount
	discount := req.Discount
	taxRate := req.TaxRate
	taxAmount := req.TaxAmount
	if taxAmount == 0 && taxRate > 0 {
		taxAmount = (subtotal - discount) * (taxRate / 100.0)
	}
	grandTotal := subtotal - discount + taxAmount
	if grandTotal < 0 {
		grandTotal = 0
	}
	if req.TotalAmount > 0 {
		grandTotal = req.TotalAmount
	}

	invoice := &domain.PurchaseInvoice{
		InvoiceNumber: invoiceNumber,
		SupplierName:  strings.TrimSpace(req.SupplierName),
		InvoiceDate:   invoiceDate,
		Subtotal:      subtotal,
		Discount:      discount,
		TaxRate:       taxRate,
		TaxAmount:     taxAmount,
		TotalAmount:   grandTotal,
		BranchID:      branchID,
		CreatedByName: adminName,
		Notes:         req.Notes,
		Items:         invoiceItems,
	}

	if err := s.invRepo.CreatePurchaseInvoice(ctx, invoice, stockTxs); err != nil {
		return nil, fmt.Errorf("فشل في حفظ فاتورة المشتريات: %w", err)
	}

	return invoice, nil
}

func (s *inventoryService) GetPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]domain.PurchaseInvoice, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.invRepo.FindPurchaseInvoices(ctx, branchID, search, page, limit)
}

func (s *inventoryService) GetPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseInvoice, error) {
	return s.invRepo.FindPurchaseInvoiceByID(ctx, id)
}

func (s *inventoryService) DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error {
	return s.invRepo.DeletePurchaseInvoice(ctx, id)
}

// MaintenanceService interface & impl
type MaintenanceService interface {
	GetEmployeeLogs(ctx context.Context, empID uuid.UUID, limit int) ([]domain.MaintenanceLog, error)
	GetAllLogs(ctx context.Context, page, limit int) ([]domain.MaintenanceLog, int64, error)
}

type maintenanceService struct {
	maintenanceRepo repository.MaintenanceRepository
}

func NewMaintenanceService(maintenanceRepo repository.MaintenanceRepository) MaintenanceService {
	return &maintenanceService{maintenanceRepo: maintenanceRepo}
}

func (s *maintenanceService) GetEmployeeLogs(ctx context.Context, empID uuid.UUID, limit int) ([]domain.MaintenanceLog, error) {
	return s.maintenanceRepo.FindByEmployeeID(ctx, empID, limit)
}

func (s *maintenanceService) GetAllLogs(ctx context.Context, page, limit int) ([]domain.MaintenanceLog, int64, error) {
	return s.maintenanceRepo.FindAll(ctx, page, limit)
}

// Investigation Service
type InvestigationService interface {
	Create(ctx context.Context, req dto.CreateInvestigationRequest, supervisorID uuid.UUID) (*dto.InvestigationResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateInvestigationRequest) (*dto.InvestigationResponse, error)
	GetAll(ctx context.Context, branchID *uuid.UUID) ([]dto.InvestigationResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.InvestigationResponse, error)
	Approve(ctx context.Context, id uuid.UUID, adminID uuid.UUID) (*dto.InvestigationResponse, error)
	Reject(ctx context.Context, id uuid.UUID, adminID uuid.UUID) (*dto.InvestigationResponse, error)
	GetPendingApprovalCount(ctx context.Context) (int64, error)
}

type investigationService struct {
	investigationRepo repository.InvestigationRepository
	employeeRepo      repository.EmployeeRepository
	adminRepo         repository.AdminRepository
}

func NewInvestigationService(investigationRepo repository.InvestigationRepository, employeeRepo repository.EmployeeRepository, adminRepo repository.AdminRepository) InvestigationService {
	return &investigationService{investigationRepo: investigationRepo, employeeRepo: employeeRepo, adminRepo: adminRepo}
}

// toResponse يحوّل كائن التحقيق إلى استجابة موحّدة مع أسماء الموظف والمشرف وحالة الموافقة/الرفض.
func (s *investigationService) toResponse(inv *domain.Investigation) *dto.InvestigationResponse {
	supervisorName := ""
	if inv.Supervisor != nil {
		supervisorName = inv.Supervisor.Name
	}
	empName := ""
	if inv.Employee != nil {
		empName = inv.Employee.Name
	}

	var questions []string
	var answers []string
	var items []string
	var images []string
	json.Unmarshal([]byte(inv.Questions), &questions)
	json.Unmarshal([]byte(inv.Answers), &answers)
	json.Unmarshal([]byte(inv.Items), &items)
	json.Unmarshal([]byte(inv.Images), &images)

	g := false
	if inv.IsGuilty != nil {
		g = *inv.IsGuilty
	}

	return &dto.InvestigationResponse{
		ID:                 inv.ID,
		EmployeeID:         inv.EmployeeID,
		EmployeeName:       empName,
		NationalID:         inv.NationalID,
		SupervisorID:       inv.SupervisorID,
		SupervisorName:     supervisorName,
		Type:               inv.Type,
		Questions:          questions,
		Answers:            answers,
		ReportText:         inv.ReportText,
		Images:             images,
		Amount:             inv.Amount,
		StartDate:          inv.StartDate,
		EndDate:            inv.EndDate,
		Items:              items,
		IsGuilty:           g,
		Notes:              inv.Notes,
		DeductionMonth:     inv.DeductionMonth,
		Status:             inv.Status,
		ApprovedByName:     inv.ApprovedByName,
		ApprovedByUsername: inv.ApprovedByUsername,
		RejectedByName:     inv.RejectedByName,
		RejectedByUsername: inv.RejectedByUsername,
		ApprovedAt:         inv.ApprovedAt,
		RejectedAt:         inv.RejectedAt,
		CreatedAt:          inv.CreatedAt,
	}
}

func (s *investigationService) Create(ctx context.Context, req dto.CreateInvestigationRequest, supervisorID uuid.UUID) (*dto.InvestigationResponse, error) {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("معرف الموظف غير صالح")
	}

	emp, err := s.employeeRepo.FindByID(ctx, empID)
	if err != nil || emp == nil {
		return nil, errors.New("الموظف غير موجود")
	}

	questionsJSON, err := json.Marshal(req.Questions)
	if err != nil {
		return nil, fmt.Errorf("فشل في معالجة بيانات الأسئلة: %w", err)
	}
	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		return nil, fmt.Errorf("فشل في معالجة بيانات الإجابات: %w", err)
	}
	itemsJSON, err := json.Marshal(req.Items)
	if err != nil {
		return nil, fmt.Errorf("فشل في معالجة بيانات الأصناف: %w", err)
	}
	imagesJSON, err := json.Marshal(req.Images)
	if err != nil {
		return nil, fmt.Errorf("فشل في معالجة بيانات الصور: %w", err)
	}
	isGuilty := req.IsGuilty

	reportType := req.Type
	if reportType == "" {
		reportType = "investigation"
	}

	investigation := &domain.Investigation{
		EmployeeID:   empID,
		SupervisorID: supervisorID,
		NationalID:   emp.NationalID,
		Type:         reportType,
		Questions:    string(questionsJSON),
		Answers:      string(answersJSON),
		ReportText:   req.ReportText,
		Images:       string(imagesJSON),
		Amount:       req.Amount,
		Items:        string(itemsJSON),
		IsGuilty:     &isGuilty,
		Notes:        req.Notes,
	}

	// Parse dates if provided
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			investigation.StartDate = &t
		}
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			investigation.EndDate = &t
		}
	}

	if err := s.investigationRepo.Create(ctx, investigation); err != nil {
		return nil, fmt.Errorf("فشل في حفظ التقرير: %w", err)
	}

	// Reload with relations
	full, err := s.investigationRepo.FindByID(ctx, investigation.ID)
	if err != nil {
		return nil, fmt.Errorf("فشل في استرجاع التقرير: %w", err)
	}

	return s.toResponse(full), nil
}

func (s *investigationService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateInvestigationRequest) (*dto.InvestigationResponse, error) {
	inv, err := s.investigationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("التقرير غير موجود")
	}

	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("معرف الموظف غير صالح")
	}

	emp, err := s.employeeRepo.FindByID(ctx, empID)
	if err != nil || emp == nil {
		return nil, errors.New("الموظف غير موجود")
	}

	questionsJSON, err := json.Marshal(req.Questions)
	if err != nil {
		return nil, fmt.Errorf("فشل في معالجة بيانات الأسئلة: %w", err)
	}
	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		return nil, fmt.Errorf("فشل في معالجة بيانات الإجابات: %w", err)
	}
	itemsJSON, err := json.Marshal(req.Items)
	if err != nil {
		return nil, fmt.Errorf("فشل في معالجة بيانات الأصناف: %w", err)
	}
	imagesJSON, err := json.Marshal(req.Images)
	if err != nil {
		return nil, fmt.Errorf("فشل في معالجة بيانات الصور: %w", err)
	}

	inv.EmployeeID = empID
	inv.NationalID = emp.NationalID
	inv.Questions = string(questionsJSON)
	inv.Answers = string(answersJSON)
	inv.ReportText = req.ReportText
	inv.Images = string(imagesJSON)
	inv.Amount = req.Amount
	inv.Items = string(itemsJSON)
	inv.Notes = req.Notes
	inv.DeductionMonth = req.DeductionMonth

	if req.Type != "" {
		inv.Type = req.Type
	}

	isGuilty := req.IsGuilty
	inv.IsGuilty = &isGuilty

	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			inv.StartDate = &t
		}
	} else {
		inv.StartDate = nil
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			inv.EndDate = &t
		}
	} else {
		inv.EndDate = nil
	}

	if err := s.investigationRepo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("فشل في تحديث التقرير: %w", err)
	}

	// Reload with relations
	full, err := s.investigationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("فشل في استرجاع التقرير: %w", err)
	}

	return s.toResponse(full), nil
}

func (s *investigationService) GetAll(ctx context.Context, branchID *uuid.UUID) ([]dto.InvestigationResponse, error) {
	investigations, err := s.investigationRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.InvestigationResponse, 0, len(investigations))
	for _, inv := range investigations {
		// Filter by branch for supervisors
		if branchID != nil {
			if inv.Employee == nil || inv.Employee.BranchID == nil || *inv.Employee.BranchID != *branchID {
				continue
			}
		}

		result = append(result, *s.toResponse(&inv))
	}
	return result, nil
}

func (s *investigationService) GetByID(ctx context.Context, id uuid.UUID) (*dto.InvestigationResponse, error) {
	inv, err := s.investigationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("التحقيق غير موجود")
	}

	return s.toResponse(inv), nil
}

func (s *investigationService) Approve(ctx context.Context, id uuid.UUID, adminID uuid.UUID) (*dto.InvestigationResponse, error) {
	return s.setApproval(ctx, id, adminID, "approved")
}

func (s *investigationService) Reject(ctx context.Context, id uuid.UUID, adminID uuid.UUID) (*dto.InvestigationResponse, error) {
	return s.setApproval(ctx, id, adminID, "rejected")
}

func (s *investigationService) GetPendingApprovalCount(ctx context.Context) (int64, error) {
	return s.investigationRepo.CountPendingApprovals(ctx)
}

func (s *investigationService) setApproval(ctx context.Context, id uuid.UUID, adminID uuid.UUID, status string) (*dto.InvestigationResponse, error) {
	inv, err := s.investigationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("التقرير غير موجود")
	}

	if inv.Type != "advance" && inv.Type != "internet_advance" {
		return nil, errors.New("الموافقة أو الرفض متاح فقط للسلفة وسلفة النت")
	}

	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, errors.New("المدير غير موجود")
	}

	now := time.Now()
	inv.Status = status
	if status == "approved" {
		inv.ApprovedByName = admin.Name
		inv.ApprovedByUsername = admin.Username
		inv.ApprovedAt = &now
		inv.RejectedByName = ""
		inv.RejectedByUsername = ""
		inv.RejectedAt = nil
	} else {
		inv.RejectedByName = admin.Name
		inv.RejectedByUsername = admin.Username
		inv.RejectedAt = &now
		inv.ApprovedByName = ""
		inv.ApprovedByUsername = ""
		inv.ApprovedAt = nil
	}

	if err := s.investigationRepo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("فشل في تحديث حالة السلفة: %w", err)
	}

	return s.GetByID(ctx, id)
}

// Custody Service
type CustodyService interface {
	List(ctx context.Context, branchID *uuid.UUID) ([]dto.CustodyDayResponse, error)
	Create(ctx context.Context, req dto.CreateCustodyDayRequest, admin *domain.Admin) (*dto.CustodyDayResponse, error)
	AddAmount(ctx context.Context, req dto.AddCustodyAmountRequest, admin *domain.Admin) (*dto.CustodyDayResponse, error)
	AddExpense(ctx context.Context, dayID uuid.UUID, branchID *uuid.UUID, req dto.CreateCustodyExpenseRequest, admin *domain.Admin) (*dto.CustodyDayResponse, error)
	DeleteExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, admin *domain.Admin) (*dto.CustodyDayResponse, error)
	GetLogs(ctx context.Context, filter dto.CustodyLogFilter) ([]domain.CustodyLog, int64, error)
	DeleteLog(ctx context.Context, logID uuid.UUID, branchID *uuid.UUID, admin *domain.Admin) error
}

type custodyService struct {
	repo repository.CustodyRepository
}

func NewCustodyService(repo repository.CustodyRepository) CustodyService {
	return &custodyService{repo: repo}
}

var custodyCategories = map[string]bool{
	"fuel":        true,
	"license":     true,
	"spare_parts": true,
	"other":       true,
}

func (s *custodyService) toResponse(day *domain.CustodyDay) *dto.CustodyDayResponse {
	resp := &dto.CustodyDayResponse{
		ID:             day.ID,
		BranchID:       day.BranchID,
		Date:           day.Date,
		OpeningBalance: day.OpeningBalance,
		AddedAmount:    day.AddedAmount,
		CustodyValue:   day.OpeningBalance + day.AddedAmount,
		TotalExpenses:  0,
		ClosingBalance: day.ClosingBalance,
		Expenses:       make([]dto.CustodyExpenseResponse, 0, len(day.Expenses)),
		CreatedAt:      day.CreatedAt,
	}

	if day.Branch != nil {
		resp.BranchName = day.Branch.Name
	}

	for _, e := range day.Expenses {
		resp.Expenses = append(resp.Expenses, dto.CustodyExpenseResponse{
			ID:            e.ID,
			CustodyDayID:  e.CustodyDayID,
			Category:      e.Category,
			Amount:        e.Amount,
			RecipientName: e.RecipientName,
			CreatedAt:     e.CreatedAt,
		})
		resp.TotalExpenses += e.Amount

		switch e.Category {
		case "fuel":
			resp.Totals.Fuel += e.Amount
		case "license":
			resp.Totals.License += e.Amount
		case "spare_parts":
			resp.Totals.SpareParts += e.Amount
		default:
			resp.Totals.Other += e.Amount
		}
	}

	return resp
}

func (s *custodyService) recomputeClosing(day *domain.CustodyDay) {
	total := 0.0
	for _, e := range day.Expenses {
		total += e.Amount
	}
	day.ClosingBalance = day.OpeningBalance + day.AddedAmount - total
}

func (s *custodyService) List(ctx context.Context, branchID *uuid.UUID) ([]dto.CustodyDayResponse, error) {
	days, err := s.repo.FindAll(ctx, branchID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.CustodyDayResponse, 0, len(days))
	for i := range days {
		result = append(result, *s.toResponse(&days[i]))
	}
	return result, nil
}

func (s *custodyService) Create(ctx context.Context, req dto.CreateCustodyDayRequest, admin *domain.Admin) (*dto.CustodyDayResponse, error) {
	if req.BranchID == nil {
		return nil, errors.New("يجب تحديد الفرع")
	}

	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return nil, errors.New("التاريخ غير صالح")
	}

	existing, _ := s.repo.FindDayByDate(ctx, req.BranchID, req.Date)
	if existing != nil {
		return nil, errors.New("توجد عهدة بالفعل لهذا اليوم - يرجى استخدام أمر 'إضافة عهدة جديدة' لإضافة مبلغ إضافي")
	}

	opening := 0.0
	if last, err := s.repo.FindLastDay(ctx, req.BranchID); err == nil && last != nil {
		opening = last.ClosingBalance
	}

	day := &domain.CustodyDay{
		BranchID:       req.BranchID,
		Date:           req.Date,
		OpeningBalance: opening,
		AddedAmount:    req.AddedAmount,
		ClosingBalance: opening + req.AddedAmount,
	}

	if err := s.repo.CreateDay(ctx, day); err != nil {
		return nil, fmt.Errorf("فشل في فتح العهدة: %w", err)
	}

	// Create log entry
	var adminID *uuid.UUID
	adminName, adminUsername := "النظام", "system"
	if admin != nil {
		adminID = &admin.ID
		adminName = admin.Name
		adminUsername = admin.Username
	}

	logEntry := &domain.CustodyLog{
		BranchID:      req.BranchID,
		CustodyDayID:  day.ID,
		Date:          day.Date,
		ActionType:    "ADD_CUSTODY",
		Category:      "custody",
		Amount:        req.AddedAmount,
		Description:   fmt.Sprintf("فتح عهدة جديدة لمبلغ %.2f ريال", req.AddedAmount),
		AdminID:       adminID,
		AdminName:     adminName,
		AdminUsername: adminUsername,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateLog(ctx, logEntry)

	full, err := s.repo.FindDayByID(ctx, day.ID)
	if err != nil {
		return nil, fmt.Errorf("فشل في استرجاع العهدة: %w", err)
	}

	return s.toResponse(full), nil
}

func (s *custodyService) AddAmount(ctx context.Context, req dto.AddCustodyAmountRequest, admin *domain.Admin) (*dto.CustodyDayResponse, error) {
	day, err := s.repo.FindDayByID(ctx, req.CustodyDayID)
	if err != nil {
		return nil, errors.New("العهدة غير موجودة")
	}
	if req.BranchID != nil && (day.BranchID == nil || *day.BranchID != *req.BranchID) {
		return nil, errors.New("ليس لديك صلاحية الوصول لهذه العهدة - الفرع غير مطابق")
	}
	if req.AddedAmount <= 0 {
		return nil, errors.New("المبلغ المضاف يجب أن يكون أكبر من 0")
	}

	day.AddedAmount += req.AddedAmount
	s.recomputeClosing(day)

	if err := s.repo.UpdateDay(ctx, day); err != nil {
		return nil, fmt.Errorf("فشل في زيادة مبلغ العهدة: %w", err)
	}

	// Create log entry
	var adminID *uuid.UUID
	adminName, adminUsername := "النظام", "system"
	if admin != nil {
		adminID = &admin.ID
		adminName = admin.Name
		adminUsername = admin.Username
	}

	logEntry := &domain.CustodyLog{
		BranchID:      day.BranchID,
		CustodyDayID:  day.ID,
		Date:          day.Date,
		ActionType:    "ADD_CUSTODY",
		Category:      "custody",
		Amount:        req.AddedAmount,
		Description:   fmt.Sprintf("إضافة مبلغ عهدة إضافي بقيمة %.2f ريال", req.AddedAmount),
		AdminID:       adminID,
		AdminName:     adminName,
		AdminUsername: adminUsername,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateLog(ctx, logEntry)

	full, err := s.repo.FindDayByID(ctx, day.ID)
	if err != nil {
		return nil, fmt.Errorf("فشل في استرجاع العهدة: %w", err)
	}

	return s.toResponse(full), nil
}

func (s *custodyService) AddExpense(ctx context.Context, dayID uuid.UUID, branchID *uuid.UUID, req dto.CreateCustodyExpenseRequest, admin *domain.Admin) (*dto.CustodyDayResponse, error) {
	if !custodyCategories[req.Category] {
		return nil, errors.New("بند المصروف غير صالح")
	}

	day, err := s.repo.FindDayByID(ctx, dayID)
	if err != nil {
		return nil, errors.New("العهدة غير موجودة")
	}
	if branchID != nil && (day.BranchID == nil || *day.BranchID != *branchID) {
		return nil, errors.New("ليس لديك صلاحية الوصول لهذه العهدة - الفرع غير مطابق")
	}

	var adminID *uuid.UUID
	adminName, adminUsername := "النظام", "system"
	if admin != nil {
		adminID = &admin.ID
		adminName = admin.Name
		adminUsername = admin.Username
	}

	expense := &domain.CustodyExpense{
		CustodyDayID:      dayID,
		Category:          req.Category,
		Amount:            req.Amount,
		RecipientName:     strings.TrimSpace(req.RecipientName),
		CreatedByID:       adminID,
		CreatedByName:     adminName,
		CreatedByUsername: adminUsername,
		CreatedAt:         time.Now(),
	}

	if err := s.repo.CreateExpense(ctx, expense); err != nil {
		return nil, fmt.Errorf("فشل في إضافة المصروف: %w", err)
	}

	// Create log entry
	logEntry := &domain.CustodyLog{
		BranchID:      day.BranchID,
		CustodyDayID:  day.ID,
		Date:          day.Date,
		ActionType:    "ADD_EXPENSE",
		Category:      req.Category,
		Amount:        req.Amount,
		RecipientName: strings.TrimSpace(req.RecipientName),
		Description:   fmt.Sprintf("إضافة مصروف (%s) بقيمة %.2f ريال للمستلم: %s", req.Category, req.Amount, req.RecipientName),
		AdminID:       adminID,
		AdminName:     adminName,
		AdminUsername: adminUsername,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateLog(ctx, logEntry)

	day, err = s.repo.FindDayByID(ctx, dayID)
	if err != nil {
		return nil, fmt.Errorf("فشل في استرجاع العهدة: %w", err)
	}

	s.recomputeClosing(day)
	_ = s.repo.UpdateDay(ctx, day)

	return s.toResponse(day), nil
}

func (s *custodyService) DeleteExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, admin *domain.Admin) (*dto.CustodyDayResponse, error) {
	expense, err := s.repo.FindExpenseByID(ctx, expenseID)
	if err != nil {
		return nil, errors.New("المصروف غير موجود")
	}

	day, err := s.repo.FindDayByID(ctx, expense.CustodyDayID)
	if err != nil {
		return nil, errors.New("العهدة غير موجودة")
	}
	if branchID != nil && (day.BranchID == nil || *day.BranchID != *branchID) {
		return nil, errors.New("ليس لديك صلاحية الوصول لهذه العهدة - الفرع غير مطابق")
	}

	if err := s.repo.DeleteExpense(ctx, expenseID); err != nil {
		return nil, fmt.Errorf("فشل في حذف المصروف: %w", err)
	}

	// Create log entry
	var adminID *uuid.UUID
	adminName, adminUsername := "النظام", "system"
	if admin != nil {
		adminID = &admin.ID
		adminName = admin.Name
		adminUsername = admin.Username
	}

	logEntry := &domain.CustodyLog{
		BranchID:      day.BranchID,
		CustodyDayID:  day.ID,
		Date:          day.Date,
		ActionType:    "DELETE_EXPENSE",
		Category:      expense.Category,
		Amount:        expense.Amount,
		RecipientName: expense.RecipientName,
		Description:   fmt.Sprintf("حذف مصروف (%s) بقيمة %.2f ريال للمستلم: %s", expense.Category, expense.Amount, expense.RecipientName),
		AdminID:       adminID,
		AdminName:     adminName,
		AdminUsername: adminUsername,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateLog(ctx, logEntry)

	day, err = s.repo.FindDayByID(ctx, expense.CustodyDayID)
	if err != nil {
		return nil, fmt.Errorf("فشل في استرجاع العهدة: %w", err)
	}

	s.recomputeClosing(day)
	_ = s.repo.UpdateDay(ctx, day)

	return s.toResponse(day), nil
}

func (s *custodyService) GetLogs(ctx context.Context, filter dto.CustodyLogFilter) ([]domain.CustodyLog, int64, error) {
	return s.repo.FindLogs(ctx, filter)
}

func (s *custodyService) DeleteLog(ctx context.Context, logID uuid.UUID, branchID *uuid.UUID, admin *domain.Admin) error {
	logEntry, err := s.repo.FindLogByID(ctx, logID)
	if err != nil {
		return errors.New("حركة العهدة غير موجودة")
	}

	if branchID != nil && (logEntry.BranchID == nil || *logEntry.BranchID != *branchID) {
		return errors.New("ليس لديك صلاحية الوصول لعمل هذه الحركة - الفرع غير مطابق")
	}

	day, err := s.repo.FindDayByID(ctx, logEntry.CustodyDayID)
	if err != nil {
		return errors.New("العهدة المرتبطة غير موجودة")
	}

	var adminID *uuid.UUID
	adminName, adminUsername := "النظام", "system"
	if admin != nil {
		adminID = &admin.ID
		adminName = admin.Name
		adminUsername = admin.Username
	}

	if logEntry.ActionType == "ADD_CUSTODY" {
		// Revert the added amount
		if day.AddedAmount >= logEntry.Amount {
			day.AddedAmount -= logEntry.Amount
		} else {
			day.AddedAmount = 0
		}
		s.recomputeClosing(day)
		if err := s.repo.UpdateDay(ctx, day); err != nil {
			return fmt.Errorf("فشل في تحديث مبلغ العهدة: %w", err)
		}

		// Delete the log item
		_ = s.repo.DeleteLog(ctx, logID)

		// Create deletion record log
		delLog := &domain.CustodyLog{
			BranchID:      day.BranchID,
			CustodyDayID:  day.ID,
			Date:          day.Date,
			ActionType:    "DELETE_CUSTODY",
			Category:      "custody",
			Amount:        logEntry.Amount,
			Description:   fmt.Sprintf("حذف/إلغاء مبلغ عهدة مضاف بقيمة %.2f ريال", logEntry.Amount),
			AdminID:       adminID,
			AdminName:     adminName,
			AdminUsername: adminUsername,
			CreatedAt:     time.Now(),
		}
		_ = s.repo.CreateLog(ctx, delLog)
		return nil
	}

	if logEntry.ActionType == "ADD_EXPENSE" {
		// If it's an expense addition log, also attempt to delete the expense if still exists
		_ = s.repo.DeleteLog(ctx, logID)
		s.recomputeClosing(day)
		_ = s.repo.UpdateDay(ctx, day)
		return nil
	}

	return s.repo.DeleteLog(ctx, logID)
}

// SettingService interface & impl
type SettingService interface {
	GetSettings(ctx context.Context) (*dto.AppSettingsResponse, error)
	UpdateSettings(ctx context.Context, req dto.UpdateAppSettingsRequest) error
	GetByKey(ctx context.Context, key string) (string, error)
	SetByKey(ctx context.Context, key, value string) error
}

type settingService struct {
	settingRepo repository.SettingRepository
}

func NewSettingService(settingRepo repository.SettingRepository) SettingService {
	return &settingService{settingRepo: settingRepo}
}

const (
	SettingKeySiteName = "site_name"
	SettingKeyLogoURL  = "logo_url"
)

func (s *settingService) GetSettings(ctx context.Context) (*dto.AppSettingsResponse, error) {
	settings, err := s.settingRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("فشل في جلب الإعدادات: %w", err)
	}
	resp := &dto.AppSettingsResponse{SiteName: "نظام إدارة التوصيل", LogoURL: ""}
	for _, st := range settings {
		switch st.Key {
		case SettingKeySiteName:
			if st.Value != "" {
				resp.SiteName = st.Value
			}
		case SettingKeyLogoURL:
			resp.LogoURL = st.Value
		}
	}
	return resp, nil
}

func (s *settingService) UpdateSettings(ctx context.Context, req dto.UpdateAppSettingsRequest) error {
	siteName := strings.TrimSpace(req.SiteName)
	if siteName == "" {
		siteName = "نظام إدارة التوصيل"
	}
	if err := s.settingRepo.Upsert(ctx, &domain.AppSetting{Key: SettingKeySiteName, Value: siteName}); err != nil {
		return fmt.Errorf("فشل في تحديث اسم الموقع: %w", err)
	}
	if err := s.settingRepo.Upsert(ctx, &domain.AppSetting{Key: SettingKeyLogoURL, Value: req.LogoURL}); err != nil {
		return fmt.Errorf("فشل في تحديث اللوجو: %w", err)
	}
	return nil
}

func (s *settingService) GetByKey(ctx context.Context, key string) (string, error) {
	setting, err := s.settingRepo.GetByKey(ctx, key)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *settingService) SetByKey(ctx context.Context, key, value string) error {
	return s.settingRepo.Upsert(ctx, &domain.AppSetting{Key: key, Value: value})
}

// ------------------------------------------------------------------
// Vehicle Service Implementation (الدبابات والمركبات)
// ------------------------------------------------------------------

type VehicleService interface {
	Create(ctx context.Context, req dto.CreateVehicleRequest) (*domain.Vehicle, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateVehicleRequest) (*domain.Vehicle, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*dto.VehicleResponse, error)
	GetByPlateNumber(ctx context.Context, plateNumber string) (*domain.Vehicle, error)
	GetAll(ctx context.Context, filter dto.VehicleFilter) ([]dto.VehicleResponse, int64, error)
	GetLatestKM(ctx context.Context, plateNumber string) (float64, error)
	RecordOilChange(ctx context.Context, id uuid.UUID) error
}

type vehicleService struct {
	vehicleRepo repository.VehicleRepository
}

func NewVehicleService(vehicleRepo repository.VehicleRepository) VehicleService {
	return &vehicleService{vehicleRepo: vehicleRepo}
}

func (s *vehicleService) Create(ctx context.Context, req dto.CreateVehicleRequest) (*domain.Vehicle, error) {
	plate := strings.TrimSpace(req.PlateNumber)
	if plate == "" {
		return nil, errors.New("رقم اللوحة مطلوب")
	}

	vType := req.VehicleType
	if vType == "" {
		vType = "motorcycle"
	}

	// Check if vehicle exists (including soft-deleted records)
	existingUnscoped, _ := s.vehicleRepo.FindByPlateNumberUnscoped(ctx, plate)
	if existingUnscoped != nil {
		// Case 1: Vehicle was soft-deleted → restore it and update its data
		if existingUnscoped.DeletedAt.Valid {
			existingUnscoped.Brand = req.Brand
			existingUnscoped.ModelYear = req.ModelYear
			existingUnscoped.KeyNumber = req.KeyNumber
			existingUnscoped.VehicleType = vType
			existingUnscoped.Status = domain.VehicleStatusAvailable
			existingUnscoped.BranchID = req.BranchID
			existingUnscoped.Notes = req.Notes
			if req.CurrentKM > 0 {
				existingUnscoped.CurrentKM = req.CurrentKM
			}
			if req.LastOilChangeKM > 0 {
				existingUnscoped.LastOilChangeKM = req.LastOilChangeKM
			}
			if err := s.vehicleRepo.RestoreVehicle(ctx, existingUnscoped); err != nil {
				return nil, fmt.Errorf("فشل استعادة المركبة: %w", err)
			}
			return existingUnscoped, nil
		}

		// Case 2: Vehicle exists but branch_id is nil (auto-created from work sessions) → assign it to this branch
		if existingUnscoped.BranchID == nil && req.BranchID != nil {
			existingUnscoped.Brand = req.Brand
			existingUnscoped.ModelYear = req.ModelYear
			existingUnscoped.KeyNumber = req.KeyNumber
			existingUnscoped.VehicleType = vType
			existingUnscoped.Status = domain.VehicleStatusAvailable
			existingUnscoped.BranchID = req.BranchID
			existingUnscoped.Notes = req.Notes
			if req.CurrentKM > 0 {
				existingUnscoped.CurrentKM = req.CurrentKM
			}
			if req.LastOilChangeKM > 0 {
				existingUnscoped.LastOilChangeKM = req.LastOilChangeKM
			}
			if err := s.vehicleRepo.Update(ctx, existingUnscoped); err != nil {
				return nil, fmt.Errorf("فشل تحديث المركبة: %w", err)
			}
			return existingUnscoped, nil
		}

		// Case 3: Active vehicle with a different branch (or same branch) → real duplicate
		return nil, fmt.Errorf("المركبة برقم اللوحة %s مسجلة بالفعل", plate)
	}

	vehicle := &domain.Vehicle{
		ID:              uuid.New(),
		PlateNumber:     plate,
		VehicleType:     vType,
		Brand:           req.Brand,
		ModelYear:       req.ModelYear,
		KeyNumber:       req.KeyNumber,
		CurrentKM:       req.CurrentKM,
		LastOilChangeKM: req.LastOilChangeKM,
		TotalDistance:   0,
		Status:          domain.VehicleStatusAvailable,
		BranchID:        req.BranchID,
		Notes:           req.Notes,
	}

	if err := s.vehicleRepo.Create(ctx, vehicle); err != nil {
		return nil, err
	}
	return vehicle, nil
}


func (s *vehicleService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateVehicleRequest) (*domain.Vehicle, error) {
	vehicle, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("المركبة غير موجودة")
	}

	if req.PlateNumber != nil {
		vehicle.PlateNumber = strings.TrimSpace(*req.PlateNumber)
	}
	if req.VehicleType != nil {
		vehicle.VehicleType = *req.VehicleType
	}
	if req.Brand != nil {
		vehicle.Brand = *req.Brand
	}
	if req.ModelYear != nil {
		vehicle.ModelYear = *req.ModelYear
	}
	if req.KeyNumber != nil {
		vehicle.KeyNumber = *req.KeyNumber
	}
	if req.CurrentKM != nil {
		vehicle.CurrentKM = *req.CurrentKM
	}
	if req.LastOilChangeKM != nil {
		vehicle.LastOilChangeKM = *req.LastOilChangeKM
	}
	if req.Status != nil {
		vehicle.Status = *req.Status
	}
	if req.BranchID != nil {
		vehicle.BranchID = req.BranchID
	}
	if req.Notes != nil {
		vehicle.Notes = *req.Notes
	}

	if err := s.vehicleRepo.Update(ctx, vehicle); err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *vehicleService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.vehicleRepo.Delete(ctx, id)
}

func (s *vehicleService) GetByID(ctx context.Context, id uuid.UUID) (*dto.VehicleResponse, error) {
	vehicle, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("المركبة غير موجودة")
	}

	interval := oilChangeInterval(vehicle.VehicleType)
	drivenSinceOil := vehicle.CurrentKM - vehicle.LastOilChangeKM
	if drivenSinceOil < 0 {
		drivenSinceOil = 0
	}
	remaining := interval - drivenSinceOil
	if remaining < 0 {
		remaining = 0
	}

	return &dto.VehicleResponse{
		Vehicle:        *vehicle,
		NeedsOilChange: drivenSinceOil >= interval,
		RemainingOilKM: remaining,
	}, nil
}

func (s *vehicleService) GetByPlateNumber(ctx context.Context, plateNumber string) (*domain.Vehicle, error) {
	return s.vehicleRepo.FindByPlateNumber(ctx, plateNumber)
}

func (s *vehicleService) GetAll(ctx context.Context, filter dto.VehicleFilter) ([]dto.VehicleResponse, int64, error) {
	vehicles, total, err := s.vehicleRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.VehicleResponse, len(vehicles))
	for i, v := range vehicles {
		interval := oilChangeInterval(v.VehicleType)
		drivenSinceOil := v.CurrentKM - v.LastOilChangeKM
		if drivenSinceOil < 0 {
			drivenSinceOil = 0
		}
		remaining := interval - drivenSinceOil
		if remaining < 0 {
			remaining = 0
		}

		responses[i] = dto.VehicleResponse{
			Vehicle:        v,
			NeedsOilChange: drivenSinceOil >= interval,
			RemainingOilKM: remaining,
		}
	}

	return responses, total, nil
}

func (s *vehicleService) GetLatestKM(ctx context.Context, plateNumber string) (float64, error) {
	return s.vehicleRepo.FindLatestVehicleKM(ctx, plateNumber)
}

func (s *vehicleService) RecordOilChange(ctx context.Context, id uuid.UUID) error {
	vehicle, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("المركبة غير موجودة")
	}
	return s.vehicleRepo.RecordOilChange(ctx, id, vehicle.CurrentKM)
}

// ------------------------------------------------------------------
// 1. FuelLogService
// ------------------------------------------------------------------
type FuelLogService interface {
	Create(ctx context.Context, req dto.CreateFuelLogRequest, adminBranchID *uuid.UUID) (*domain.FuelLog, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateFuelLogRequest) (*domain.FuelLog, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.FuelLog, error)
	GetAll(ctx context.Context, filter dto.FuelLogFilter, adminBranchID *uuid.UUID) ([]domain.FuelLog, int64, error)
	GetStats(ctx context.Context, branchID *uuid.UUID, startDate, endDate string) (totalCost float64, totalLiters float64, totalLogs int64, err error)
}

type fuelLogService struct {
	repo repository.FuelLogRepository
}

func NewFuelLogService(repo repository.FuelLogRepository) FuelLogService {
	return &fuelLogService{repo: repo}
}

func (s *fuelLogService) Create(ctx context.Context, req dto.CreateFuelLogRequest, adminBranchID *uuid.UUID) (*domain.FuelLog, error) {
	fuelDate := time.Now()
	if req.FuelDate != "" {
		if t, err := time.Parse("2006-01-02", req.FuelDate); err == nil {
			fuelDate = t
		}
	}

	branchID := req.BranchID
	if branchID == nil && adminBranchID != nil {
		branchID = adminBranchID
	}

	log := &domain.FuelLog{
		ID:              uuid.New(),
		EmployeeID:      req.EmployeeID,
		VehiclePlate:    req.VehiclePlate,
		ShiftID:         req.ShiftID,
		Amount:          req.Amount,
		Liters:          req.Liters,
		FuelDate:        fuelDate,
		StationName:     req.StationName,
		InvoiceImageURL: req.InvoiceImageURL,
		BranchID:        branchID,
		Notes:           req.Notes,
	}

	if err := s.repo.Create(ctx, log); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, log.ID)
}

func (s *fuelLogService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateFuelLogRequest) (*domain.FuelLog, error) {
	log, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("سجل الوقود غير موجود")
	}

	if req.EmployeeID != nil {
		log.EmployeeID = req.EmployeeID
	}
	if req.VehiclePlate != nil {
		log.VehiclePlate = *req.VehiclePlate
	}
	if req.Amount != nil {
		log.Amount = *req.Amount
	}
	if req.Liters != nil {
		log.Liters = *req.Liters
	}
	if req.FuelDate != nil && *req.FuelDate != "" {
		if t, err := time.Parse("2006-01-02", *req.FuelDate); err == nil {
			log.FuelDate = t
		}
	}
	if req.StationName != nil {
		log.StationName = *req.StationName
	}
	if req.InvoiceImageURL != nil {
		log.InvoiceImageURL = *req.InvoiceImageURL
	}
	if req.Notes != nil {
		log.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, log); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *fuelLogService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *fuelLogService) GetByID(ctx context.Context, id uuid.UUID) (*domain.FuelLog, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *fuelLogService) GetAll(ctx context.Context, filter dto.FuelLogFilter, adminBranchID *uuid.UUID) ([]domain.FuelLog, int64, error) {
	if adminBranchID != nil {
		filter.BranchID = adminBranchID
	}
	return s.repo.FindAll(ctx, filter)
}

func (s *fuelLogService) GetStats(ctx context.Context, branchID *uuid.UUID, startDate, endDate string) (float64, float64, int64, error) {
	return s.repo.GetFuelStats(ctx, branchID, startDate, endDate)
}

// ------------------------------------------------------------------
// 2. TrafficViolationService
// ------------------------------------------------------------------
type TrafficViolationService interface {
	Create(ctx context.Context, req dto.CreateTrafficViolationRequest, adminBranchID *uuid.UUID) (*domain.TrafficViolation, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateTrafficViolationRequest) (*domain.TrafficViolation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TrafficViolation, error)
	GetAll(ctx context.Context, filter dto.TrafficViolationFilter, adminBranchID *uuid.UUID) ([]domain.TrafficViolation, int64, error)
	GetStats(ctx context.Context, branchID *uuid.UUID) (totalAmount float64, deductedAmount float64, totalCount int64, err error)
}

type trafficViolationService struct {
	repo repository.TrafficViolationRepository
}

func NewTrafficViolationService(repo repository.TrafficViolationRepository) TrafficViolationService {
	return &trafficViolationService{repo: repo}
}

func (s *trafficViolationService) Create(ctx context.Context, req dto.CreateTrafficViolationRequest, adminBranchID *uuid.UUID) (*domain.TrafficViolation, error) {
	vDate := time.Now()
	if req.ViolationDate != "" {
		if t, err := time.Parse("2006-01-02", req.ViolationDate); err == nil {
			vDate = t
		}
	}

	status := "RECORDED"
	if req.Status != "" {
		status = req.Status
	}

	branchID := req.BranchID
	if branchID == nil && adminBranchID != nil {
		branchID = adminBranchID
	}

	v := &domain.TrafficViolation{
		ID:              uuid.New(),
		ViolationNumber: req.ViolationNumber,
		EmployeeID:      req.EmployeeID,
		VehiclePlate:    req.VehiclePlate,
		Amount:          req.Amount,
		Reason:          req.Reason,
		ViolationDate:   vDate,
		City:            req.City,
		Status:          status,
		BranchID:        branchID,
		Notes:           req.Notes,
	}

	if err := s.repo.Create(ctx, v); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, v.ID)
}

func (s *trafficViolationService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateTrafficViolationRequest) (*domain.TrafficViolation, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("المخالفة غير موجودة")
	}

	if req.ViolationNumber != nil {
		v.ViolationNumber = *req.ViolationNumber
	}
	if req.EmployeeID != nil {
		v.EmployeeID = req.EmployeeID
	}
	if req.VehiclePlate != nil {
		v.VehiclePlate = *req.VehiclePlate
	}
	if req.Amount != nil {
		v.Amount = *req.Amount
	}
	if req.Reason != nil {
		v.Reason = *req.Reason
	}
	if req.ViolationDate != nil && *req.ViolationDate != "" {
		if t, err := time.Parse("2006-01-02", *req.ViolationDate); err == nil {
			v.ViolationDate = t
		}
	}
	if req.City != nil {
		v.City = *req.City
	}
	if req.Status != nil {
		v.Status = *req.Status
	}
	if req.Notes != nil {
		v.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, v); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *trafficViolationService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *trafficViolationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.TrafficViolation, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *trafficViolationService) GetAll(ctx context.Context, filter dto.TrafficViolationFilter, adminBranchID *uuid.UUID) ([]domain.TrafficViolation, int64, error) {
	if adminBranchID != nil {
		filter.BranchID = adminBranchID
	}
	return s.repo.FindAll(ctx, filter)
}

func (s *trafficViolationService) GetStats(ctx context.Context, branchID *uuid.UUID) (float64, float64, int64, error) {
	return s.repo.GetViolationStats(ctx, branchID)
}

// ------------------------------------------------------------------
// 3. MaintenanceRequestService
// ------------------------------------------------------------------
type MaintenanceRequestService interface {
	Create(ctx context.Context, req dto.CreateMaintenanceRequestRequest, adminBranchID *uuid.UUID) (*domain.MaintenanceRequest, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateMaintenanceRequestRequest) (*domain.MaintenanceRequest, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.MaintenanceRequest, error)
	GetAll(ctx context.Context, filter dto.MaintenanceRequestFilter, adminBranchID *uuid.UUID) ([]domain.MaintenanceRequest, int64, error)
}

type maintenanceRequestService struct {
	repo repository.MaintenanceRequestRepository
}

func NewMaintenanceRequestService(repo repository.MaintenanceRequestRepository) MaintenanceRequestService {
	return &maintenanceRequestService{repo: repo}
}

func (s *maintenanceRequestService) Create(ctx context.Context, req dto.CreateMaintenanceRequestRequest, adminBranchID *uuid.UUID) (*domain.MaintenanceRequest, error) {
	priority := "MEDIUM"
	if req.Priority != "" {
		priority = req.Priority
	}
	status := "OPEN"
	if req.Status != "" {
		status = req.Status
	}

	branchID := req.BranchID
	if branchID == nil && adminBranchID != nil {
		branchID = adminBranchID
	}

	m := &domain.MaintenanceRequest{
		ID:               uuid.New(),
		VehiclePlate:     req.VehiclePlate,
		EmployeeID:       req.EmployeeID,
		IssueDescription: req.IssueDescription,
		Priority:         priority,
		EstimatedCost:    req.EstimatedCost,
		ActualCost:       req.ActualCost,
		WorkshopName:     req.WorkshopName,
		Status:           status,
		BranchID:         branchID,
		Notes:            req.Notes,
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, m.ID)
}

func (s *maintenanceRequestService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateMaintenanceRequestRequest) (*domain.MaintenanceRequest, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("طلب الصيانة غير موجود")
	}

	if req.VehiclePlate != nil {
		m.VehiclePlate = *req.VehiclePlate
	}
	if req.EmployeeID != nil {
		m.EmployeeID = req.EmployeeID
	}
	if req.IssueDescription != nil {
		m.IssueDescription = *req.IssueDescription
	}
	if req.Priority != nil {
		m.Priority = *req.Priority
	}
	if req.EstimatedCost != nil {
		m.EstimatedCost = *req.EstimatedCost
	}
	if req.ActualCost != nil {
		m.ActualCost = *req.ActualCost
	}
	if req.WorkshopName != nil {
		m.WorkshopName = *req.WorkshopName
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	if req.Notes != nil {
		m.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *maintenanceRequestService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *maintenanceRequestService) GetByID(ctx context.Context, id uuid.UUID) (*domain.MaintenanceRequest, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *maintenanceRequestService) GetAll(ctx context.Context, filter dto.MaintenanceRequestFilter, adminBranchID *uuid.UUID) ([]domain.MaintenanceRequest, int64, error) {
	if adminBranchID != nil {
		filter.BranchID = adminBranchID
	}
	return s.repo.FindAll(ctx, filter)
}

// ------------------------------------------------------------------
// 4. EmployeeDocumentService
// ------------------------------------------------------------------
type EmployeeDocumentService interface {
	Create(ctx context.Context, req dto.CreateEmployeeDocumentRequest) (*domain.EmployeeDocument, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeDocumentRequest) (*domain.EmployeeDocument, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDocument, error)
	GetAll(ctx context.Context, filter dto.EmployeeDocumentFilter) ([]domain.EmployeeDocument, int64, error)
	GetExpiringSoon(ctx context.Context, days int) ([]domain.EmployeeDocument, error)
}

type employeeDocumentService struct {
	repo repository.EmployeeDocumentRepository
}

func NewEmployeeDocumentService(repo repository.EmployeeDocumentRepository) EmployeeDocumentService {
	return &employeeDocumentService{repo: repo}
}

func (s *employeeDocumentService) Create(ctx context.Context, req dto.CreateEmployeeDocumentRequest) (*domain.EmployeeDocument, error) {
	var issueDate *time.Time
	if req.IssueDate != nil && *req.IssueDate != "" {
		if t, err := time.Parse("2006-01-02", *req.IssueDate); err == nil {
			issueDate = &t
		}
	}
	var expiryDate *time.Time
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		if t, err := time.Parse("2006-01-02", *req.ExpiryDate); err == nil {
			expiryDate = &t
		}
	}

	status := "VALID"
	if req.Status != "" {
		status = req.Status
	}

	doc := &domain.EmployeeDocument{
		ID:         uuid.New(),
		EmployeeID: req.EmployeeID,
		DocType:    req.DocType,
		Title:      req.Title,
		DocNumber:  req.DocNumber,
		FileURL:    req.FileURL,
		IssueDate:  issueDate,
		ExpiryDate: expiryDate,
		Status:     status,
		Notes:      req.Notes,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, doc.ID)
}

func (s *employeeDocumentService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeDocumentRequest) (*domain.EmployeeDocument, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("المستند غير موجود")
	}

	if req.DocType != nil {
		doc.DocType = *req.DocType
	}
	if req.Title != nil {
		doc.Title = *req.Title
	}
	if req.DocNumber != nil {
		doc.DocNumber = *req.DocNumber
	}
	if req.FileURL != nil {
		doc.FileURL = *req.FileURL
	}
	if req.IssueDate != nil && *req.IssueDate != "" {
		if t, err := time.Parse("2006-01-02", *req.IssueDate); err == nil {
			doc.IssueDate = &t
		}
	}
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		if t, err := time.Parse("2006-01-02", *req.ExpiryDate); err == nil {
			doc.ExpiryDate = &t
		}
	}
	if req.Status != nil {
		doc.Status = *req.Status
	}
	if req.Notes != nil {
		doc.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, doc); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *employeeDocumentService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *employeeDocumentService) GetByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDocument, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *employeeDocumentService) GetAll(ctx context.Context, filter dto.EmployeeDocumentFilter) ([]domain.EmployeeDocument, int64, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *employeeDocumentService) GetExpiringSoon(ctx context.Context, days int) ([]domain.EmployeeDocument, error) {
	if days <= 0 {
		days = 30
	}
	return s.repo.FindExpiringSoon(ctx, days)
}

// ------------------------------------------------------------------
// 5. EmployeeBankAccountService
// ------------------------------------------------------------------
type EmployeeBankAccountService interface {
	Create(ctx context.Context, req dto.CreateEmployeeBankAccountRequest) (*domain.EmployeeBankAccount, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeBankAccountRequest) (*domain.EmployeeBankAccount, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeBankAccount, error)
	GetAll(ctx context.Context, filter dto.EmployeeBankAccountFilter) ([]domain.EmployeeBankAccount, int64, error)
}

type employeeBankAccountService struct {
	repo repository.EmployeeBankAccountRepository
}

func NewEmployeeBankAccountService(repo repository.EmployeeBankAccountRepository) EmployeeBankAccountService {
	return &employeeBankAccountService{repo: repo}
}

func (s *employeeBankAccountService) Create(ctx context.Context, req dto.CreateEmployeeBankAccountRequest) (*domain.EmployeeBankAccount, error) {
	acc := &domain.EmployeeBankAccount{
		ID:               uuid.New(),
		EmployeeID:       req.EmployeeID,
		BankName:         req.BankName,
		IBAN:             req.IBAN,
		AccountOwnerName: req.AccountOwnerName,
		IsDefault:        req.IsDefault,
	}

	if err := s.repo.Create(ctx, acc); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, acc.ID)
}

func (s *employeeBankAccountService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeBankAccountRequest) (*domain.EmployeeBankAccount, error) {
	acc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("الحساب البنكي غير موجود")
	}

	if req.BankName != nil {
		acc.BankName = *req.BankName
	}
	if req.IBAN != nil {
		acc.IBAN = *req.IBAN
	}
	if req.AccountOwnerName != nil {
		acc.AccountOwnerName = *req.AccountOwnerName
	}
	if req.IsDefault != nil {
		acc.IsDefault = *req.IsDefault
	}

	if err := s.repo.Update(ctx, acc); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *employeeBankAccountService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *employeeBankAccountService) GetByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeBankAccount, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *employeeBankAccountService) GetAll(ctx context.Context, filter dto.EmployeeBankAccountFilter) ([]domain.EmployeeBankAccount, int64, error) {
	return s.repo.FindAll(ctx, filter)
}

// ------------------------------------------------------------------
// 6. LeaveRequestService
// ------------------------------------------------------------------
type LeaveRequestService interface {
	Create(ctx context.Context, req dto.CreateLeaveRequestRequest) (*domain.LeaveRequest, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, req dto.UpdateLeaveRequestStatusRequest) (*domain.LeaveRequest, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error)
	GetAll(ctx context.Context, filter dto.LeaveRequestFilter) ([]domain.LeaveRequest, int64, error)
}

type leaveRequestService struct {
	repo repository.LeaveRequestRepository
}

func NewLeaveRequestService(repo repository.LeaveRequestRepository) LeaveRequestService {
	return &leaveRequestService{repo: repo}
}

func (s *leaveRequestService) Create(ctx context.Context, req dto.CreateLeaveRequestRequest) (*domain.LeaveRequest, error) {
	days := req.DaysCount
	if days <= 0 {
		days = 1
	}

	leave := &domain.LeaveRequest{
		ID:         uuid.New(),
		EmployeeID: req.EmployeeID,
		LeaveType:  req.LeaveType,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		DaysCount:  days,
		Reason:     req.Reason,
		Status:     "PENDING",
	}

	if err := s.repo.Create(ctx, leave); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, leave.ID)
}

func (s *leaveRequestService) UpdateStatus(ctx context.Context, id uuid.UUID, req dto.UpdateLeaveRequestStatusRequest) (*domain.LeaveRequest, error) {
	leave, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("طلب الإجازة غير موجود")
	}

	leave.Status = req.Status
	leave.ApprovedByName = req.ApprovedByName

	if err := s.repo.Update(ctx, leave); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *leaveRequestService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *leaveRequestService) GetByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *leaveRequestService) GetAll(ctx context.Context, filter dto.LeaveRequestFilter) ([]domain.LeaveRequest, int64, error) {
	return s.repo.FindAll(ctx, filter)
}

// ------------------------------------------------------------------
// 7. SupportTicketService
// ------------------------------------------------------------------
type SupportTicketService interface {
	Create(ctx context.Context, req dto.CreateSupportTicketRequest, adminBranchID *uuid.UUID) (*domain.SupportTicket, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateSupportTicketRequest) (*domain.SupportTicket, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SupportTicket, error)
	GetAll(ctx context.Context, filter dto.SupportTicketFilter, adminBranchID *uuid.UUID) ([]domain.SupportTicket, int64, error)
}

type supportTicketService struct {
	repo repository.SupportTicketRepository
}

func NewSupportTicketService(repo repository.SupportTicketRepository) SupportTicketService {
	return &supportTicketService{repo: repo}
}

func (s *supportTicketService) Create(ctx context.Context, req dto.CreateSupportTicketRequest, adminBranchID *uuid.UUID) (*domain.SupportTicket, error) {
	category := "OPERATIONAL"
	if req.Category != "" {
		category = req.Category
	}
	priority := "MEDIUM"
	if req.Priority != "" {
		priority = req.Priority
	}

	branchID := req.BranchID
	if branchID == nil && adminBranchID != nil {
		branchID = adminBranchID
	}

	ticketNum := fmt.Sprintf("TCK-%d", time.Now().Unix()%1000000)

	ticket := &domain.SupportTicket{
		ID:           uuid.New(),
		TicketNumber: ticketNum,
		EmployeeID:   req.EmployeeID,
		Subject:      req.Subject,
		Category:     category,
		Priority:     priority,
		Status:       "OPEN",
		Description:  req.Description,
		BranchID:     branchID,
	}

	if err := s.repo.Create(ctx, ticket); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, ticket.ID)
}

func (s *supportTicketService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateSupportTicketRequest) (*domain.SupportTicket, error) {
	ticket, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("التذكرة غير موجودة")
	}

	if req.Subject != nil {
		ticket.Subject = *req.Subject
	}
	if req.Category != nil {
		ticket.Category = *req.Category
	}
	if req.Priority != nil {
		ticket.Priority = *req.Priority
	}
	if req.Status != nil {
		ticket.Status = *req.Status
	}
	if req.Description != nil {
		ticket.Description = *req.Description
	}
	if req.Resolution != nil {
		ticket.Resolution = *req.Resolution
	}

	if err := s.repo.Update(ctx, ticket); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *supportTicketService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *supportTicketService) GetByID(ctx context.Context, id uuid.UUID) (*domain.SupportTicket, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *supportTicketService) GetAll(ctx context.Context, filter dto.SupportTicketFilter, adminBranchID *uuid.UUID) ([]domain.SupportTicket, int64, error) {
	if adminBranchID != nil {
		filter.BranchID = adminBranchID
	}
	return s.repo.FindAll(ctx, filter)
}



type NotificationService interface {
	GetMyNotifications(ctx context.Context, adminID uuid.UUID) ([]dto.NotificationResponse, error)
	MarkAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, adminID uuid.UUID) error
}

type notificationService struct {
	notifRepo repository.NotificationRepository
	adminRepo repository.AdminRepository
}

func NewNotificationService(notifRepo repository.NotificationRepository, adminRepo repository.AdminRepository) NotificationService {
	return &notificationService{notifRepo: notifRepo, adminRepo: adminRepo}
}

func (s *notificationService) GetMyNotifications(ctx context.Context, adminID uuid.UUID) ([]dto.NotificationResponse, error) {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	notifs, err := s.notifRepo.FindUnreadByAdmin(ctx, admin.ID, admin.BranchID)
	if err != nil {
		return nil, err
	}

	var res []dto.NotificationResponse
	for _, n := range notifs {
		res = append(res, dto.NotificationResponse{
			ID:        n.ID,
			Title:     n.Title,
			Body:      n.Body,
			Type:      n.Type,
			Status:    n.Status,
			CreatedAt: n.CreatedAt,
		})
	}
	return res, nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	return s.notifRepo.MarkAsRead(ctx, id, adminID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, adminID uuid.UUID) error {
	return s.notifRepo.MarkAllAsRead(ctx, adminID)
}

// ------------------------------------------------------------------
// 8. ArchiveService (سجل الأرشيف والمحذوفات)
// ------------------------------------------------------------------
type ArchiveService interface {
	GetArchivedItems(ctx context.Context, filter dto.ArchiveFilter) (dto.ArchiveResponseDTO, error)
	Restore(ctx context.Context, itemType string, id uuid.UUID) error
	PermanentDelete(ctx context.Context, itemType string, id uuid.UUID) error
	BulkRestore(ctx context.Context, itemType string, ids []uuid.UUID) error
	BulkPermanentDelete(ctx context.Context, itemType string, ids []uuid.UUID) error
}

type archiveService struct {
	archiveRepo repository.ArchiveRepository
}

func NewArchiveService(archiveRepo repository.ArchiveRepository) ArchiveService {
	return &archiveService{archiveRepo: archiveRepo}
}

func (s *archiveService) GetArchivedItems(ctx context.Context, filter dto.ArchiveFilter) (dto.ArchiveResponseDTO, error) {
	items, total, stats, err := s.archiveRepo.GetArchivedItems(ctx, filter)
	if err != nil {
		return dto.ArchiveResponseDTO{}, err
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	return dto.ArchiveResponseDTO{
		Data:       items,
		Stats:      stats,
		Total:      total,
		Page:       filter.Page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *archiveService) Restore(ctx context.Context, itemType string, id uuid.UUID) error {
	return s.archiveRepo.Restore(ctx, itemType, id)
}

func (s *archiveService) PermanentDelete(ctx context.Context, itemType string, id uuid.UUID) error {
	return s.archiveRepo.PermanentDelete(ctx, itemType, id)
}

func (s *archiveService) BulkRestore(ctx context.Context, itemType string, ids []uuid.UUID) error {
	return s.archiveRepo.BulkRestore(ctx, itemType, ids)
}

func (s *archiveService) BulkPermanentDelete(ctx context.Context, itemType string, ids []uuid.UUID) error {
	return s.archiveRepo.BulkPermanentDelete(ctx, itemType, ids)
}

