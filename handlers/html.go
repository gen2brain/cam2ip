// Package handlers.
package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

// HTML handler.
type HTML struct {
	Template []byte
}

// NewHTML returns new HTML handler.
func NewHTML(bind string, width, height float64) *HTML {
	h := &HTML{}

	b := strings.Split(bind, ":")
	if b[0] == "" {
		bind = "127.0.0.1" + bind
	}

	html = strings.Replace(html, "{BIND}", bind, -1)
	html = strings.Replace(html, "{WIDTH}", fmt.Sprintf("%.0f", width), -1)
	html = strings.Replace(html, "{HEIGHT}", fmt.Sprintf("%.0f", height), -1)

	h.Template = []byte(html)
	return h
}

// ServeHTTP handles requests on incoming connections.
func (h *HTML) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		msg := fmt.Sprintf("405 Method Not Allowed (%s)", r.Method)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(h.Template)
}

var html = `<html>
    <head>
        <title>cam2ip</title>
        <script>
        var url = "ws://{BIND}/socket";
        ws = new WebSocket(url);
	var image = new Image();

	ws.onopen = function() {
		var context = document.getElementById("canvas").getContext("2d");
		image.onload = function() {
			context.drawImage(image, 0, 0);
		}
	}

        ws.onmessage = function(e) {
		image.setAttribute("src", "data:image/jpeg;base64," + e.data);
        }
        </script>
    </head>
	<body style="background-color: #000000">
		<table style="width:100%; height:100%">
			<tr style="height:100%">
				<td style="height:100%; text-align:center">
					<canvas id="canvas" width="{WIDTH}" height="{HEIGHT}"></canvas>
				</td>
			</tr>
		</table>
	</body>
</html>`
