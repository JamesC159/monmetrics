package handlers

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/jamesc159/monmetrics/internal/models"
)

// normalizeEmail converts email to lowercase and trims whitespace
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// sanitizeName removes dangerous characters and normalizes whitespace
func sanitizeName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Remove control characters and normalize spaces
	var result strings.Builder
	prevSpace := false
	for _, r := range name {
		if unicode.IsControl(r) {
			continue
		}
		if unicode.IsSpace(r) {
			if !prevSpace {
				result.WriteRune(' ')
				prevSpace = true
			}
			continue
		}
		result.WriteRune(r)
		prevSpace = false
	}

	return result.String()
}

// validateEmail validates email format using regex
func validateEmail(email string) bool {
	// RFC 5322 simplified email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	return emailRegex.MatchString(email) && len(email) <= 254
}

// validatePassword checks password strength according to OWASP guidelines
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}

	// Check for at least one uppercase, lowercase, digit, and special character
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}

	return nil
}

// validateName checks if name is valid
func validateName(name string) error {
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}

	if len(name) > 50 {
		return fmt.Errorf("name must not exceed 50 characters")
	}

	// Check for at least one letter
	hasLetter := false
	for _, r := range name {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}

	if !hasLetter {
		return fmt.Errorf("name must contain at least one letter")
	}

	return nil
}

// validateRegisterRequest validates and normalizes registration request data
func (h *Handlers) validateRegisterRequest(req *models.RegisterRequest) error {
	// Normalize inputs
	req.Email = normalizeEmail(req.Email)
	req.FirstName = sanitizeName(req.FirstName)
	req.LastName = sanitizeName(req.LastName)

	// Validate all fields are present
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return fmt.Errorf("all fields are required")
	}

	// Validate email format
	if !validateEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate password strength
	if err := validatePassword(req.Password); err != nil {
		return err
	}

	// Validate names
	if err := validateName(req.FirstName); err != nil {
		return fmt.Errorf("invalid first name: %v", err)
	}

	if err := validateName(req.LastName); err != nil {
		return fmt.Errorf("invalid last name: %v", err)
	}

	return nil
}

// hashPassword hashes a password using bcrypt with cost factor 12
func (h *Handlers) hashPassword(password string) (string, error) {
	// Use bcrypt with cost factor 12 (OWASP recommended minimum is 10)
	// Cost 12 provides good balance between security and performance
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against a bcrypt hash
func (h *Handlers) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

var (
	hexColorRegex   = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	certNumberRegex = regexp.MustCompile(`^[A-Za-z0-9-]{1,20}$`)
	purchaseDateFmt = "2006-01-02"
)

var (
	validTimeRanges   = map[string]bool{"1d": true, "7d": true, "30d": true, "90d": true, "1y": true, "5y": true}
	validChartSources = map[string]bool{"all": true, "ebay": true, "tcgplayer": true}
	validItemTypes    = map[string]bool{models.ItemTypeRawCard: true, models.ItemTypeGradedCard: true, models.ItemTypeSealed: true}
	validConditions   = map[string]bool{"NM": true, "LP": true, "MP": true, "HP": true, "DMG": true}
	validFinishes     = map[string]bool{"": true, "normal": true, "holo": true, "reverse_holo": true, "foil": true, "1st_edition": true}
	validGraders      = map[string]bool{"PSA": true, "BGS": true, "CGC": true, "SGC": true}
)

const (
	maxPortfolioQuantity = 10000
	maxMoneyValue        = 10_000_000
)

// cleanText strips control characters (optionally keeping newlines) and trims.
func cleanText(s string, allowNewlines bool) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\n' && allowNewlines {
			b.WriteRune(r)
			continue
		}
		if unicode.IsControl(r) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func validMoney(v float64) bool {
	return v >= 0 && v <= maxMoneyValue && !math.IsNaN(v) && !math.IsInf(v, 0)
}

// validateSavedChart normalizes and validates a saved chart request in place.
func validateSavedChart(req *models.SavedChartRequest, maxIndicators int) error {
	req.Name = sanitizeName(req.Name)
	req.Description = cleanText(req.Description, true)

	if req.Name == "" || len([]rune(req.Name)) > 100 {
		return fmt.Errorf("name must be 1-100 characters")
	}
	if len([]rune(req.Description)) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}
	if !validTimeRanges[req.TimeRange] {
		return fmt.Errorf("invalid time range")
	}
	if req.Source == "" {
		req.Source = "all"
	}
	if !validChartSources[req.Source] {
		return fmt.Errorf("invalid source")
	}
	if req.Indicators == nil {
		req.Indicators = []models.ChartIndicator{}
	}
	if len(req.Indicators) > maxIndicators {
		return fmt.Errorf("exceeded maximum indicators limit (%d)", maxIndicators)
	}
	for i := range req.Indicators {
		ind := &req.Indicators[i]
		params, err := validateIndicatorParams(ind.Type, ind.Parameters)
		if err != nil {
			return err
		}
		ind.Parameters = params
		if ind.Color != "" && !hexColorRegex.MatchString(ind.Color) {
			return fmt.Errorf("indicator color must be a hex value like #22c55e")
		}
	}
	if req.ChartType == "" {
		req.ChartType = "line"
	}
	if req.ChartType != "line" && req.ChartType != "candle" {
		return fmt.Errorf("invalid chart type")
	}
	if req.Drawings == nil {
		req.Drawings = []models.ChartDrawing{}
	}
	return validateDrawings(req.Drawings)
}

// intParam reads a whole-number JSON parameter within [min, max].
func intParam(params map[string]interface{}, key string, min, max int) (int, error) {
	v, ok := params[key].(float64)
	if !ok || v != math.Trunc(v) || v < float64(min) || v > float64(max) {
		return 0, fmt.Errorf("indicator %s must be a whole number between %d and %d", key, min, max)
	}
	return int(v), nil
}

// validateIndicatorParams returns a clean parameter map containing only known keys.
func validateIndicatorParams(kind string, params map[string]interface{}) (map[string]interface{}, error) {
	switch kind {
	case "ema", "sma":
		period, err := intParam(params, "period", 2, 200)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"period": period}, nil
	case "bollinger":
		period, err := intParam(params, "period", 2, 200)
		if err != nil {
			return nil, err
		}
		sd, ok := params["stddev"].(float64)
		if !ok || math.IsNaN(sd) || sd < 0.5 || sd > 4 {
			return nil, fmt.Errorf("indicator stddev must be between 0.5 and 4")
		}
		return map[string]interface{}{"period": period, "stddev": sd}, nil
	case "rsi":
		period, err := intParam(params, "period", 2, 100)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"period": period}, nil
	case "macd":
		fast, err := intParam(params, "fast", 2, 199)
		if err != nil {
			return nil, err
		}
		slow, err := intParam(params, "slow", 3, 200)
		if err != nil {
			return nil, err
		}
		signal, err := intParam(params, "signal", 2, 50)
		if err != nil {
			return nil, err
		}
		if fast >= slow {
			return nil, fmt.Errorf("MACD fast period must be less than slow period")
		}
		return map[string]interface{}{"fast": fast, "slow": slow, "signal": signal}, nil
	}
	return nil, fmt.Errorf("unsupported indicator type %q", kind)
}

const (
	maxDrawingsPerChart = 100
	maxDrawingPoints    = 5000
	maxFreehandPoints   = 500
	maxDrawingTime      = 4102444800 // 2100-01-01
)

// drawingPointCount maps drawing types to their required anchor count (0 = variable).
var drawingPointCount = map[string]int{"trendline": 2, "hline": 1, "rect": 2, "text": 1, "freehand": 0, "fib": 2}

var drawingIDRegex = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)

// validateDrawings normalizes and validates chart drawings in place.
func validateDrawings(drawings []models.ChartDrawing) error {
	if len(drawings) > maxDrawingsPerChart {
		return fmt.Errorf("a chart can have at most %d drawings", maxDrawingsPerChart)
	}
	total := 0
	seen := map[string]bool{}
	for i := range drawings {
		d := &drawings[i]
		want, ok := drawingPointCount[d.Type]
		if !ok {
			return fmt.Errorf("unsupported drawing type %q", d.Type)
		}
		if !drawingIDRegex.MatchString(d.ID) || seen[d.ID] {
			return fmt.Errorf("drawing ids must be unique and alphanumeric")
		}
		seen[d.ID] = true
		n := len(d.Points)
		if want > 0 && n != want {
			return fmt.Errorf("%s drawings need exactly %d points", d.Type, want)
		}
		if want == 0 && (n < 2 || n > maxFreehandPoints) {
			return fmt.Errorf("freehand drawings need 2-%d points", maxFreehandPoints)
		}
		total += n
		for _, p := range d.Points {
			if p.Time <= 0 || p.Time > maxDrawingTime || !validMoney(p.Price) {
				return fmt.Errorf("drawing point out of range")
			}
		}
		if !hexColorRegex.MatchString(d.Color) {
			return fmt.Errorf("drawing color must be a hex value like #22c55e")
		}
		if d.LineWidth == 0 {
			d.LineWidth = 2
		}
		if d.LineWidth < 1 || d.LineWidth > 6 {
			return fmt.Errorf("drawing line width must be 1-6")
		}
		d.Text = cleanText(d.Text, false)
		if d.Type == "text" && (d.Text == "" || len([]rune(d.Text)) > 200) {
			return fmt.Errorf("text notes must be 1-200 characters")
		}
		if d.Type != "text" && len([]rune(d.Text)) > 200 {
			return fmt.Errorf("drawing label must not exceed 200 characters")
		}
	}
	if total > maxDrawingPoints {
		return fmt.Errorf("drawings exceed %d total points", maxDrawingPoints)
	}
	return nil
}

var validAlertDirections = map[string]bool{"up": true, "down": true, "either": true}

// validateAlert normalizes and validates a price alert request in place, clearing fields
// that don't apply to the chosen condition.
func validateAlert(req *models.PriceAlertRequest) error {
	if req.Source == "" {
		req.Source = "all"
	}
	if !validChartSources[req.Source] {
		return fmt.Errorf("invalid source")
	}
	req.Note = cleanText(req.Note, false)
	if len([]rune(req.Note)) > 200 {
		return fmt.Errorf("note must not exceed 200 characters")
	}

	target, pct, days, period, dir := req.TargetPrice, req.Pct, req.Days, req.EMAPeriod, req.Direction
	req.TargetPrice, req.Pct, req.Days, req.EMAPeriod, req.Direction = 0, 0, 0, 0, ""
	switch req.Condition {
	case models.AlertAbove, models.AlertBelow:
		if !validMoney(target) || target <= 0 {
			return fmt.Errorf("target price must be greater than 0")
		}
		req.TargetPrice = math.Round(target*100) / 100
	case models.AlertPctChange:
		if math.IsNaN(pct) || pct < 0.5 || pct > 1000 {
			return fmt.Errorf("percent change must be between 0.5 and 1000")
		}
		if days < 1 || days > 365 {
			return fmt.Errorf("days must be between 1 and 365")
		}
		req.Pct, req.Days = pct, days
	case models.AlertEMACross:
		if period < 2 || period > 200 {
			return fmt.Errorf("EMA period must be between 2 and 200")
		}
		req.EMAPeriod = period
	default:
		return fmt.Errorf("unsupported alert condition %q", req.Condition)
	}
	if req.Condition == models.AlertPctChange || req.Condition == models.AlertEMACross {
		if dir == "" {
			dir = "either"
		}
		if !validAlertDirections[dir] {
			return fmt.Errorf("invalid direction")
		}
		req.Direction = dir
	}

	switch req.Mode {
	case "", models.AlertModeOnce:
		req.Mode = models.AlertModeOnce
		req.CooldownHours = 0
	case models.AlertModeRecurring:
		if req.CooldownHours == 0 {
			req.CooldownHours = 24
		}
		if req.CooldownHours < 1 || req.CooldownHours > 168 {
			return fmt.Errorf("cooldown must be between 1 and 168 hours")
		}
	default:
		return fmt.Errorf("invalid mode")
	}
	return nil
}

// validPortfolioGrade enforces per-company grade steps.
func validPortfolioGrade(company string, grade float64) bool {
	if grade < 1 || grade > 10 {
		return false
	}
	if company == "PSA" {
		return grade == math.Trunc(grade)
	}
	return grade*2 == math.Trunc(grade*2)
}

// validatePortfolioItem normalizes and validates a portfolio item request in place.
// Card existence is checked by the caller.
func validatePortfolioItem(req *models.PortfolioItemRequest) (*time.Time, error) {
	req.CardID = strings.TrimSpace(req.CardID)
	req.CustomName = cleanText(req.CustomName, false)
	req.CustomGame = cleanText(req.CustomGame, false)
	req.CustomSet = cleanText(req.CustomSet, false)
	req.CustomImageURL = strings.TrimSpace(req.CustomImageURL)
	req.Language = cleanText(req.Language, false)
	req.PurchaseSource = cleanText(req.PurchaseSource, false)
	req.Notes = cleanText(req.Notes, true)

	if !validItemTypes[req.ItemType] {
		return nil, fmt.Errorf("item_type must be raw_card, graded_card, or sealed")
	}
	if req.CardID == "" {
		if req.CustomName == "" {
			return nil, fmt.Errorf("custom items require a name")
		}
	} else if _, err := primitive.ObjectIDFromHex(req.CardID); err != nil {
		return nil, fmt.Errorf("invalid card_id")
	}
	if len([]rune(req.CustomName)) > 120 || len([]rune(req.CustomGame)) > 60 || len([]rune(req.CustomSet)) > 120 {
		return nil, fmt.Errorf("custom name/game/set too long")
	}
	if req.CustomImageURL != "" {
		u, err := url.Parse(req.CustomImageURL)
		if err != nil || u.Scheme != "https" || u.Host == "" || len(req.CustomImageURL) > 500 {
			return nil, fmt.Errorf("custom image URL must be a valid https URL")
		}
	}
	if req.Quantity < 1 || req.Quantity > maxPortfolioQuantity {
		return nil, fmt.Errorf("quantity must be between 1 and %d", maxPortfolioQuantity)
	}
	if !validFinishes[req.Finish] {
		return nil, fmt.Errorf("invalid finish")
	}
	if len([]rune(req.Language)) > 30 || len([]rune(req.PurchaseSource)) > 60 {
		return nil, fmt.Errorf("language or purchase source too long")
	}
	if len([]rune(req.Notes)) > 1000 {
		return nil, fmt.Errorf("notes must not exceed 1000 characters")
	}
	if !validMoney(req.PurchasePrice) {
		return nil, fmt.Errorf("purchase price must be between 0 and %d", maxMoneyValue)
	}
	if req.ManualValue != nil && !validMoney(*req.ManualValue) {
		return nil, fmt.Errorf("manual value must be between 0 and %d", maxMoneyValue)
	}

	switch req.ItemType {
	case models.ItemTypeGradedCard:
		if req.Grading == nil {
			return nil, fmt.Errorf("graded cards require grading details")
		}
		req.Grading.Company = strings.ToUpper(strings.TrimSpace(req.Grading.Company))
		req.Grading.CertNumber = strings.TrimSpace(req.Grading.CertNumber)
		if !validGraders[req.Grading.Company] {
			return nil, fmt.Errorf("grading company must be PSA, BGS, CGC, or SGC")
		}
		if !validPortfolioGrade(req.Grading.Company, req.Grading.Grade) {
			return nil, fmt.Errorf("invalid grade for %s", req.Grading.Company)
		}
		if req.Grading.CertNumber != "" && !certNumberRegex.MatchString(req.Grading.CertNumber) {
			return nil, fmt.Errorf("cert number must be alphanumeric (max 20)")
		}
		req.Condition = ""
	case models.ItemTypeRawCard:
		if req.Grading != nil {
			return nil, fmt.Errorf("grading is only allowed for graded cards")
		}
		if !validConditions[req.Condition] {
			return nil, fmt.Errorf("condition must be NM, LP, MP, HP, or DMG")
		}
	case models.ItemTypeSealed:
		if req.Grading != nil {
			return nil, fmt.Errorf("grading is only allowed for graded cards")
		}
		req.Condition = ""
		req.Finish = ""
	}

	var purchaseDate *time.Time
	if req.PurchaseDate != "" {
		d, err := time.Parse(purchaseDateFmt, req.PurchaseDate)
		if err != nil {
			return nil, fmt.Errorf("purchase date must be YYYY-MM-DD")
		}
		if d.After(time.Now().UTC()) {
			return nil, fmt.Errorf("purchase date cannot be in the future")
		}
		purchaseDate = &d
	}
	return purchaseDate, nil
}
