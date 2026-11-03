package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jamesc159/monmetrics/internal/models"
)

var gradePattern = regexp.MustCompile(`(?i)\b(PSA|BGS|Beckett|CGC|SGC|TAG|ACE)\s*#?\s*(10|9\.5|9|8\.5|8|7\.5|7|6\.5|6|5\.5|5|4\.5|4|3\.5|3|2\.5|2|1\.5|1)\b`)
var numberPattern = regexp.MustCompile(`(?i)\b0*([0-9]{1,4})\s*/\s*0*([0-9]{1,4})\b`)
var searchPunctuation = regexp.MustCompile(`[^a-z0-9]+`)
var lotPattern = regexp.MustCompile(`\b(?:lot|playset|bundle|collection)\b|\b(?:x\s*[2-9]|[2-9]\s*x)\b|\b[2-9]\s+copies\b`)

func DetectGrade(title string) string {
	match := gradePattern.FindStringSubmatch(title)
	if len(match) != 3 {
		return "raw"
	}
	company := strings.ToLower(match[1])
	if company == "beckett" {
		company = "bgs"
	}
	return fmt.Sprintf("%s-%s", company, strings.TrimSpace(match[2]))
}

func Relevant(card models.Card, title string) (bool, string) {
	lower := strings.ToLower(title)
	if card.Number != "" && !hasCardNumber(lower, card.Number) {
		return false, "card number not present"
	}
	if card.Name != "" {
		titleWords := make(map[string]bool)
		for _, word := range strings.Fields(searchPunctuation.ReplaceAllString(lower, " ")) {
			titleWords[word] = true
		}
		for _, word := range strings.Fields(searchPunctuation.ReplaceAllString(strings.ToLower(card.Name), " ")) {
			if len(word) > 1 && !titleWords[word] {
				return false, "card name not present"
			}
		}
	}
	for _, word := range card.EbayExclude {
		word = strings.ToLower(strings.TrimSpace(word))
		if word != "" && strings.Contains(lower, word) {
			return false, "matched card exclude keyword: " + word
		}
	}
	for _, word := range []string{"proxy", "orica", "custom", "code card", "digital code"} {
		if strings.Contains(lower, word) {
			return false, "excluded listing: " + strings.TrimSpace(word)
		}
	}
	if lotPattern.MatchString(lower) {
		return false, "excluded listing: multiple cards or lot"
	}
	return true, ""
}

func hasCardNumber(title, number string) bool {
	wanted := numberPattern.FindStringSubmatch(number)
	if len(wanted) != 3 {
		return strings.Contains(title, strings.ToLower(strings.TrimSpace(number)))
	}
	for _, match := range numberPattern.FindAllStringSubmatch(title, -1) {
		if match[1] == wanted[1] && match[2] == wanted[2] {
			return true
		}
	}
	return false
}
