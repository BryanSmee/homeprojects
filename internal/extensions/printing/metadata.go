package printing

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html"
)

const (
	fetchTimeout = 6 * time.Second
	maxBodyBytes = 1 << 20 // 1 MiB is plenty for <head> metadata
	userAgent    = "HomeProjects/1.0 (+link-thumbnail)"
)

type linkPreview struct {
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

// resolveLinkPreview fetches a page and extracts its OpenGraph image/title.
// Works across Printables, Thingiverse, MakerWorld, Cults3D, etc., which all
// emit og:image. Best-effort: returns an empty preview if nothing is found.
func resolveLinkPreview(ctx context.Context, rawURL string) (linkPreview, error) {
	u, err := validatePublicURL(rawURL)
	if err != nil {
		return linkPreview{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return linkPreview{}, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := previewClient().Do(req)
	if err != nil {
		return linkPreview{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return linkPreview{}, fmt.Errorf("fetch %s: status %d", u, resp.StatusCode)
	}

	return parseMeta(io.LimitReader(resp.Body, maxBodyBytes), resp.Request.URL), nil
}

func previewClient() *http.Client {
	return &http.Client{
		Timeout: fetchTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			_, err := validatePublicURL(req.URL.String())
			return err
		},
	}
}

// parseMeta scans HTML for OpenGraph/Twitter image and title meta tags.
func parseMeta(r io.Reader, base *url.URL) linkPreview {
	var out linkPreview
	z := html.NewTokenizer(r)
	for {
		switch z.Next() {
		case html.ErrorToken:
			return out
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == "meta" {
				applyMeta(&out, t, base)
			}
		}
	}
}

func applyMeta(out *linkPreview, t html.Token, base *url.URL) {
	var key, content string
	for _, a := range t.Attr {
		switch a.Key {
		case "property", "name":
			key = a.Val
		case "content":
			content = a.Val
		}
	}
	if content == "" {
		return
	}
	switch key {
	case "og:image", "og:image:url", "og:image:secure_url", "twitter:image":
		if out.ThumbnailURL == "" {
			out.ThumbnailURL = absURL(base, content)
		}
	case "og:title", "twitter:title":
		if out.Title == "" {
			out.Title = content
		}
	}
}

func absURL(base *url.URL, ref string) string {
	r, err := url.Parse(ref)
	if err != nil || base == nil {
		return ref
	}
	return base.ResolveReference(r).String()
}

// validatePublicURL allows only http(s) URLs whose host resolves exclusively to
// public IPs, mitigating SSRF against internal services.
func validatePublicURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return nil, fmt.Errorf("host %q is not allowed", u.Hostname())
		}
	}
	return u, nil
}

func isPublicIP(ip net.IP) bool {
	return !(ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast())
}
