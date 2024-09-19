package main

import (
	"fmt"
	"github.com/Joker/jade"
	"html/template"
	"os"
	"path"
)

const VIEWS_PATH string = "./views/"
const PAGES_PATH string = "/pages/"
const COMPONENTS_PATH string = "/components/"
const PAGE_404_PATH = "/pages/404.pug"

const TEMPLATE_404_KEY string = "404"
const TEMPLATE_404_CONTENT string = "<p>404 Not Found</p>"
var TEMPLATES map[string]*template.Template = make(map[string]*template.Template)

func loadTemplate(key string) (*template.Template, error) {
	filepath := path.Join(VIEWS_PATH, key)
	data, err := os.ReadFile(filepath)

	if err != nil {
		fmt.Printf("Failed to read file %s: %s\n", filepath, err)
		return nil, err
	}

	jadeTemplate, err := jade.Parse("jade", data)

	if err != nil {
		fmt.Printf("Failed to parse jade template: %s\n", err)
		return nil, err
	}

	goTemplate, err := template.New(key).Parse(jadeTemplate)

	if err != nil {
		fmt.Printf("Failed to convert jade template to go template: %s\n", err);
		return nil, err
	}
	
	TEMPLATES[key] = goTemplate
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

func GetTemplate(key string) *template.Template {
	templ := TEMPLATES[key]

	if templ != nil {
		return templ
	}
	
	templ, err := loadTemplate(key)

	if err != nil {
		fmt.Printf("Failed to get template %s\n", key)
		return nil
	}

	return templ
}

func GetPage(key string) *template.Template {
	filepath := path.Join(PAGES_PATH, key)
	templ := GetTemplate(filepath)

	if templ == nil {
		fmt.Printf("Failed to get page %s\n", key)
		return GetTemplate(PAGE_404_PATH)
	}

	return templ
}

func GetComponent(key string) *template.Template {
	filepath := path.Join(COMPONENTS_PATH, key)
	templ := GetTemplate(filepath)

	if templ == nil {
		fmt.Printf("Failed to get component %s\n", key)
		return GetNotFoundTemplate()
	}

	return templ
}
