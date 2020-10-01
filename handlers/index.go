package handlers

import 
(
	"net/http"
)

// Index handler.
type Index struct
{
}

// NewIndex returns new Index handler.
func NewIndex() *Index 
{
	return &Index{}
}

// ServeHTTP handles requests on incoming connections.
func (i *Index) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" 
	{
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Write([]byte(`<html>
                        <head><title>cam2ip</title></head>
                        <body>
                        <h1>cam2ip</h1>
                        <p><a href='/html'>html</a></p>
                        <p><a href='/jpeg'>jpeg</a></p>
                        <p><a href='/mjpeg'>mjpeg</a></p>
                        </body>
                        </html>`))
}
