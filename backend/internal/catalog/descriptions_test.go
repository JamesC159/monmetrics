package catalog

import "testing"

func TestCleanCardTextConvertsTCGplayerMarkupToPlainText(t *testing.T) {
	cases := map[string]string{
		"Each case includes 6 Pikachu VMAX Special Collections. <br><br> Includes five Crown Zenith booster packs.": "Each case includes 6 Pikachu VMAX Special Collections.\n\nIncludes five Crown Zenith booster packs.",
		"Summoning Star, \"VSTAR\"\n<br>Once per game.":                                                             "Summoning Star, \"VSTAR\"\n\nOnce per game.",
		"Ready<br/>Set<BR />Go":                         "Ready\nSet\nGo",
		"<em>(Each pack contains 10 random cards)</em>": "(Each pack contains 10 random cards)",
		"Tom &amp; Jerry &lt;b&gt;":                     "Tom & Jerry <b>",
		"Attack does 10 < 20 damage":                    "Attack does 10 < 20 damage",
		"A<br>\n<br>\nB":                                "A\n\nB",
	}
	for input, want := range cases {
		if got := cleanCardText(input); got != want {
			t.Errorf("cleanCardText(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCleanCardTextIsIdempotentOnFeedText(t *testing.T) {
	once := cleanCardText("A Legendary Storm Is Stirring!<br>\nEach blister comes with 3 booster packs.<br>\n<em>Bonus</em>")
	if again := cleanCardText(once); again != once {
		t.Errorf("second pass changed %q to %q", once, again)
	}
}

func TestNormalizeProductStoresCleanDescription(t *testing.T) {
	product := Product{ProductID: 9, Name: "Pokemon Case", ExtendedData: []ExtendedField{
		{Name: "CardText", DisplayName: "Card Text", Value: "6 packs.<br><br>Includes five."},
	}}
	card := NormalizeProduct("tcgcsv", Category{CategoryID: 3, DisplayName: "Pokemon"}, Group{GroupID: 1, Name: "Set"}, product)
	if want := "6 packs.\n\nIncludes five."; card.Description != want {
		t.Errorf("description = %q, want %q", card.Description, want)
	}
}
