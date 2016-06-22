
package main
import (
	"net/http"
	"fmt"
	"io/ioutil"
	"regexp"
	"os"
	"path/filepath"
	"log"
	"mime"
)

func main() {

	log.SetFlags(log.LstdFlags)

	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s address\b", os.Args[0])
	}

	var cwd string
	if c, err := os.Getwd(); err != nil {
		log.Fatalf("Failed to get CWD: %s\n", err.Error())
	} else {
		cwd = c
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		var path string = filepath.Clean(cwd + r.URL.Path)
		if path == cwd {
			path = cwd + "/index.html"
		}
		if !filepath.HasPrefix(path, cwd) {
			writeResponse(w, r, 400, "ERROR: Invalid path")
			return
		}
		if matched, _ := regexp.MatchString("[a-zA-Z0-9_\\-\\./]+", path); !matched {
			writeResponse(w, r, 400, "ERROR: Invalid path")
			return
		}

		if r.Method == "PUT" {

			// read body
			defer r.Body.Close()
			content, err := ioutil.ReadAll(r.Body)
			if err != nil {
				writeResponse(w, r, 500, fmt.Sprintf("Failed to write: %s", err.Error()))
				return
			}

			// write file
			err = ioutil.WriteFile(path, content, 0644)
			if err != nil {
				writeResponse(w, r, 500, fmt.Sprintf("Failed to write: %s", err.Error()))
				return
			}

			// response
			writeResponse(w, r, 204, "")

		} else if r.Method == "GET" {

			// check file exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				writeResponse(w, r, 404, "Not found")
				return
			}

			// read file
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				writeResponse(w, r, 500, fmt.Sprintf("Failed to read: %s", err.Error()))
				return
			}

			// write body
			w.Header().Add("Content-Type", mime.TypeByExtension(filepath.Ext(path)))

			writeResponseBytes(w, r, 200, bytes)
			return

		} else if r.Method == "DELETE" {

			// check file exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				writeResponse(w, r, 404, "Not found")
				return
			}

			// delete file
			if err := os.Remove(path); err != nil {
				writeResponse(w, r, 500, fmt.Sprintf("Failed to delete: %s", err.Error()))
				return
			}

			// response
			writeResponse(w, r, 204, "")
			return

		} else {
			writeResponse(w, r, 405, "ERROR: Method not supported")
			return
		}


	})

	log.Printf("Listening on %s\n", os.Args[1])
	http.ListenAndServe(os.Args[1], nil)

}

func writeResponse(w http.ResponseWriter, r *http.Request, status int, content string) {
	writeResponseBytes(w, r, status, []byte(content))
}

func writeResponseBytes(w http.ResponseWriter, r *http.Request, status int, content []byte) {
	log.Printf("REQUEST %s \"%s %s %s\" %d %d\n", r.RemoteAddr, r.Method, r.URL, r.Proto, status, len(content))
	w.WriteHeader(status)
	w.Write(content)
}


