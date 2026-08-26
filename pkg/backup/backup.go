package backup

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Service performs periodic PostgreSQL backups using pg_dump.
// It writes binary (custom-format) dumps that can be restored with pg_restore.
type Service struct {
	pgDumpPath string
	host       string
	port       string
	user       string
	password   string
	dbname     string
	backupDir  string
	interval   time.Duration
	maxBackups int
	stopCh     chan struct{}
}

// NewService creates a PostgreSQL backup service.
func NewService(pgDumpPath, host, port, user, password, dbname, backupDir string, interval time.Duration, maxBackups int) *Service {
	return &Service{
		pgDumpPath: pgDumpPath,
		host:       host,
		port:       port,
		user:       user,
		password:   password,
		dbname:     dbname,
		backupDir:  backupDir,
		interval:   interval,
		maxBackups: maxBackups,
		stopCh:     make(chan struct{}),
	}
}

// Start begins the backup loop. Call it with `go svc.Start()`.
func (s *Service) Start() {
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		log.Printf("[نسخ احتياطي] فشل إنشاء مجلد النسخ: %v", err)
		return
	}

	log.Printf("[نسخ احتياطي] تم البدء — نسخ PostgreSQL كل %v | الاحتفاظ بآخر %d نسخة", s.interval, s.maxBackups)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.dump()
		case <-s.stopCh:
			log.Println("[نسخ احتياطي] تم إيقاف الخدمة")
			return
		}
	}
}

// Stop signals the backup loop to exit gracefully.
func (s *Service) Stop() {
	close(s.stopCh)
}

// dump creates a single pg_dump custom-format backup.
func (s *Service) dump() {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupName := fmt.Sprintf("delivery_db_%s.dump", timestamp)
	backupPath := filepath.Join(s.backupDir, backupName)

	args := []string{
		"-h", s.host,
		"-p", s.port,
		"-U", s.user,
		"-d", s.dbname,
		"-Fc", // custom (compressed) format — restorable via pg_restore
		"-f", backupPath,
	}

	cmd := exec.Command(s.pgDumpPath, args...)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+s.password)

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[نسخ احتياطي] فشل pg_dump: %v: %s", err, strings.TrimSpace(string(out)))
		return
	}

	log.Printf("[نسخ احتياطي] تم إنشاء نسخة: %s", backupName)
	s.cleanOldBackups()
}

// cleanOldBackups removes the oldest backup files when the count exceeds maxBackups.
func (s *Service) cleanOldBackups() {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return
	}

	var backups []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".dump" {
			backups = append(backups, e)
		}
	}

	if len(backups) <= s.maxBackups {
		return
	}

	// Sort by name (timestamp-based) — oldest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Name() < backups[j].Name()
	})

	toDelete := len(backups) - s.maxBackups
	for i := 0; i < toDelete; i++ {
		path := filepath.Join(s.backupDir, backups[i].Name())
		if err := os.Remove(path); err != nil {
			log.Printf("[نسخ احتياطي] فشل حذف النسخة القديمة %s: %v", backups[i].Name(), err)
		} else {
			log.Printf("[نسخ احتياطي] تم حذف نسخة قديمة: %s", backups[i].Name())
		}
	}
}
