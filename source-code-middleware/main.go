package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/panic/", panicDemo)
	mux.HandleFunc("/panic-after/", panicAfterDemo)
	mux.HandleFunc("/debug/", sourceCodeHandler)

	http.ListenAndServe(":8080", DevMw(mux))
}

func DevMw(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				stack := debug.Stack()
				log.Println(string(stack))
				fmt.Fprintf(w, "<h1>panic: %v</h1><pre>%s</pre>", err, makeLinks(string(stack)))
			}
		}()
		next.ServeHTTP(w, r)
	}
}

func panicDemo(w http.ResponseWriter, r *http.Request) {
	funcThatPanics()
}
func panicAfterDemo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Panic after demo\n")
	funcThatPanics()
}

func funcThatPanics() {
	panic("Oh no!")
}

func sourceCodeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	lineStr := r.FormValue("line")
	line, err := strconv.Atoi(lineStr)
	if err != nil {
		line = -1
	}
	// path := "/Users/ilyas/Desktop/go-excersice/source-code-middleware/cmd/api/main.go"
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// sfile, err := io.ReadAll(file)
	// if err != nil {
	// 	panic(err)
	// }
	b := bytes.NewBuffer(nil)
	_, err = io.Copy(b, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var lines [][2]int
	if line > 0 {
		lines = append(lines, [2]int{line, line})
	}
	lexer := lexers.Get("go")
	iterator, err := lexer.Tokenise(nil, b.String())
	style := styles.Get("github")
	if style == nil {
		style = styles.Fallback
	}
	formatter := html.New(html.TabWidth(2), html.WithLineNumbers(true), html.HighlightLines(lines))
	w.Header().Set("Content-Type", "text/html")
	formatter.Format(w, style, iterator)
	// _ = quick.Highlight(w, b.String(), "go", "html", "monokai")
	// err = quick.Highlight(w, b.String(), "go", "html", "monokai")
	// if err != nil {
	// 	panic(err)
	// }

	// io.Copy(w, file)
}

func makeLinks(stack string) string {
	lines := strings.Split(stack, "\n")
	for li, line := range lines {
		if len(line) == 0 || line[0] != '\t' {
			continue
		}

		file := ""
		for i, char := range line {
			if char == ':' {
				file = line[1:i]
				break
			}
		}

		var lineStr strings.Builder
		for i := len(file) + 2; i < len(line); i++ {
			if line[i] < '0' || line[i] > '9' {
				break
			}
			lineStr.WriteByte(line[i])
		}
		// line, err := strconv.Atoi(lineStr.String())
		// if err != nil {
		// 	line = 0
		// }
		v := url.Values{}
		v.Set("path", file)
		v.Set("line", lineStr.String())
		lines[li] = "\t<a href=\"/debug/?" + v.Encode() + "\">" + file + ":" + lineStr.String() + "</a>" + line[len(file)+2+len(lineStr.String()):]
	}

	return strings.Join(lines, "\n")
}
