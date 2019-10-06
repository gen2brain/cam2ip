// Package server.
package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/abbot/go-http-auth"

	"github.com/gen2brain/cam2ip/handlers"
	"github.com/gen2brain/cam2ip/reader"
)

// Server struct.
type Server struct {
	Name    string
	Version string

	Bind     string
	Htpasswd string

	Index int
	Delay int

	FrameWidth  float64
	FrameHeight float64

	Rotate int

	NoWebGL bool

	FileName string

	Reader reader.ImageReader
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

	http.Handle("/html", newAuthHandler(handlers.NewHTML(s.FrameWidth, s.FrameHeight, s.NoWebGL), basic))
	http.Handle("/jpeg", newAuthHandler(handlers.NewJPEG(s.Reader), basic))
	http.Handle("/mjpeg", newAuthHandler(handlers.NewMJPEG(s.Reader, s.Delay), basic))
	http.Handle("/socket", newAuthHandler(handlers.NewSocket(s.Reader, s.Delay), basic))

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{}

	listener, err := net.Listen("tcp4", s.Bind)
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
