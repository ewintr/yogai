package handler

import (
	"ewintr.nl/yogai/storage"
	"fmt"
	"golang.org/x/exp/slog"
	"miniflux.app/logger"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
)

type Server struct {
	apis   map[string]http.Handler
	logger *slog.Logger
}

func NewServer(videoRepo storage.VideoRepository, logger *slog.Logger) *Server {
	return &Server{
		apis: map[string]http.Handler{
			"video": NewVideoAPI(videoRepo, logger),
		},
		logger: logger,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	originalPath := r.URL.Path
	rec := httptest.NewRecorder() // records the response to be able to mix writing headers and content

	w.Header().Add("Content-Type", "application/json")

	// route to api
	head, tail := ShiftPath(r.URL.Path)
	if len(head) == 0 {
		Index(rec)
		returnResponse(w, rec)
		return
	}
	api, ok := s.apis[head]
	if !ok {
		Error(rec, http.StatusNotFound, "Not found", fmt.Errorf("%s is not a valid path", r.URL.Path))
	} else {
		r.URL.Path = tail
		api.ServeHTTP(rec, r)
	}

	returnResponse(w, rec)
	logger.Info("request served", "path", originalPath, "status", rec.Code)
}

func returnResponse(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
	w.WriteHeader(rec.Code)
	for k, v := range rec.Header() {
		w.Header()[k] = v
	}
	w.Write(rec.Body.Bytes())
}

// ShiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
// See https://blog.merovius.de/posts/2017-06-18-how-not-to-use-an-http-router/
func ShiftPath(p string) (string, string) {
	p = path.Clean("/" + p)

	// restore iri prefixes that might be mangled by path.Clean
	for k, v := range map[string]string{
		"http:/":  "http://",
		"https:/": "https://",
	} {
		p = strings.Replace(p, k, v, -1)
	}

	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
