package frontend

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestBundledFiles(t *testing.T) {
	handler := Handler()
	err := fs.WalkDir(files, "static", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || path == "static/index.html" {
			return nil
		}
		url := strings.TrimPrefix(path, "static")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, url, nil))
		if w.Code != http.StatusOK || w.Body.Len() == 0 {
			t.Errorf("%s: status %d, length %d", url, w.Code, w.Body.Len())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHomepageAssets(t *testing.T) {
	handler := Handler()
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("homepage: %d %s", w.Code, w.Header())
	}
	if !strings.Contains(w.Body.String(), "Minotar") {
		t.Fatal("missing website")
	}
	doc, err := html.Parse(strings.NewReader(w.Body.String()))
	if err != nil {
		t.Fatal(err)
	}
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if attr.Key == "src" || (node.Data == "link" && attr.Key == "href") {
					if !strings.HasPrefix(attr.Val, "/") || strings.HasPrefix(attr.Val, "//") {
						t.Errorf("external page resource: %s", attr.Val)
					}
					if node.Data != "img" {
						asset := httptest.NewRecorder()
						handler.ServeHTTP(asset, httptest.NewRequest(http.MethodGet, attr.Val, nil))
						if asset.Code != http.StatusOK {
							t.Errorf("missing page asset %s", attr.Val)
						}
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(doc)
	head := httptest.NewRecorder()
	handler.ServeHTTP(head, httptest.NewRequest(http.MethodHead, "/", nil))
	if head.Code != http.StatusOK || head.Body.Len() != 0 {
		t.Fatal("HEAD must return headers without the page body")
	}
}

func TestUnknownPathsAndDirectories(t *testing.T) {
	for _, path := range []string{"/missing", "/assets/", "/addons/ipb/", "/../frontend.go", "/static/index.html"} {
		w := httptest.NewRecorder()
		Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusNotFound {
			t.Errorf("%s: %d", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	Handler().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /: %d", w.Code)
	}
}
