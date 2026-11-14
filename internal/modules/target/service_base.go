package target

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TargetService interface {
	ITargetContract
	GetDashboardSummary(ctx context.Context, month string, branch string) (*TargetDashboardSummaryDTO, error)
	ListIdentifiers(ctx context.Context, search, status, month string, branch string) ([]IdentifierPerformanceDTO, error)
	GetIdentifierDetails(ctx context.Context, id uuid.UUID, month string) (*IdentifierDetailsDTO, error)
	CreateIdentifier(ctx context.Context, name, appName, code string, monthlyTarget, dailyTarget int) (*Identifier, error)
	UpdateIdentifier(ctx context.Context, id uuid.UUID, name, appName, code string, monthlyTarget, dailyTarget int, isActive bool, accountStatus string) error
	UpdateIdentifierByIDOrCode(ctx context.Context, idOrCode, name, appName, code string, monthlyTarget, dailyTarget int, isActive bool, accountStatus string) error
	LinkEmployeeToIdentifier(ctx context.Context, idOrCode string, employeeID *string) error
	DeleteIdentifier(ctx context.Context, id uuid.UUID) error
	DeleteAllIdentifiers(ctx context.Context) error

	ListDrivers(ctx context.Context, search, month string, branch string) ([]DriverPerformanceDTO, error)
	ListAlerts(ctx context.Context, date string, unresolvedOnly bool, branch string) ([]TargetAlertDTO, error)
	ResolveAlert(ctx context.Context, id uuid.UUID) error
	ResolveAllAlerts(ctx context.Context, branch string, date string) error

	GetTargetSettings(ctx context.Context) (*TargetSettingsDTO, error)
	UpdateTargetSettings(ctx context.Context, monthlyTarget, dailyTarget int) error

	ListImportBatches(ctx context.Context, limit int) ([]ImportBatch, error)
	DeleteImportBatch(ctx context.Context, id uuid.UUID) error
	DeleteSheetByDate(ctx context.Context, orderDate string) error
}

type targetService struct {
	targetRepo TargetRepository
}

func NewTargetService(targetRepo TargetRepository) TargetService {
	return &targetService{targetRepo: targetRepo}
}

var masterDriversOrder = []string{
	"لابون شندر",
	"هريدي",
	"محمد سليمان",
	"محمد سيد سالم",
	"صفات",
	"مكرم",
	"محمد اسامة",
	"سعيد",
	"جيهانغير",
	"مد روبل",
	"دلوار",
	"رقيب",
	"شوهاج",
	"ابو الخير",
	"ابو كوثر",
	"راجو",
	"تمال",
	"اكاش",
	"اسو عاشور",
	"اسلام عاطف",
	"ايمون",
	"جهيدال",
	"شوهيل الرحمن",
	"حماد حنيف",
	"شوهيل رانا",
	"منهاز",
	"محمد نيون",
	"علي رفعت",
	"محمود اوسين",
	"عمر فاروق",
	"نور زمان",
	"محبوب سردار",
	"مومن جمان",
	"اسماعيل حسين",
	"رياض",
	"حنيف",
	"عبيدول",
	"شهاب الدين",
	"راشد",
	"بشير خميس",
	"شوهان شيك",
	"بلال محمد",
	"رحمان يونس",
	"ابوسفيان",
	"غلام",
	"مصطفي عبد العال",
	"نور علم",
	"نهيد",
	"ميلون",
	"اشرفول",
	"فردين",
	"طارق الاسلام",
	"هايبر شيك",
	"هايدر شيك",
	"محمد فيصل",
	"محمد وقاص",
	"مد منير",
	"محمد متولي",
	"شهادات",
	"غاجي شوهيل",
	"كريم الشريف",
	"عدنان",
	"مدروبل اوسين",
	"مدرويل اوسين",
	"دلوار اوسين",
	"محمود اسماعيل",
	"إبراهيم خليل",
}

func normalizeDriverName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "أ", "ا")
	s = strings.ReplaceAll(s, "إ", "ا")
	s = strings.ReplaceAll(s, "آ", "ا")
	s = strings.ReplaceAll(s, "ة", "ه")
	s = strings.ReplaceAll(s, "ى", "ي")
	s = strings.ReplaceAll(s, "عبد ", "عبد")
	s = strings.ReplaceAll(s, "ابو ", "ابو")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func getDriverMasterRank(name string) int {
	norm := normalizeDriverName(name)
	for idx, mName := range masterDriversOrder {
		if normalizeDriverName(mName) == norm {
			return idx
		}
	}
	return 999999
}


func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func isAllDigits(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
