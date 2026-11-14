package work

import "strings"

func normalizeNameToken(token string) string {
	token = strings.TrimSpace(token)
	token = strings.ReplaceAll(token, "أ", "ا")
	token = strings.ReplaceAll(token, "إ", "ا")
	token = strings.ReplaceAll(token, "آ", "ا")
	token = strings.ReplaceAll(token, "ة", "ه")
	token = strings.ReplaceAll(token, "ى", "ي")
	return token
}

func translateSupervisorName(rawName string, lang string) string {
	rawName = strings.TrimSpace(rawName)
	if rawName == "" {
		switch lang {
		case "en":
			return "Supervisor"
		case "bn":
			return "সুপারভাইজার"
		case "ur":
			return "نگران"
		default:
			return "المشرف"
		}
	}

	normRaw := normalizeNameToken(rawName)
	for key, trans := range knownSupervisors {
		if strings.HasPrefix(normRaw, normalizeNameToken(key)) {
			switch lang {
			case "en":
				return trans.en
			case "bn":
				return trans.bn
			case "ur":
				return trans.ur
			default:
				return trans.ar
			}
		}
	}

	parts := strings.Fields(rawName)
	if len(parts) == 0 {
		return rawName
	}

	firstNames := []string{parts[0]}
	partIdx := 1
	if (parts[0] == "عبد" || parts[0] == "ابو" || parts[0] == "أبو") && len(parts) > 1 {
		firstNames = append(firstNames, parts[1])
		partIdx = 2
	}
	if partIdx < len(parts) {
		firstNames = append(firstNames, parts[partIdx])
		if (parts[partIdx] == "عبد" || parts[partIdx] == "ابو" || parts[partIdx] == "أبو") && len(parts) > partIdx+1 {
			firstNames = append(firstNames, parts[partIdx+1])
		}
	}

	twoNames := strings.Join(firstNames, " ")
	normTwo := normalizeNameToken(twoNames)
	for key, trans := range knownSupervisors {
		if strings.HasPrefix(normTwo, normalizeNameToken(key)) {
			switch lang {
			case "en":
				return trans.en
			case "bn":
				return trans.bn
			case "ur":
				return trans.ur
			default:
				return trans.ar
			}
		}
	}

	translatedParts := make([]string, 0, len(firstNames))
	for _, p := range firstNames {
		normP := normalizeNameToken(p)
		if entry, ok := supervisorWordDict[normP]; ok {
			switch lang {
			case "en":
				translatedParts = append(translatedParts, entry.en)
			case "bn":
				translatedParts = append(translatedParts, entry.bn)
			case "ur":
				translatedParts = append(translatedParts, entry.ur)
			default:
				translatedParts = append(translatedParts, entry.ar)
			}
		} else {
			translatedParts = append(translatedParts, p)
		}
	}

	if len(translatedParts) > 0 {
		return strings.Join(translatedParts, " ")
	}

	return twoNames
}
