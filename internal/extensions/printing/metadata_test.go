package printing

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseMeta(t *testing.T) {
	html := `<html><head>
		<meta property="og:title" content="Cool Bracket">
		<meta property="og:image" content="/media/preview.png">
		<meta name="twitter:image" content="https://cdn.example.com/other.png">
	</head><body>ignored</body></html>`
	base, _ := url.Parse("https://www.printables.com/model/123-cool-bracket")

	got := parseMeta(strings.NewReader(html), base)
	if got.Title != "Cool Bracket" {
		t.Errorf("title = %q, want %q", got.Title, "Cool Bracket")
	}
	// og:image wins over twitter:image and is resolved against the base URL.
	want := "https://www.printables.com/media/preview.png"
	if got.ThumbnailURL != want {
		t.Errorf("thumbnail = %q, want %q", got.ThumbnailURL, want)
	}
}

func TestValidatePublicURLRejectsPrivate(t *testing.T) {
	for _, raw := range []string{
		"ftp://example.com/x",
		"http://127.0.0.1/",
		"http://10.0.0.5/admin",
		"http://169.254.169.254/latest/meta-data",
	} {
		if _, err := validatePublicURL(raw); err == nil {
			t.Errorf("expected %q to be rejected", raw)
		}
	}
}
