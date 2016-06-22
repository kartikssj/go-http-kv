
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
	"flag"
	"strings"
)

func main() {

	log.SetFlags(log.LstdFlags)

	var mode string
	var listen int
	var index string
	var root string
	flag.StringVar(&mode, "mode", "ws", "Mode {ws|kv}")
	flag.IntVar(&listen, "listen", 8080, "Listen port")
	flag.StringVar(&index, "index", "index.html", "Index file")
	flag.StringVar(&root, "root", "", "Root directory")
	flag.Parse()

	if root == "" {
		if c, err := os.Getwd(); err != nil {
			log.Fatalf("Failed to get CWD: %s\n", err.Error())
		} else {
			root = c
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		var path string = filepath.Clean(root + r.URL.Path)
		if mode == "ws" && (r.URL.Path == "" || strings.HasSuffix(r.URL.Path, "/")) {
			path = root + "/" + index
		}
		if !filepath.HasPrefix(path, root) {
			writeResponse(w, r, 400, "ERROR: Invalid path")
			return
		}
		if matched, _ := regexp.MatchString("[a-zA-Z0-9_\\-\\./]+", path); !matched {
			writeResponse(w, r, 400, "ERROR: Invalid path")
			return
		}

		if r.Method == "GET" {

			// check file exists
			stat, err := os.Stat(path)
			if os.IsNotExist(err) {
				writeResponse(w, r, 404, "Not found")
				return
			}

			// read file
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				writeResponse(w, r, 500, fmt.Sprintf("Failed to read: %s", err.Error()))
				return
			}

			writeHeaders(w, mode, path, stat)
			writeResponseBytes(w, r, 200, bytes)
			return

		} else if r.Method == "HEAD" {

			// check file exists
			stat, err := os.Stat(path)
			if os.IsNotExist(err) {
				writeResponse(w, r, 404, "Not found")
				return
			}

			writeHeaders(w, mode, path, stat)
			writeResponse(w, r, 200, "")
			return

		} else if mode == "kv" && r.Method == "PUT" {

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
			return

		} else if mode == "kv" && r.Method == "DELETE" {

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

	log.Printf("Listening on %d\n", listen)
	http.ListenAndServe(fmt.Sprintf(":%d", listen), nil)

}

func writeHeaders(w http.ResponseWriter, mode string, path string, stat os.FileInfo) {
	if mode == "ws" {
		mtype := mime.TypeByExtension(filepath.Ext(path))
		if mtype != "" {
			w.Header().Add("Content-Type", mtype)
		}
	} else {
		w.Header().Add("X-Name", stat.Name())
		w.Header().Add("X-Size", fmt.Sprintf("%d", stat.Size()))
		w.Header().Add("X-Modified", stat.ModTime().String())
	}
}

func writeResponse(w http.ResponseWriter, r *http.Request, status int, content string) {
	writeResponseBytes(w, r, status, []byte(content))
}

func writeResponseBytes(w http.ResponseWriter, r *http.Request, status int, content []byte) {
	log.Printf("REQUEST %s \"%s %s %s\" %d %d\n", r.RemoteAddr, r.Method, r.URL, r.Proto, status, len(content))
	if len(content) > 0 && r.Method != "HEAD" {
		w.Header().Add("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(status)
		w.Write(content)
	} else {
		w.Header().Add("Content-Length", "0")
		w.WriteHeader(status)
	}
}


