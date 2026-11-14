package main

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"delivery-backend/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type operationVideoPurgeService interface {
	PurgeExpiredVideos(ctx context.Context, uploadBaseDir string) (int, error)
}

func runSafeBackgroundTask(taskName string, fn func()) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("[BULKHEAD_CRON][task=%s] Recovered panic: %v\n%s", taskName, rec, string(debug.Stack()))
		}
	}()
	fn()
}

func startIqamaExpirationChecker(db *gorm.DB) {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		runSafeBackgroundTask("iqama_expiration_checker", func() {
			checkIqamaExpirations(db)
		})
		for range ticker.C {
			runSafeBackgroundTask("iqama_expiration_checker", func() {
				checkIqamaExpirations(db)
			})
		}
	}()
}

func checkIqamaExpirations(db *gorm.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var emps []domain.Employee
	threshold := time.Now().AddDate(0, 0, 60)

	if err := db.WithContext(ctx).Where("iqama_expiration_date IS NOT NULL AND iqama_expiration_date <= ?", threshold).Find(&emps).Error; err != nil {
		log.Printf("[Cron] Error checking iqama expirations: %v", err)
		return
	}

	todayStr := time.Now().Format("2006-01-02")
	count := 0

	for _, emp := range emps {
		var existing domain.Notification
		err := db.WithContext(ctx).Where("employee_id = ? AND type = ? AND DATE(created_at) = ?", emp.ID, "iqama_expiry", todayStr).First(&existing).Error
		if err != nil {
			expTime := time.Now()
			if emp.IqamaExpirationDate != nil {
				if parsed, parseErr := time.Parse("2006-01-02", *emp.IqamaExpirationDate); parseErr == nil {
					expTime = parsed
				}
			}
			daysLeft := int(time.Until(expTime).Hours() / 24)

			title := "تنبيه اقتراب انتهاء إقامة"
			body := "إقامة الموظف " + emp.Name + " تنتهي بعد " + fmt.Sprintf("%d", daysLeft) + " يوم."
			if daysLeft <= 0 {
				title = "تنبيه انتهاء إقامة"
				body = "إقامة الموظف " + emp.Name + " منتهية!"
			}

			notif := domain.Notification{
				BranchID:   emp.BranchID,
				EmployeeID: &emp.ID,
				Title:      title,
				Body:       body,
				Type:       "iqama_expiry",
				Status:     "unread",
			}
			db.WithContext(ctx).Create(&notif)
			count++
		}
	}

	if count > 0 {
		log.Printf("[Cron] Generated %d iqama expiration notifications", count)
	}
}

func seedDemoReviewer(db *gorm.DB) {
	runSafeBackgroundTask("seed_demo_reviewer", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		demoID := uuid.MustParse("00000000-0000-0000-0000-000001234596")
		var emp domain.Employee
		if err := db.WithContext(ctx).Where("national_id = ?", "1234596").First(&emp).Error; err != nil {
			hash, _ := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
			demoEmployee := domain.Employee{
				ID:               demoID,
				Name:             "تست جوجل",
				NationalID:       "1234596",
				PasswordHash:     string(hash),
				EmployeeNumber:   "EMP-1234596",
				JobRole:          "DRIVER",
				Phone:            "0500000000",
				MotorcycleNumber: "7777",
				KeyNumber:        "KEY-01",
				VehicleType:      "motorcycle",
				Shift:            "morning",
			}
			if createErr := db.WithContext(ctx).Create(&demoEmployee).Error; createErr == nil {
				log.Println("[Seed] Created Google Play Demo Reviewer account (NationalID: 1234596 / Pass: 1234)")
			}
		} else {
			_ = db.WithContext(ctx).Model(&emp).Update("name", "تست جوجل").Error
		}
	})
}

func startOperationVideoCleanupWorker(svc operationVideoPurgeService, uploadDir string) {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		runPurge := func() {
			runSafeBackgroundTask("operation_video_purge", func() {
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				if purged, err := svc.PurgeExpiredVideos(ctx, uploadDir); err != nil {
					log.Printf("[Cron-OperationVideo] Purge error: %v", err)
				} else if purged > 0 {
					log.Printf("[Cron-OperationVideo] Removed %d expired videos", purged)
				}
			})
		}

		runPurge()
		for range ticker.C {
			runPurge()
		}
	}()
}
