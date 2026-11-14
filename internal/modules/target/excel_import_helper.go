package target

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func getCell(row []string, idx int) string {
	if idx >= 0 && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func parseInt(val string) int {
	if val == "" {
		return 0
	}
	val = strings.ReplaceAll(val, ",", "")
	val = strings.TrimSpace(val)
	n, err := strconv.Atoi(val)
	if err != nil {
		if f, err2 := strconv.ParseFloat(val, 64); err2 == nil {
			return int(f)
		}
		return 0
	}
	return n
}

func isFooterOrEmptyRow(row []string) bool {
	if len(row) == 0 {
		return true
	}
	joined := strings.TrimSpace(strings.Join(row, " "))
	if joined == "" || joined == "`" {
		return true
	}
	lower := strings.ToLower(joined)
	if strings.Contains(lower, "توقيع المشرف") ||
		strings.Contains(lower, "مجموع الطلبات") ||
		strings.Contains(lower, "#ref!") ||
		strings.Contains(lower, "بيان يومي") {
		return true
	}
	return false
}

func extractDateFromFilename(filename string) string {
	base := filepath.Base(filename)

	re1 := regexp.MustCompile(`(\d{1,2})[-_](\d{1,2})[-_](\d{4})`)
	if match := re1.FindStringSubmatch(base); len(match) == 4 {
		d, _ := strconv.Atoi(match[1])
		m, _ := strconv.Atoi(match[2])
		y, _ := strconv.Atoi(match[3])
		return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
	}

	re2 := regexp.MustCompile(`(\d{4})[-_](\d{1,2})[-_](\d{1,2})`)
	if match := re2.FindStringSubmatch(base); len(match) == 4 {
		y, _ := strconv.Atoi(match[1])
		m, _ := strconv.Atoi(match[2])
		d, _ := strconv.Atoi(match[3])
		return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
	}

	re3 := regexp.MustCompile(`^(\d{1,2})[-_](\d{1,2})\.xlsx?$`)
	if match := re3.FindStringSubmatch(base); len(match) == 3 {
		d, _ := strconv.Atoi(match[1])
		m, _ := strconv.Atoi(match[2])
		y := time.Now().Year()
		if y < 2026 {
			y = 2026
		}
		return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
	}

	return ""
}

func extractDateFromRows(rows [][]string) string {
	for i := 0; i < len(rows) && i < 3; i++ {
		for _, cell := range rows[i] {
			cell = strings.TrimSpace(cell)
			if cell == "" {
				continue
			}
			reISO := regexp.MustCompile(`(\d{4})[-/](\d{1,2})[-/](\d{1,2})`)
			if match := reISO.FindStringSubmatch(cell); len(match) == 4 {
				y, _ := strconv.Atoi(match[1])
				m, _ := strconv.Atoi(match[2])
				d, _ := strconv.Atoi(match[3])
				return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
			}
			reDMY := regexp.MustCompile(`(\d{1,2})[-/](\d{1,2})[-/](\d{4})`)
			if match := reDMY.FindStringSubmatch(cell); len(match) == 4 {
				d, _ := strconv.Atoi(match[1])
				m, _ := strconv.Atoi(match[2])
				y, _ := strconv.Atoi(match[3])
				return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
			}
			reMDY2 := regexp.MustCompile(`^(\d{2})[-/](\d{2})[-/](\d{2})$`)
			if match := reMDY2.FindStringSubmatch(cell); len(match) == 4 {
				m, _ := strconv.Atoi(match[1])
				d, _ := strconv.Atoi(match[2])
				y, _ := strconv.Atoi(match[3])
				if y < 100 {
					y += 2000
				}
				if m <= 12 && d <= 31 {
					return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
				}
			}
		}
	}
	return ""
}
