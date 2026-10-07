// Package api exposes the catalog as a read-only JSON API and serves data files and the frontend.
package api

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"open-niscat/internal/catalog"
)

type server struct {
	store       *catalog.Store
	web         fs.FS
	defaultLang catalog.Lang
	appVersion  string
}

type handlerFunc func(r *http.Request) (any, error)

// New returns the HTTP handler: /api/... JSON, /files/... data files, /health probe, anything else the SPA.
func New(store *catalog.Store, web fs.FS, defaultLang catalog.Lang, appVersion string) http.Handler {
	s := &server{store: store, web: web, defaultLang: defaultLang, appVersion: appVersion}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/meta", s.json(s.meta))
	mux.HandleFunc("GET /api/vin/{vin}", s.json(s.vin))
	mux.HandleFunc("GET /api/catalogs", s.json(s.catalogs))
	mux.HandleFunc("GET /api/catalogs/{cat}/models", s.json(s.models))
	mux.HandleFunc("GET /api/catalogs/{cat}/groups", s.json(s.groups))
	mux.HandleFunc("GET /api/catalogs/{cat}/groups/{group}", s.json(s.groupDetail))
	mux.HandleFunc("GET /api/vehicle", s.json(s.vehicle))
	mux.HandleFunc("GET /api/sections/{etd}/{sec}", s.json(s.section))
	mux.HandleFunc("GET /api/search", s.json(s.search))
	mux.HandleFunc("GET /api/parts/{ref}", s.json(s.part))
	mux.HandleFunc("GET /api/lines", s.json(s.lines))
	// Must carry the method: "/api/" next to "GET /" is an ambiguous pattern pair and makes ServeMux panic.
	mux.HandleFunc("GET /api/", s.json(func(*http.Request) (any, error) {
		return nil, fmt.Errorf("%w: unknown API route", catalog.ErrNotFound)
	}))
	mux.Handle("GET /files/", http.StripPrefix("/files", s.files()))
	mux.HandleFunc("GET /", s.spa)
	return noIndex(apiMethods(mux))
}

// apiMethods answers non-GET API requests with a JSON 405 (ServeMux would answer text/plain).
func apiMethods(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") && r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method_not_allowed", Message: "only GET is supported"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func noIndex(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		next.ServeHTTP(w, r)
	})
}

func (s *server) json(fn handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v, err := fn(r)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, v)
	}
}

func lang(r *http.Request) (catalog.Lang, error) {
	v := r.URL.Query().Get("lang")
	if v == "" { // English: defaultLang only tells the front which language to offer first (/api/meta)
		return catalog.LangEN, nil
	}
	l, err := catalog.ParseLang(v)
	if err != nil {
		return "", fmt.Errorf("lang: %w", err)
	}
	return l, nil
}

func (s *server) scope(r *http.Request) (*catalog.Scope, error) {
	q := r.URL.Query()
	sc, err := s.store.ResolveScope(r.Context(), catalog.ScopeParams{VIN: q.Get("vin"), Cat: q.Get("cat"), Model: q.Get("model")})
	if err != nil {
		return nil, fmt.Errorf("scope: %w", err)
	}
	return sc, nil
}

// langAndScope parses the common parameters.
func (s *server) langAndScope(r *http.Request) (catalog.Lang, *catalog.Scope, error) {
	l, err := lang(r)
	if err != nil {
		return "", nil, err
	}
	sc, err := s.scope(r)
	if err != nil {
		return "", nil, err
	}
	return l, sc, nil
}

// health is the container probe: 200 while the database answers, 503 otherwise.
func (s *server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		if !errors.Is(err, context.Canceled) {
			slog.Error("health check failed", "err", err)
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) meta(*http.Request) (any, error) {
	m := s.store.Manifest()
	return map[string]any{
		"schema": m.Schema, "version": m.Version, "source": m.Source, "build": m.Build,
		"defaultLang": string(s.defaultLang), "appVersion": s.appVersion,
	}, nil
}

func (s *server) vin(r *http.Request) (any, error) {
	l, err := lang(r)
	if err != nil {
		return nil, err
	}
	res, err := s.store.IdentifyVIN(r.Context(), r.PathValue("vin"), l)
	if err != nil {
		return nil, fmt.Errorf("identify VIN: %w", err)
	}
	return res, nil
}

func (s *server) catalogs(r *http.Request) (any, error) {
	cats, err := s.store.Catalogs(r.Context())
	if err != nil {
		return nil, fmt.Errorf("catalogs: %w", err)
	}
	return cats, nil
}

func (s *server) models(r *http.Request) (any, error) {
	l, err := lang(r)
	if err != nil {
		return nil, err
	}
	etd, grupo, err := catalog.ParseCat(r.PathValue("cat"))
	if err != nil {
		return nil, fmt.Errorf("models: %w", err)
	}
	models, err := s.store.Models(r.Context(), etd, grupo, l)
	if err != nil {
		return nil, fmt.Errorf("models: %w", err)
	}
	return models, nil
}

func (s *server) groups(r *http.Request) (any, error) {
	l, err := lang(r)
	if err != nil {
		return nil, err
	}
	etd, grupo, err := catalog.ParseCat(r.PathValue("cat"))
	if err != nil {
		return nil, fmt.Errorf("groups: %w", err)
	}
	if _, err := s.store.Catalog(r.Context(), etd, grupo); err != nil {
		return nil, fmt.Errorf("groups: %w", err)
	}
	groups, err := s.store.Groups(r.Context(), etd, l)
	if err != nil {
		return nil, fmt.Errorf("groups: %w", err)
	}
	return groups, nil
}

func (s *server) groupDetail(r *http.Request) (any, error) {
	l, sc, err := s.langAndScope(r)
	if err != nil {
		return nil, err
	}
	etd, grupo, err := catalog.ParseCat(r.PathValue("cat"))
	if err != nil {
		return nil, fmt.Errorf("group: %w", err)
	}
	if _, err := s.store.Catalog(r.Context(), etd, grupo); err != nil {
		return nil, fmt.Errorf("group: %w", err)
	}
	d, err := s.store.GroupDetail(r.Context(), etd, grupo, strings.ToUpper(r.PathValue("group")), sc, l)
	if err != nil {
		return nil, fmt.Errorf("group: %w", err)
	}
	return d, nil
}

func (s *server) vehicle(r *http.Request) (any, error) {
	l, sc, err := s.langAndScope(r)
	if err != nil {
		return nil, err
	}
	v, err := s.store.Vehicle(r.Context(), sc, l)
	if err != nil {
		return nil, fmt.Errorf("vehicle: %w", err)
	}
	return v, nil
}

func (s *server) section(r *http.Request) (any, error) {
	l, sc, err := s.langAndScope(r)
	if err != nil {
		return nil, err
	}
	sec, err := s.store.Section(r.Context(), r.PathValue("etd"), r.PathValue("sec"), sc, l)
	if err != nil {
		return nil, fmt.Errorf("section: %w", err)
	}
	return sec, nil
}

func intParam(r *http.Request, name string) (int, error) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%w: %s must be a non-negative integer", catalog.ErrInvalid, name)
	}
	return n, nil
}

func (s *server) search(r *http.Request) (any, error) {
	l, sc, err := s.langAndScope(r)
	if err != nil {
		return nil, err
	}
	kind, err := catalog.ParseSearchKind(r.URL.Query().Get("type"))
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	limit, err := intParam(r, "limit")
	if err != nil {
		return nil, err
	}
	offset, err := intParam(r, "offset")
	if err != nil {
		return nil, err
	}
	q := catalog.SearchQuery{Q: r.URL.Query().Get("q"), Kind: kind, Limit: limit, Offset: offset}
	res, err := s.store.Search(r.Context(), q, sc, l)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	return res, nil
}

func (s *server) part(r *http.Request) (any, error) {
	l, err := lang(r)
	if err != nil {
		return nil, err
	}
	p, err := s.store.Part(r.Context(), r.PathValue("ref"), l)
	if err != nil {
		return nil, fmt.Errorf("part: %w", err)
	}
	return p, nil
}

func (s *server) lines(r *http.Request) (any, error) {
	l, err := lang(r)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, id := range strings.Split(r.URL.Query().Get("ids"), ",") {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	res, err := s.store.Lines(r.Context(), ids, l)
	if err != nil {
		return nil, fmt.Errorf("lines: %w", err)
	}
	return res, nil
}

// files serves img/, gindex/ and cinfo/ from the data directory; nothing else, no directory listings.
func (s *server) files() http.Handler {
	root := os.DirFS(s.store.DataDir())
	fileServer := http.FileServerFS(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean("/" + r.URL.Path)
		kind, _, _ := strings.Cut(strings.TrimPrefix(clean, "/"), "/")
		allowed := kind == "img" || kind == "gindex" || kind == "cinfo"
		if !allowed || strings.HasSuffix(r.URL.Path, "/") || clean != r.URL.Path {
			writeError(w, fmt.Errorf("%w: file", catalog.ErrNotFound))
			return
		}
		if st, err := fs.Stat(root, strings.TrimPrefix(clean, "/")); err != nil || st.IsDir() {
			writeError(w, fmt.Errorf("%w: file", catalog.ErrNotFound))
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=86400")
		fileServer.ServeHTTP(w, r)
	})
}

// spa serves built assets as-is and index.html for every other path (client-side routing).
func (s *server) spa(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name != "" && name != "index.html" {
		if st, err := fs.Stat(s.web, name); err == nil && !st.IsDir() {
			http.ServeFileFS(w, r, s.web, name)
			return
		}
	}
	if strings.HasPrefix(name, "assets/") {
		http.NotFound(w, r) // a stale hashed asset must not receive the HTML page
		return
	}
	index, err := fs.ReadFile(s.web, "index.html")
	if err != nil {
		http.Error(w, "frontend not built: run `make web` first", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(index)
}
