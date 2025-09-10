package httpserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/Uva337/WBL0v1/internal/interfaces"
)

type Server struct {
	cache interfaces.Cache
	repo  interfaces.Repository
}

func New(c interfaces.Cache, r interfaces.Repository) *Server {
	return &Server{cache: c, repo: r}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/order/", s.handleGetOrder)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "web/index.html")
			return
		}
		http.ServeFile(w, r, "web/"+strings.TrimPrefix(r.URL.Path, "/"))
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{Addr: ":" + port, Handler: logRequests(mux)}
	log.Printf("HTTP listening on :%s", port)
	return srv.ListenAndServe()
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/order/")
	if id == "" {
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}

	if v, ok := s.cache.Get(id); ok {
		writeJSON(w, v, http.StatusOK)
		return
	}

	o, ok, err := s.repo.GetOrder(r.Context(), id)
	if err != nil {
		log.Printf("error getting order from repo: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.NotFound(w, r)
		return
	}

	s.cache.Set(o.OrderUID, o)
	writeJSON(w, o, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error writing json response: %v", err)
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
