// Package web serves the game world as a rendered hex map over HTTP.
package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Rioverde/gongeons/internal/game"
)

// Config configures the HTTP server.
type Config struct {
	// TilesDir is the filesystem path to the directory containing terrain tile PNGs.
	TilesDir string

	// Radius is the hex radius passed to world generation.
	Radius int

	// Seed is the seed used for the initial world. When zero a per-request seed must be supplied.
	Seed int64
}

// Server owns the current world and renders it on each HTTP request. The world can be
// regenerated concurrently by passing ?seed=N in the query string.
type Server struct {
	cfg  Config
	tmpl *template.Template

	mu    sync.RWMutex
	world *game.World
	seed  int64
}

// NewServer parses the embedded HTML template and generates the initial world.
func NewServer(cfg Config) (*Server, error) {
	tmpl, err := template.ParseFS(staticFS, "static/index.html.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	s := &Server{cfg: cfg, tmpl: tmpl}
	s.regenerate(cfg.Seed)
	return s, nil
}

// Handler returns the HTTP handler serving the UI, static assets and tile PNGs.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("/static/", http.FileServer(http.FS(staticOnlyFS{})))
	mux.Handle("/tiles/", http.StripPrefix("/tiles/", http.FileServer(http.Dir(s.cfg.TilesDir))))
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if raw := r.URL.Query().Get("seed"); raw != "" {
		if seed, err := strconv.ParseInt(raw, 10, 64); err == nil {
			s.regenerate(seed)
		}
	}

	s.mu.RLock()
	vm := buildViewModel(s.world, s.seed)
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.Execute(w, vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) regenerate(seed int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if seed == 0 {
		seed = s.seed
	}
	world := game.NewWorld()
	world.Generate(s.cfg.Radius, seed)
	s.world = world
	s.seed = seed
}

// staticOnlyFS exposes the embedded static directory while hiding server-only files such as
// Go templates from the public URL space.
type staticOnlyFS struct{}

func (staticOnlyFS) Open(name string) (fs.File, error) {
	if strings.HasSuffix(name, ".tmpl") {
		return nil, fs.ErrNotExist
	}
	return staticFS.Open(name)
}
