package services

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Custom template funcs
var templateFuncs = template.FuncMap{
	"ucfirst":    ucfirst,
	"now":        time.Now,
	"formatDate": formatDate,
}

func Render(c *fiber.Ctx, userRole string, relativePath string, data map[string]interface{}) error {
	if userRole == "examiner" {
		userRole = "student"
	}
	if userRole == "admin" {
		userRole = "admin" // оставляем admin как есть
	}

	tplPath := filepath.Join("views", "pages", userRole, relativePath)

	if _, err := os.Stat(tplPath); os.IsNotExist(err) {
		return c.Status(404).SendString("Template not found: " + tplPath)
	}

	tmpl := template.New(filepath.Base(tplPath)).Funcs(template.FuncMap{
		"add1":  func(i int) int { return i + 1 },
		"slice": func(args ...int) []int { return args },
	})

	tmpl, err := tmpl.ParseFiles(tplPath)
	if err != nil {
		return c.Status(500).SendString("Template parse error: " + err.Error())
	}

	var rendered strings.Builder
	if err := tmpl.Execute(&rendered, data); err != nil {
		return c.Status(500).SendString("Template exec error: " + err.Error())
	}

	return c.Type("html").SendString(rendered.String())
}

// Example custom template functions
func ucfirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func formatDate(t time.Time, layout string) string {
	return t.Format(layout)
}
