package scraper

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/jamesc159/monmetrics/internal/models"
)

var (
	itemIDPattern = regexp.MustCompile(`/itm/(?:[^/?]+/)?([0-9]{8,16})`)
	moneyPattern  = regexp.MustCompile(`(?:US\s*)?\$\s*([0-9][0-9,]*(?:\.[0-9]{1,2})?)`)
	moneyRange    = regexp.MustCompile(`(?i)\$\s*[0-9][0-9,]*(?:\.[0-9]{1,2})?\s+(?:to|-)\s*\$?\s*[0-9]`)
	soldDateRegex = regexp.MustCompile(`(?i)\b(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\s+\d{1,2},?\s+\d{4}\b`)
	datePatterns  = []string{"Jan 2, 2006", "January 2, 2006", "Jan 2 2006", "January 2 2006"}
)

type parsedSale struct {
	ItemID    string
	Title     string
	Price     float64
	Shipping  float64
	Currency  string
	SoldAt    time.Time
	BestOffer bool
	URL       string
}

func ParseSoldHTML(body string, now time.Time) ([]parsedSale, error) {
	root, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	text := strings.ToLower(nodeText(root))
	if strings.Contains(text, "captcha") || strings.Contains(text, "verify you're a human") || strings.Contains(text, "robot check") {
		return nil, ErrBlocked
	}

	var cards []*html.Node
	walk(root, func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "li" && (hasClass(node, "s-item") || hasClass(node, "s-card")) {
			cards = append(cards, node)
		}
	})

	results := make([]parsedSale, 0, len(cards))
	seen := make(map[string]struct{}, len(cards))
	for _, card := range cards {
		sale, ok := parseCard(card, now)
		if !ok {
			continue
		}
		if _, exists := seen[sale.ItemID]; exists {
			continue
		}
		seen[sale.ItemID] = struct{}{}
		results = append(results, sale)
	}
	return results, nil
}

func ParseSoldListingsHTML(body string, now time.Time) ([]models.SoldSale, error) {
	parsed, err := ParseSoldHTML(body, now)
	if err != nil {
		return nil, err
	}
	sales := make([]models.SoldSale, 0, len(parsed))
	for _, sale := range parsed {
		sales = append(sales, models.SoldSale{
			ItemID: sale.ItemID, Title: sale.Title, Price: sale.Price, Shipping: sale.Shipping,
			Currency: sale.Currency, SoldAt: sale.SoldAt, BestOffer: sale.BestOffer, URL: sale.URL,
		})
	}
	return sales, nil
}

func parseCard(card *html.Node, now time.Time) (parsedSale, bool) {
	var links []*html.Node
	walk(card, func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			if href := attr(node, "href"); strings.Contains(href, "/itm/") {
				links = append(links, node)
			}
		}
	})
	if len(links) == 0 {
		return parsedSale{}, false
	}

	link := links[0]
	itemURL, err := url.Parse(attr(link, "href"))
	if err != nil {
		return parsedSale{}, false
	}
	match := itemIDPattern.FindStringSubmatch(itemURL.Path)
	if len(match) != 2 {
		return parsedSale{}, false
	}
	title := ""
	walk(card, func(node *html.Node) {
		if title == "" && node.Type == html.ElementNode && (node.Data == "h3" || hasClass(node, "s-item__title") || hasClass(node, "s-card__title")) {
			title = strings.TrimSpace(nodeText(node))
		}
	})
	if title == "" || strings.EqualFold(title, "Shop on eBay") {
		return parsedSale{}, false
	}

	cardText := nodeText(card)
	priceText := firstClassText(card, "s-item__price", "s-card__price")
	if priceText == "" {
		priceText = cardText
	}
	if moneyRange.MatchString(priceText) {
		return parsedSale{}, false
	}
	price, currency, ok := parseMoney(priceText)
	if !ok || currency != "USD" {
		return parsedSale{}, false
	}
	date, ok := parseSoldDate(cardText, now)
	if !ok {
		return parsedSale{}, false
	}
	shipping := 0.0
	shippingText := firstClassText(card, "s-item__shipping", "s-card__shipping")
	if parsed, _, found := parseMoney(shippingText); found {
		shipping = parsed
	}

	return parsedSale{
		ItemID: match[1], Title: title, Price: price, Shipping: shipping,
		Currency: currency, SoldAt: date, BestOffer: strings.Contains(strings.ToLower(cardText), "best offer accepted"),
		URL: itemURL.String(),
	}, true
}

func parseMoney(value string) (float64, string, bool) {
	value = strings.TrimSpace(value)
	currency := "USD"
	if strings.Contains(strings.ToUpper(value), "C$") || strings.Contains(strings.ToUpper(value), "CA $") {
		currency = "CAD"
	} else if strings.Contains(strings.ToUpper(value), "£") {
		currency = "GBP"
	} else if strings.Contains(strings.ToUpper(value), "€") {
		currency = "EUR"
	} else if strings.Contains(strings.ToUpper(value), "AU $") {
		currency = "AUD"
	} else if !strings.Contains(value, "$") {
		return 0, "", false
	}
	match := moneyPattern.FindStringSubmatch(value)
	if len(match) != 2 {
		return 0, "", false
	}
	amount, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
	return amount, currency, err == nil
}

func parseSoldDate(text string, now time.Time) (time.Time, bool) {
	lower := strings.ToLower(text)
	index := strings.Index(lower, "sold ")
	if index < 0 {
		return time.Time{}, false
	}
	value := strings.TrimSpace(text[index+len("sold "):])
	date := soldDateRegex.FindString(value)
	if date == "" {
		return time.Time{}, false
	}
	for _, pattern := range datePatterns {
		if parsed, err := time.ParseInLocation(pattern, date, now.Location()); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func firstClassText(node *html.Node, classes ...string) string {
	result := ""
	walk(node, func(child *html.Node) {
		if result != "" || child.Type != html.ElementNode {
			return
		}
		for _, class := range classes {
			if hasClass(child, class) {
				result = strings.TrimSpace(nodeText(child))
				return
			}
		}
	})
	return result
}

func nodeText(node *html.Node) string {
	var builder strings.Builder
	var collect func(*html.Node)
	collect = func(current *html.Node) {
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
			builder.WriteByte(' ')
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			collect(child)
		}
	}
	collect(node)
	return strings.Join(strings.Fields(builder.String()), " ")
}

func walk(node *html.Node, visit func(*html.Node)) {
	visit(node)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walk(child, visit)
	}
}

func hasClass(node *html.Node, class string) bool {
	for _, value := range strings.Fields(attr(node, "class")) {
		if value == class {
			return true
		}
	}
	return false
}

func attr(node *html.Node, key string) string {
	for _, attribute := range node.Attr {
		if attribute.Key == key {
			return attribute.Val
		}
	}
	return ""
}
