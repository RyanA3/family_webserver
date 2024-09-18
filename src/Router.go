package main

import (
	"fmt"
	"github.com/Joker/jade"
	"html/template"
	"os"
)

const TEMPLATE_404_KEY string = "404"
const TEMPLATE_404_CONTENT string = "<p>404 Not Found</p>"
var TEMPLATES map[string]*template.Template = make(map[string]*template.Template)

func loadTemplate(path string) (*template.Template, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		fmt.Printf("Failed to read file: %s\n", path)
		return nil, err
	}

	jadeTemplate, err := jade.Parse("jade", data)

	if err != nil {
		fmt.Printf("Failed to parse jade template: %s\n", err)
		return nil, err
	}

	goTemplate, err := template.New(path).Parse(jadeTemplate)

	if err != nil {
		fmt.Printf("Failed to convert jade template to go template: %s\n", err);
		return nil, err
	}
	
	TEMPLATES[path] = goTemplate
	return goTemplate, nil
}

func GetNotFoundTemplate() *template.Template {
	templ := TEMPLATES[TEMPLATE_404_KEY]

	if templ != nil {
		return templ
	}

	templ, err := template.New(TEMPLATE_404_KEY).Parse(TEMPLATE_404_CONTENT)

	if err != nil {
		fmt.Printf("Failed to create 404 template: %s\n", err);
	}

	TEMPLATES[TEMPLATE_404_KEY] = templ
	return templ
}

func GetTemplate(path string) *template.Template {
	templ := TEMPLATES[path]

	if templ != nil {
		return templ
	}
	
	templ, err := loadTemplate(path)

	if err != nil {
		fmt.Printf("Failed to get template %s\n", path)
		return GetNotFoundTemplate()
	}

	return templ
}

