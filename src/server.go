package main

import (
	"fmt"
	"net/http"
	"os"
)

var (
	HTTP_ADDRESS = ":8080"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/pages/{page}", getPage)
	mux.HandleFunc("/", getIndex)

	err := http.ListenAndServe(HTTP_ADDRESS, mux)

	if err != nil {
		fmt.Printf("Error in server listen: %s\n", err)
	}

	fmt.Printf("Server closed\n")
	os.Exit(0);
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET INDEX\n")
	
	template := GetTemplate("index.pug")

	template.Execute(w, "No Args")
}

func getPage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	template := GetTemplate(path)
	err := template.Execute(w, struct { URL string } { path })
	
	fmt.Printf("GET: %s\n", path)

	if err != nil {
		fmt.Printf("Failed to realize tempalte: %s\n", err)
	}
}
