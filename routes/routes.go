package routes

import (
	"errors"
	"io"
	"net/http"

	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/labstack/echo/v4"
)

// Define the template registry struct
type TemplateRegistry struct {
	templates map[string]*template.Template
}

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	// if we are loading a partial base will be missing
	base := tmpl.Lookup("base.html")
	if base == nil {
		return tmpl.ExecuteTemplate(w, name, data)
	}
	return tmpl.ExecuteTemplate(w, "base.html", data)

}

func Setup(e *echo.Echo) {
	templates := make(map[string]*template.Template)
	templates["home.html"] = template.Must(template.ParseFiles("templates/pages/home.html", "templates/base.html"))
	templates["clicked.html"] = template.Must(template.ParseFiles("templates/components/clicked.html"))
	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

	// set funcmap for all templates
	for _, template := range templates {
		template.Funcs(sprig.FuncMap())
	}

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "home.html", map[string]interface{}{})
	})
	e.POST("/clicked", func(c echo.Context) error {
		return c.Render(http.StatusOK, "clicked.html", map[string]interface{}{})
	})
}
