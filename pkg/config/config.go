package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	SQLitePath       string
	PGDumpPath       string
	JWTSecret        string
	JWTRefreshSecret string
	S3Bucket         string
	S3Region         string
	S3Endpoint       string
	StorageMode      string // "local" or "s3"
	AllowedOrigins   string // comma-separated CORS origins
}

func LoadConfig() *Config {
	// Load .env file with explicit path (look both in current dir and one dir below)
	cwd, _ := os.Getwd()
	cwdCandidates := []string{".env", filepath.Join(cwd, ".env"), filepath.Join(cwd, "backend", ".env"), filepath.Join(cwd, "..", ".env")}
	for _, p := range cwdCandidates {
		abs, _ := filepath.Abs(p)
		if st, err := os.Stat(abs); err == nil && !st.IsDir() {
			_ = godotenv.Load(abs)
			break
		}
	}

	port := getEnv("PORT", "8080")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "delivery_db")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")
	jwtSecret := getEnv("JWT_SECRET", "")
	jwtRefreshSecret := getEnv("JWT_REFRESH_SECRET", "")
	s3Bucket := getEnv("S3_BUCKET", "delivery-assets")
	s3Region := getEnv("S3_REGION", "us-east-1")
	s3Endpoint := getEnv("S3_ENDPOINT", "")
	storageMode := getEnv("STORAGE_MODE", "local")
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,https://aams.kerd2sy.com,https://aams-logistics.kerd2sy.com,https://aams-logistic.com,https://www.aams-logistic.com")
	sqlitePath := getEnv("SQLITE_PATH", `D:\aams\backend\delivery_local.db`)
	pgDumpPath := getEnv("PG_DUMP_PATH", "pg_dump")

	// Validate required security settings
	if jwtSecret == "" {
		panic("FATAL: JWT_SECRET environment variable is required and must not be empty")
	}
	if jwtRefreshSecret == "" {
		panic("FATAL: JWT_REFRESH_SECRET environment variable is required and must not be empty")
	}

	return &Config{
		Port:             port,
		DBHost:           dbHost,
		DBPort:           dbPort,
		DBUser:           dbUser,
		DBPassword:       dbPassword,
		DBName:           dbName,
		DBSSLMode:        dbSSLMode,
		JWTSecret:        jwtSecret,
		JWTRefreshSecret: jwtRefreshSecret,
		S3Bucket:         s3Bucket,
		S3Region:         s3Region,
		S3Endpoint:       s3Endpoint,
		StorageMode:      storageMode,
		AllowedOrigins:   allowedOrigins,
		SQLitePath:       sqlitePath,
		PGDumpPath:       pgDumpPath,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
