package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func registerAppVersionRoute(r *gin.Engine) {
	r.GET("/api/v1/app-version", func(c *gin.Context) {
		minAndroidCode := 20
		if raw := strings.TrimSpace(os.Getenv("MIN_ANDROID_VERSION_CODE")); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				minAndroidCode = parsed
			}
		}

		latestAndroid := strings.TrimSpace(os.Getenv("LATEST_ANDROID_VERSION"))
		if latestAndroid == "" {
			latestAndroid = "2.0.0"
		}

		minIOS := strings.TrimSpace(os.Getenv("MIN_IOS_VERSION"))
		if minIOS == "" {
			minIOS = "2.0.0"
		}

		storeURL := strings.TrimSpace(os.Getenv("PLAY_STORE_URL"))
		if storeURL == "" {
			storeURL = "https://play.google.com/store/apps/details?id=com.aams.manadeeb"
		}

		msgAr := strings.TrimSpace(os.Getenv("FORCE_UPDATE_MESSAGE_AR"))
		if msgAr == "" {
			msgAr = "يتوفر إصدار جديد ومهم للتطبيق على المتجر. يرجى التحديث الآن للاستمرار في العمل."
		}

		msgEn := strings.TrimSpace(os.Getenv("FORCE_UPDATE_MESSAGE_EN"))
		if msgEn == "" {
			msgEn = "An important new version is available on the Store. Please update now to continue."
		}

		c.JSON(http.StatusOK, gin.H{
			"min_android_version_code": minAndroidCode,
			"latest_android_version":   latestAndroid,
			"min_ios_version":          minIOS,
			"play_store_url":           storeURL,
			"message_ar":               msgAr,
			"message_en":               msgEn,
		})
	})
}
