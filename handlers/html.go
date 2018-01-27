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
func NewHTML(bind string, width, height float64, nogl bool) *HTML {
	h := &HTML{}

	b := strings.Split(bind, ":")
	if b[0] == "" {
		bind = "127.0.0.1" + bind
	}

	tpl := htmlWebGL
	if nogl {
		tpl = html
	}

	tpl = strings.Replace(tpl, "{BIND}", bind, -1)
	tpl = strings.Replace(tpl, "{WIDTH}", fmt.Sprintf("%.0f", width), -1)
	tpl = strings.Replace(tpl, "{HEIGHT}", fmt.Sprintf("%.0f", height), -1)

	h.Template = []byte(tpl)
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
        <meta charset="utf-8"/>
        <title>cam2ip</title>
        <script>
        ws = new WebSocket("ws://{BIND}/socket");
        var image = new Image();

        ws.onopen = function() {
            var context = document.getElementById("canvas").getContext("2d", {alpha: false});
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

var htmlWebGL = `<html>
    <head>
        <meta charset="utf-8"/>
        <title>cam2ip</title>
        <script>
		var texture, vloc, tloc, vertexBuff, textureBuff;

		ws = new WebSocket("ws://{BIND}/socket");
		var image = new Image();

		ws.onopen = function() {
			var gl = document.getElementById('canvas').getContext('webgl');

			var vertexShaderSrc =
				"attribute vec2 aVertex;" +
				"attribute vec2 aUV;" +
				"varying vec2 vTex;" +
				"void main(void) {" +
				"  gl_Position = vec4(aVertex, 0.0, 1.0);" +
				"  vTex = aUV;" +
				"}";

			var fragmentShaderSrc =
				"precision mediump float;" +
				"varying vec2 vTex;" +
				"uniform sampler2D sampler0;" +
				"void main(void){" +
				"  gl_FragColor = texture2D(sampler0, vTex);"+
				"}";

			var vertShaderObj = gl.createShader(gl.VERTEX_SHADER);
			var fragShaderObj = gl.createShader(gl.FRAGMENT_SHADER);
			gl.shaderSource(vertShaderObj, vertexShaderSrc);
			gl.shaderSource(fragShaderObj, fragmentShaderSrc);
			gl.compileShader(vertShaderObj);
			gl.compileShader(fragShaderObj);

			var program = gl.createProgram();
			gl.attachShader(program, vertShaderObj);
			gl.attachShader(program, fragShaderObj);

			gl.linkProgram(program);
			gl.useProgram(program);

			gl.viewport(0, 0, {WIDTH}, {HEIGHT});

			vertexBuff = gl.createBuffer();
			gl.bindBuffer(gl.ARRAY_BUFFER, vertexBuff);
			gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, 1, -1, -1, 1, -1, 1, 1]), gl.STATIC_DRAW);

			textureBuff = gl.createBuffer();
			gl.bindBuffer(gl.ARRAY_BUFFER, textureBuff);
			gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([0, 1, 0, 0, 1, 0, 1, 1]), gl.STATIC_DRAW);

			vloc = gl.getAttribLocation(program, "aVertex");
			tloc = gl.getAttribLocation(program, "aUV");

			texture = gl.createTexture();

			image.onload = function() {
				gl.bindTexture(gl.TEXTURE_2D, texture);

				gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
				gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
				gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
				gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);

				gl.pixelStorei(gl.UNPACK_FLIP_Y_WEBGL, true);
				gl.texImage2D(gl.TEXTURE_2D, 0,  gl.RGBA,  gl.RGBA, gl.UNSIGNED_BYTE, image);

				gl.bindBuffer(gl.ARRAY_BUFFER, vertexBuff);
				gl.enableVertexAttribArray(vloc);
				gl.vertexAttribPointer(vloc, 2, gl.FLOAT, false, 0, 0);

				gl.bindBuffer(gl.ARRAY_BUFFER, textureBuff);
				gl.enableVertexAttribArray(tloc);
				gl.vertexAttribPointer(tloc, 2, gl.FLOAT, false, 0, 0);

				gl.drawArrays(gl.TRIANGLE_FAN, 0, 4);
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
