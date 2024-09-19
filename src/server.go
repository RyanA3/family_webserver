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
	mux.HandleFunc("/components/{component}", getComponent)
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
	
	templ := GetTemplate("index.pug")

	templ.Execute(w, "No Args")
}

func getPage(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	templ := GetPage(page)
	err := templ.Execute(w, struct { URL string } { r.URL.Path })
	
	fmt.Printf("GET: %s\n", page)

	if err != nil {
		fmt.Printf("Failed to realize page template: %s\n", err)
	}
}

func getComponent(w http.ResponseWriter, r *http.Request) {
	component := r.PathValue("component")
	templ := GetComponent(component)
	err := templ.Execute(w, 
		struct { URL string; Query map[string][]string } { r.URL.Path, r.URL.Query() })

	fmt.Printf("GETL %s\n", component)

	if err != nil {
		fmt.Printf("Failed to realize component template: %s\n", err)
	}
}
