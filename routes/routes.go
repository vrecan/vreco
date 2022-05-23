package routes

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	"vreco/broadcast"

	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/labstack/echo/v4"
	"github.com/russross/blackfriday/v2"
)

var bc *broadcast.BroadCast

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

func markDowner(args ...interface{}) template.HTML {
	fmt.Println("Calling markdown!: ", args)
	s := blackfriday.Run([]byte(fmt.Sprintf("%s", args...)))
	return template.HTML(s)
}

func Setup(e *echo.Echo) {
	if bc == nil {
		bc = broadcast.NewBroadcast()
	}

	functionMap := template.FuncMap{
		"markdown": markDowner,
	}
	for k, v := range sprig.FuncMap() {
		functionMap[k] = v
	}

	templates := make(map[string]*template.Template)
	templates["home.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles(
		"templates/pages/home.html",
		"templates/base.html"))
	templates["404.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles(
		"templates/pages/404.html",
		"templates/base.html"))
	templates["blog.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles(
		"templates/pages/blog.html",
		"templates/base.html",
		"templates/partials/chat_input.html"))
	templates["about.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles("templates/pages/about.html", "templates/base.html"))
	templates["clicked.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles("templates/partials/clicked.html"))
	templates["chat_msg.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles("templates/partials/chat_msg.html"))
	templates["chat_input.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles("templates/partials/chat_input.html"))

	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "home.html", map[string]interface{}{})
	})
	e.GET("/blog", func(c echo.Context) error {
		return c.Render(http.StatusOK, "blog.html", map[string]interface{}{})
	})
	e.GET("/404", func(c echo.Context) error {
		return c.Render(http.StatusOK, "404.html", map[string]interface{}{})
	})
	e.GET("/about", func(c echo.Context) error {
		return c.Render(http.StatusOK, "about.html", map[string]interface{}{})
	})
	e.POST("/clicked", func(c echo.Context) error {
		return c.Render(http.StatusOK, "clicked.html", map[string]interface{}{})
	})

	e.GET("/chatroom", func(c echo.Context) error {
		handler := handleSSE(c, e.Renderer)
		handler(c.Response().Writer, c.Request())
		return nil
	})

	e.POST("/sendChat", func(c echo.Context) error {
		msg := c.FormValue("msg")

		if bc != nil && msg != "" {
			errs := bc.Send(msg)
			for id, err := range errs {
				e.Logger.Errorf("listener: %s %s", id, err)
			}
		}
		return c.Render(http.StatusOK, "chat_input.html", map[string]interface{}{})
	})

}

func handleSSE(c echo.Context, t echo.Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// prepare the header
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, _ := w.(http.Flusher)

		list := bc.AddListener()
		defer bc.RemoveListener(list)

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {

			select {

			case msg := <-list.Chan:
				t.Render(w, "chat_msg.html", map[string]interface{}{
					"msg": msg,
				}, c)
				fmt.Fprintf(w, "\n\n")
				flusher.Flush()
			case <-ticker.C:
				fmt.Fprintf(w, "keepalive: \n\n")
				flusher.Flush()
			case <-r.Context().Done():
				return

			}
		}

	}
}
