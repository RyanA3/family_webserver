package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	HTTP_ADDRESS = ":8080"
	PAGE_ROUTE = "/pages"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getIndex)
	mux.HandleFunc(PAGE_ROUTE, getPage)

	err := http.ListenAndServe(HTTP_ADDRESS, mux)

	if err != nil {
		fmt.Printf("Error in server listen: %s\n", err)
	}

	fmt.Printf("Server closed\n")
	os.Exit(0);
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET INDEX\n")
	io.WriteString(w, "Response")
}

func getPage(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET PAGE\n");
	
	path := r.URL.Path[len(PAGE_ROUTE):]

	template := GetTemplate(path)
	
	err := template.Execute(w, "No Args")

	if err != nil {
		fmt.Printf("Failed to realize tempalte: %s\n", err)
	}
}
