package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"time"
)

var tagRegex = regexp.MustCompile(`\[webhook@(\d{4}-\d{2}-\d{2}\s\d{2}:\d{2})\]`)

func ExtractReminders(filePath, content string) []Reminder {
	var reminders []Reminder
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "- [x]") || strings.HasPrefix(trimmedLine, "- [-]") {
			continue
		}

		matches := tagRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			dateStr := matches[1]

			parsedDate, err := time.ParseInLocation("2006-01-02 15:04", dateStr, time.Local)
			if err != nil {
				continue
			}

			cleanText := tagRegex.ReplaceAllString(line, "")
			cleanText = strings.TrimSpace(cleanText)
			cleanText = strings.TrimPrefix(cleanText, "- [ ] ")
			cleanText = strings.TrimPrefix(cleanText, "- ")

			hash := sha256.Sum256([]byte(filePath + line))
			id := hex.EncodeToString(hash[:])[:16]

			reminders = append(reminders, Reminder{
				ID:      id,
				File:    filePath,
				Content: cleanText,
				DueDate: parsedDate,
			})
		}
	}

	return reminders
}
