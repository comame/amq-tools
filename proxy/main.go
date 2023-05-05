package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/comame/readenv-go"
)

type env struct {
	FileDir string `env:"FILEDIR"`
}

var CATBOX_HOST = "https://files.catbox.moe"

func main() {
	var env env
	readenv.Read(&env)

	fileroot, err := filepath.Abs(env.FileDir)
	if err != nil {
		panic(err)
	}

	fmt.Println(fileroot)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, file := path.Split(r.URL.Path)
		if len(file) < 2 {
			w.WriteHeader(404)
			return
		}

		fmt.Println(file)

		f, err := readFromCache(fileroot, file)
		if err != nil {
			fmt.Println("  Request origin")
			proxyOrigin(w, r)
			return
		}
		defer f.Close()

		fmt.Println(" Cached")
		io.Copy(w, f)
	})

	fmt.Println("Server started localhost:54321")
	http.ListenAndServe(":54321", nil)
}

func readFromCache(fileroot, filename string) (*os.File, error) {
	f, err := os.Open(calculateFilePath(fileroot, filename))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func proxyOrigin(w http.ResponseWriter, r *http.Request) {
	rp := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			newUrl, _ := url.Parse(CATBOX_HOST)
			pr.SetURL(newUrl)
		},
	}
	rp.ServeHTTP(w, r)
}

func calculateFilePath(base, filename string) string {
	dirname := filename[:2]
	return path.Join(base, "file", dirname, filename)
}
