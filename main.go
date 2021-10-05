package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"text/template"
)

var indexTextTemplate = template.Must(template.New("Index").Parse(`# Properties

Runtime: {{.Runtime}}
TARGETPLATFORM: {{.TARGETPLATFORM}}
GOOS: {{.GOOS}}
GOARCH: {{.GOARCH}}

# Request Headers
{{ range $header := .RequestHeaders}}
{{- range .Values }}
{{$header.Name}}: {{.}}
{{- end}}
{{- end}}
`))

var indexTemplate = template.Must(template.New("Index").Parse(`<!DOCTYPE html>
<html>
<head>
<title>example-docker-buildx-go</title>
<style>
body {
	font-family: monospace;
	color: #555;
	background: #e6edf4;
	padding: 1.25rem;
	margin: 0;
}
table {
	background: #fff;
	border: .0625rem solid #c4cdda;
	border-radius: 0 0 .25rem .25rem;
	border-spacing: 0;
	margin-bottom: 1.25rem;
	padding: .75rem 1.25rem;
	text-align: left;
	white-space: pre;
}
table > caption {
	background: #f1f6fb;
	text-align: left;
	font-weight: bold;
	padding: .75rem 1.25rem;
	border: .0625rem solid #c4cdda;
	border-radius: .25rem .25rem 0 0;
	border-bottom: 0;
}
table td, table th {
	padding: .25rem;
}
table > tbody > tr:hover {
	background: #f1f6fb;
}
</style>
</head>
<body>
	<table>
		<caption>Properties</caption>
		<tbody>
			<tr>
				<th>Runtime</th>
				<td>{{.Runtime}}</td>
			</tr>
			<tr>
				<th>TARGETPLATFORM</th>
				<td>{{.TARGETPLATFORM}}</td>
			</tr>
			<tr>
				<th>GOOS</th>
				<td>{{.GOOS}}</td>
			</tr>
			<tr>
				<th>GOARCH</th>
				<td>{{.GOARCH}}</td>
			</tr>
		</tbody>
	</table>
	<table>
		<caption>Request Headers</caption>
		<tbody>
			{{- range $header := .RequestHeaders}}
			{{- range .Values }}
			<tr>
				<th>{{$header.Name}}</th>
				<td>{{.}}</td>
			</tr>
			{{- end}}
			{{- end}}
		</tbody>
	</table>
</body>
</html>
`))

type header struct {
	Name   string
	Values []string
}

type headers []header

func headersFromHttpHeaders(httpHeaders http.Header) headers {
	result := make(headers, 0, len(httpHeaders))
	for k := range httpHeaders {
		result = append(result, header{
			Name:   k,
			Values: httpHeaders[k],
		})
	}
	return result
}

func (a headers) Len() int           { return len(a) }
func (a headers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a headers) Less(i, j int) bool { return strings.ToLower(a[i].Name) < strings.ToLower(a[j].Name) }

type indexData struct {
	Runtime        string
	GOOS           string
	GOARCH         string
	TARGETPLATFORM string
	RequestHeaders headers
}

var (
	TARGETPLATFORM string // NB this is set by the linker with -X.
)

func main() {
	log.SetFlags(0)

	log.Printf("%s", runtime.Version())
	log.Printf("TARGETPLATFORM=%s", TARGETPLATFORM)
	log.Printf("GOOS=%s", runtime.GOOS)
	log.Printf("GOARCH=%s", runtime.GOARCH)
	//log.Printf("GOARM=%s", runtime.GOARM) // NB there is no GOARM.

	var listenAddress = flag.String("listen", "", "Listen address")

	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
		log.Fatalf("\nERROR You MUST NOT pass any positional arguments")
	}

	if *listenAddress == "" || *listenAddress == "no" {
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		headers := headersFromHttpHeaders(r.Header)
		sort.Sort(headers)

		var t *template.Template
		var contentType string

		switch r.URL.Query().Get("format") {
		case "text":
			t = indexTextTemplate
			contentType = "text/plain"
		default:
			t = indexTemplate
			contentType = "text/html"
		}

		w.Header().Set("Content-Type", contentType)

		err := t.ExecuteTemplate(w, "Index", indexData{
			Runtime:        runtime.Version(),
			TARGETPLATFORM: TARGETPLATFORM,
			GOOS:           runtime.GOOS,
			GOARCH:         runtime.GOARCH,
			//GOARM:          runtime.GOARM, // NB there is no GOARM.
			RequestHeaders: headers,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Printf("Listening at http://%s\n", *listenAddress)

	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatalf("Failed to ListenAndServe: %v", err)
	}
}
