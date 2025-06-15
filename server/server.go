// Package server.
package server

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/abbot/go-http-auth"

	"github.com/gen2brain/cam2ip/handlers"
)

// Server struct.
type Server struct {
	Name    string
	Version string

	Index int
	Delay int

	Width  float64
	Height float64

	Quality int
	Rotate  int
	Flip    string

	NoWebGL bool

	Timestamp  bool
	TimeFormat string

	Bind     string
	Htpasswd string

	Reader handlers.ImageReader
}

// NewServer returns new Server.
func NewServer() *Server {
	s := &Server{}

	return s
}

// ListenAndServe listens on the TCP address and serves requests.
func (s *Server) ListenAndServe() error {
	var basic *auth.BasicAuth
	if s.Htpasswd != "" {
		realm := fmt.Sprintf("%s/%s", s.Name, s.Version)
		basic = auth.NewBasicAuthenticator(realm, auth.HtpasswdFileProvider(s.Htpasswd))
	}

	http.Handle("/html", newAuthHandler(handlers.NewHTML(s.Width, s.Height, s.NoWebGL), basic))
	http.Handle("/jpeg", newAuthHandler(handlers.NewJPEG(s.Reader, s.Quality), basic))
	http.Handle("/mjpeg", newAuthHandler(handlers.NewMJPEG(s.Reader, s.Delay, s.Quality), basic))
	http.Handle("/socket", newAuthHandler(handlers.NewSocket(s.Reader, s.Delay, s.Quality), basic))

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.Handle("/", newAuthHandler(handlers.NewIndex(), basic))

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	listener, err := net.Listen("tcp", s.Bind)
	if err != nil {
		return err
	}

	return srv.Serve(listener)
}

// newAuthHandler wraps handler and checks auth.
func newAuthHandler(handler http.Handler, authenticator *auth.BasicAuth) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authenticator != nil {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", authenticator.Realm))
			if authenticator.CheckAuth(r) == "" {
				http.Error(w, "401 Unauthorized", http.StatusUnauthorized)

				return
			}
		}

		handler.ServeHTTP(w, r)
	})
}
