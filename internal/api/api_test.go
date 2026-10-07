package api_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"open-niscat/internal/api"
	"open-niscat/internal/catalog"
	"open-niscat/internal/catalog/catalogtest"
)

func newServer(t *testing.T) http.Handler {
	t.Helper()
	store, err := catalog.Open(context.Background(), catalogtest.NewDataDir(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	web := fstest.MapFS{
		"index.html":    {Data: []byte("<html>app</html>")},
		"assets/app.js": {Data: []byte("console.log(1)")},
	}
	return api.New(store, web, catalog.LangFR, "v1.2.3")
}

func get(t *testing.T, h http.Handler, url string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody))
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON %q: %v", rec.Body.String(), err)
	}
	return out
}

func TestJSONEndpoints(t *testing.T) {
	h := newServer(t)
	cases := []struct {
		url    string
		status int
		check  func(map[string]any) bool
	}{
		{"/api/meta", 200, func(m map[string]any) bool {
			return m["edition"] == "Ed. TEST" && m["defaultLang"] == "fr" && m["appVersion"] == "v1.2.3"
		}},
		{"/api/vin/VSKBEC220U0990494?lang=fr", 200, func(m map[string]any) bool {
			v := m["vehicle"].(map[string]any)
			return v["catalog"].(map[string]any)["cat"] == "AA-G01" && v["attributes"].([]any)[1].(map[string]any)["value"] == "COURT"
		}},
		{"/api/vin/0990494", 200, func(m map[string]any) bool { return len(m["candidates"].([]any)) == 1 }},
		{"/api/vin/ZZZZZZZZZ", 404, func(m map[string]any) bool { return m["error"] == "not_found" }},
		{"/api/catalogs", 200, nil},
		{"/api/catalogs/AA-G01/models", 200, nil},
		{"/api/catalogs/AA-G01/groups?lang=fr", 200, nil},
		{"/api/catalogs/AA-G01/groups/B?vin=VSKBEC220U0990494", 200, func(m map[string]any) bool {
			return len(m["sections"].([]any)) == 3
		}},
		{"/api/catalogs/bad/groups", 400, func(m map[string]any) bool { return m["error"] == "invalid" }},
		{"/api/vehicle?cat=AA-G01&model=BELC220QSKVX", 200, func(m map[string]any) bool { return m["model"] == "BELC220QSKVX" }},
		{"/api/vehicle", 400, nil},
		{"/api/sections/AA/230A?vin=VSKBEC220U0990494&lang=fr", 200, func(m map[string]any) bool {
			return m["applicable"] == true && len(m["lines"].([]any)) == 3
		}},
		{"/api/sections/AA/230A?lang=de", 400, nil},
		{"/api/sections/AA/230A?vin=NOPE", 404, nil},
		{"/api/search?q=palier&type=parts&lang=fr", 200, func(m map[string]any) bool { return m["total"] == float64(2) }},
		{"/api/search?q=%22)%20OR%201%3D1&type=sections", 200, nil},
		{"/api/search?q=x&type=bogus", 400, nil},
		{"/api/search?q=x&limit=abc", 400, nil},
		{"/api/parts/23319-D9700", 200, func(m map[string]any) bool { return len(m["next"].([]any)) == 1 }},
		{"/api/parts/00000-00000", 404, nil},
		{"/api/lines?ids=AA1,AA999", 200, func(m map[string]any) bool {
			return len(m["lines"].([]any)) == 1 && m["missing"].([]any)[0] == "AA999"
		}},
		{"/api/nope", 404, func(m map[string]any) bool { return m["error"] == "not_found" }},
	}
	for _, tc := range cases {
		rec := get(t, h, tc.url)
		if rec.Code != tc.status {
			t.Errorf("%s: status %d, want %d (%s)", tc.url, rec.Code, tc.status, rec.Body.String())
			continue
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("%s: content type %q", tc.url, ct)
		}
		if tc.check != nil && !tc.check(decode(t, rec)) {
			t.Errorf("%s: unexpected body %s", tc.url, rec.Body.String())
		}
	}
}

func TestFiles(t *testing.T) {
	h := newServer(t)
	cases := map[string]int{
		"/files/img/AA/AA230A.png":  200,
		"/files/cinfo/AA/G0101.pdf": 200,
		"/files/img/AA/missing.png": 404,
		"/files/data.db":            404,
		"/files/manifest.json":      404,
		"/files/img/":               404,
	}
	for url, want := range cases {
		if rec := get(t, h, url); rec.Code != want {
			t.Errorf("%s: status %d, want %d", url, rec.Code, want)
		}
	}
	// Traversal attempts: ServeMux answers with a redirect to the cleaned path; data.db must never be served.
	for _, url := range []string{"/files/img/../data.db", "/files/img/%2e%2e/data.db", "/files/img/..%2fdata.db"} {
		if rec := get(t, h, url); rec.Code == http.StatusOK {
			t.Errorf("%s: served with 200", url)
		}
	}
}

func TestSPAAndHeaders(t *testing.T) {
	h := newServer(t)
	for _, url := range []string{"/", "/fr/section/AA/230A?vin=X", "/list?items=AA1x2"} {
		rec := get(t, h, url)
		if rec.Code != 200 || rec.Body.String() != "<html>app</html>" {
			t.Errorf("%s: %d %q", url, rec.Code, rec.Body.String())
		}
		if rec.Header().Get("X-Robots-Tag") != "noindex, nofollow" {
			t.Errorf("%s: missing X-Robots-Tag", url)
		}
	}
	if rec := get(t, h, "/assets/app.js"); rec.Code != 200 || rec.Body.String() != "console.log(1)" {
		t.Errorf("asset: %d %q", rec.Code, rec.Body.String())
	}
}

func TestSPAWithoutBuiltFrontend(t *testing.T) {
	store, err := catalog.Open(context.Background(), catalogtest.NewDataDir(t))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	h := api.New(store, fstest.MapFS{".gitkeep": {}}, catalog.LangEN, "dev")
	if rec := get(t, h, "/"); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestMethodNotAllowedIsJSON(t *testing.T) {
	h := newServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/catalogs", http.NoBody))
	if rec.Code != http.StatusMethodNotAllowed || decode(t, rec)["error"] != "method_not_allowed" || rec.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("POST /api/catalogs = %d %q allow=%q", rec.Code, rec.Body.String(), rec.Header().Get("Allow"))
	}
}

func TestMissingAssetIs404(t *testing.T) {
	h := newServer(t)
	if rec := get(t, h, "/assets/old-1234.js"); rec.Code != http.StatusNotFound {
		t.Fatalf("missing asset = %d %q", rec.Code, rec.Body.String())
	}
}

func TestGroupDetailUnknownCatalog(t *testing.T) {
	h := newServer(t)
	if rec := get(t, h, "/api/catalogs/AA-G99/groups/B"); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown catalog = %d", rec.Code)
	}
}

func TestCancelledRequestIsNotLogged(t *testing.T) {
	h := newServer(t)
	var logs strings.Builder
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/search?q=palier", http.NoBody))
	if logs.Len() != 0 {
		t.Fatalf("cancelled request logged: %s", logs.String())
	}
}

// TestWireShapes pins the JSON contract the frontend types rely on.
func TestWireShapes(t *testing.T) {
	h := newServer(t)
	sec := decode(t, get(t, h, "/api/sections/AA/230A?vin=VSKBEC220U0990494&lang=fr"))
	line, _ := sec["lines"].([]any)[0].(map[string]any)
	if sec["from"] != "1987-04" || line["from"] != "1987-04" || line["to"] != "1987-04" {
		t.Errorf("section/line dates = %v / %v %v", sec["from"], line["from"], line["to"])
	}
	for _, gone := range []string{"dataplic", "period"} {
		if _, ok := line[gone]; ok {
			t.Errorf("line still has %q", gone)
		}
	}
	hits := decode(t, get(t, h, "/api/search?q=alternateur&type=sections&lang=fr"))
	hit, _ := hits["sections"].([]any)[0].(map[string]any)
	if g, ok := hit["group"].(map[string]any); !ok || g["code"] != "B" || g["label"] == "" {
		t.Errorf("section hit group = %v", hit["group"])
	}
	v := decode(t, get(t, h, "/api/vin/VSKBEC220U0990494"))
	vehicle, _ := v["vehicle"].(map[string]any)
	cat, _ := vehicle["catalog"].(map[string]any)
	if vehicle["prodDate"] != "1989-05" || cat["from"] != "1987-04" {
		t.Errorf("vehicle prodDate/catalog from = %v / %v", vehicle["prodDate"], cat["from"])
	}
	if langs, ok := cat["langs"].([]any); !ok || len(langs) != 2 {
		t.Errorf("catalog langs = %v", cat["langs"])
	}
}
