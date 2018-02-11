package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const indexTemplStr = `<!doctype html>
<html>
<head>
<title>Plot Viewer</title>
<style>
* {
	margin: 0;
	padding: 0;
}

html {
	height: 100%;
	font-family: sans-serif;
}

body {
	display: flex;
	flex-direction: row;
	height: 100%
}

#sidebar {
	width: 260px;
	height: 100%;
	overflow: scroll;
	padding: 0 10px;
}

#sidebar * {
	max-width: 250px;
}

#sidebar object {
	z-index: -1;
	position: relative;
}

#main-content {
	height: 100%;
	flex-grow: 100;
	display: flex;
	flex-direction: column;
}

.thumb-box:first-child {
	margin-top: 10px;
}

.thumb-box {
	display: block;
	border: 1px solid #ccc;
	border-radius: 5px;
	margin-bottom: 10px;
	padding: 5px;
	cursor: pointer;
}

#im-box {
	margin: 0;
	border: none;
	flex-grow: 100;
}

.button-bar {
	margin: 10px;
}

button {
	padding: 1px 4px;
	font-size: 1.1em;
	background: white;
	border: 1 px solid #ccc;
	border-radius: 5px;
	outline: none;
}

button:focus { outline:0; }
</style>
</head>
<body>
	<div id="sidebar">
	{{range $index, $p := .Plots}}
		<a class="thumb-box" onclick="showImage('/images/{{$index}}')">
			<object data="/images/{{$index}}" type="image/svg+xml" style="pointer-events: none;"></object>
			<div class="imtitle">{{$p}}</div>
		</a>
	{{end}}
	</div>
	<div id="main-content">
		<div class="button-bar">
			<button id="open" onclick="openInNewTab()">Open in Tab</button>
			<button id="cp-png" onclick="openPNG()">Open PNG</button>
			<button id="exit" onclick="quit()">Quit</button>
		</div>
		<object id="im-box" data="/images/0" type="image/svg+xml" style="pointer-events: none;">
		</object>
	</div>
	<script>
	function openInNewTab() {
		var url = document.getElementById("im-box").data;
		var win = window.open(url, '_blank');
		win.focus();
	}
	function showImage(url) {
		document.getElementById("im-box").data = url;
	}
	function openPNG() {
		var imUrl = document.getElementById("im-box").data;
		var imUrlSplit = imUrl.split("/");
		var imID = imUrlSplit[imUrlSplit.length-1];
		var win = window.open("/pngs/" + imID, '_blank');
		win.focus();
	}
	function quit() {
		var r = new XMLHttpRequest();
		r.open("GET", "/quit");
		r.send();
	}
	</script>
</body>
</html>
`

var indexTempl = template.Must(template.New("index").Parse(indexTemplStr))

type plotViewer struct {
	pdfPaths []string
	svgCache [][]byte
	pngCache [][]byte
}

func (v plotViewer) rootHandler(w http.ResponseWriter, r *http.Request) {
	err := indexTempl.Execute(w, struct {
		Plots []string
	}{
		v.pdfPaths,
	})
	if err != nil {
		log.Println("GET /: %v", err)
	}
}

func (v plotViewer) imageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if id < 0 || id >= len(v.pdfPaths) {
		http.Error(w, "image id out of bounds", http.StatusBadRequest)
		return
	}

	if v.svgCache[id] == nil {
		out, err := exec.Command("pdf2svg", v.pdfPaths[id], "/dev/stdout").Output()
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to convert to svg: %v", err), http.StatusInternalServerError)
			return
		}

		v.svgCache[id] = out
	}

	w.Header().Set("content-type", "image/svg+xml")
	w.Write(v.svgCache[id])
}

func (v plotViewer) pngHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if id < 0 || id >= len(v.pdfPaths) {
		http.Error(w, "image id out of bounds", http.StatusBadRequest)
		return
	}

	if v.pngCache[id] == nil {
		out, err := exec.Command("convert", "-density", "300", v.pdfPaths[id], "png:-").Output()
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to convert to png: %v", err), http.StatusInternalServerError)
			return
		}

		v.pngCache[id] = out
	}

	w.Header().Set("content-type", "image/png")
	w.Write(v.pngCache[id])
}

type quitHandler struct {
	s *http.Server
}

func (h quitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(r.Context(), 5*time.Second)
	h.s.Shutdown(ctx)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("vplots: ")
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: vplots plot.pdf...")
		os.Exit(2)
	}

	v := plotViewer{
		pdfPaths: os.Args[1:],
		svgCache: make([][]byte, len(os.Args)-1),
		pngCache: make([][]byte, len(os.Args)-1),
	}

	r := mux.NewRouter()
	r.HandleFunc("/", v.rootHandler)
	r.HandleFunc("/images/{id:[0-9]+}", v.imageHandler)
	r.HandleFunc("/pngs/{id:[0-9]+}", v.pngHandler)

	s := &http.Server{
		Addr:    "127.0.0.1:6544",
		Handler: r,
	}
	r.Handle("/quit", quitHandler{s})
	go openURL(s.Addr)
	s.ListenAndServe()
}

func openURL(url string) {
	url = "http://" + url

	for tries := 5; tries > 5; tries-- {
		if _, err := http.Get(url); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux", "freebsd", "solaris":
		cmd = exec.Command("xdg-open", url)
	default:
		log.Printf("no known open command, connect to %s", url)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("open command failed, connect to %s: %v", url, err)
	}
}
