package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"delivery-backend/internal/domain"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type SamuraiRecord struct {
	SamuraiID string
	ImageURL  string
	Name      string
	IDNumber  string
}

func main() {
	htmlPath := `D:\aams\f.html`
	dbPath := `D:\aams\backend\delivery_local.db`

	htmlBytes, err := os.ReadFile(htmlPath)
	if err != nil {
		log.Fatalf("Failed to read f.html: %v", err)
	}
	html := string(htmlBytes)

	records := parseSamuraiHTML(html)
	fmt.Printf("[IMPORT] Parsed %d captain records from f.html\n", len(records))

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open sqlite DB: %v", err)
	}

	matched := 0
	updatedImages := 0
	updatedEmpNo := 0
	missed := []string{}

	for _, r := range records {
		var emp domain.Employee
		tx := db.Where("national_id = ?", r.IDNumber).First(&emp)
		if tx.Error != nil {
			missed = append(missed, fmt.Sprintf("ID=%s Name=%s SamuraiID=%s", r.IDNumber, r.Name, r.SamuraiID))
			continue
		}
		matched++

		changes := map[string]interface{}{}

		if emp.EmployeeNumber == "" && r.SamuraiID != "" {
			changes["employee_number"] = r.SamuraiID
			updatedEmpNo++
		}
		if emp.PersonalImage == "" && r.ImageURL != "" {
			changes["personal_image"] = r.ImageURL
			updatedImages++
		}
		if emp.NationalIDImage == "" && r.ImageURL != "" {
			changes["national_id_image"] = r.ImageURL
			updatedImages++
		}

		if len(changes) > 0 {
			db.Model(&emp).Updates(changes)
		}
	}

	fmt.Printf("\n============ SAMURAI IMPORT SUMMARY ============\n")
	fmt.Printf("Parsed records     : %d\n", len(records))
	fmt.Printf("Matched by Iqama   : %d\n", matched)
	fmt.Printf("Employee numbers   : %d updated\n", updatedEmpNo)
	fmt.Printf("Images set (pers+ID): %d total updates\n", updatedImages)
	fmt.Printf("Not found in DB    : %d\n", len(missed))
	for i, m := range missed {
		if i > 30 {
			fmt.Printf("  ... (truncated, %d more)\n", len(missed)-30)
			break
		}
		fmt.Printf("  MISS: %s\n", m)
	}
	fmt.Println("=================================================")
}

func parseSamuraiHTML(html string) []SamuraiRecord {
	records := []SamuraiRecord{}

	rowRe := regexp.MustCompile(`<tr class="MuiTableRow-root RaDatagrid-row[^"]*"[^>]*>.*?</tr>`)
	idRe := regexp.MustCompile(`column-id RaDatagrid-rowCell[^<]*<span class="MuiTypography-root MuiTypography-body2[^"]*"[^>]*>([\s\S]*?)</span>`)
	imgRe := regexp.MustCompile(`column-profilePictureId RaDatagrid-rowCell.*?<img src="([^"]+)"`)
	idNumRe := regexp.MustCompile(`column-idNumber RaDatagrid-rowCell.*?<span class="MuiTypography-root MuiTypography-body2[^"]*"[^>]*>([\s\S]*?)</span>`)
	nameRe := regexp.MustCompile(`column-name RaDatagrid-rowCell.*?<span class="MuiTypography-root MuiTypography-body2[^"]*"[^>]*>([\s\S]*?)</span>`)

	rows := rowRe.FindAllString(html, -1)
	fmt.Printf("[PARSE] Detected %d <tr> rows (pre-filter)\n", len(rows))

	for _, row := range rows {
		idMatch := idRe.FindStringSubmatch(row)
		imgMatch := imgRe.FindStringSubmatch(row)
		nameMatch := nameRe.FindStringSubmatch(row)
		idNumMatch := idNumRe.FindStringSubmatch(row)

		if idMatch == nil || idNumMatch == nil {
			continue
		}

		rec := SamuraiRecord{
			SamuraiID: strings.TrimSpace(idMatch[1]),
			Name:      strings.TrimSpace(safeGroup(nameMatch)),
			IDNumber:  strings.TrimSpace(idNumMatch[1]),
			ImageURL:  strings.TrimSpace(safeGroup(imgMatch)),
		}

		if rec.IDNumber != "" && rec.SamuraiID != "" {
			records = append(records, rec)
		}
	}
	return records
}

func safeGroup(m []string) string {
	if m == nil || len(m) < 2 {
		return ""
	}
	return m[1]
}
