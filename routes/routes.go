package routes

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/labstack/echo/v4"
)

var messageChan chan string

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
	if messageChan == nil {
		messageChan = make(chan string)
	}
	templates := make(map[string]*template.Template)
	templates["home.html"] = template.Must(template.ParseFiles(
		"templates/pages/home.html",
		"templates/base.html",
		"templates/partials/chat_input.html"))
	templates["about.html"] = template.Must(template.ParseFiles("templates/pages/about.html", "templates/base.html"))
	templates["clicked.html"] = template.Must(template.ParseFiles("templates/partials/clicked.html"))
	templates["chat_msg.html"] = template.Must(template.ParseFiles("templates/partials/chat_msg.html"))
	templates["chat_input.html"] = template.Must(template.ParseFiles("templates/partials/chat_input.html"))

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

		if messageChan != nil && msg != "" {

			// send the message through the available channel
			select {
			case messageChan <- msg:
				log.Println("message sent to channel: ", msg)
			default:
				log.Println("no one listening for messages, not sending: ", msg)
			}
		}
		return c.Render(http.StatusOK, "chat_input.html", map[string]interface{}{})
	})

}

func handleSSE(c echo.Context, t echo.Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Get handshake from client")
		// prepare the header
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// instantiate the channel

		// prepare the flusher
		flusher, _ := w.(http.Flusher)

		// trap the request under loop forever
		for {

			select {

			// message will received here and printed
			case msg := <-messageChan:
				fmt.Println("Sending message through chatroom", msg)
				t.Render(c.Response().Writer, "chat_msg.html", map[string]interface{}{
					"msg": msg,
				}, c)
				fmt.Fprintf(w, "\n\n")
				flusher.Flush()

			// connection is closed then defer will be executed
			case <-r.Context().Done():
				log.Println("Context is done exiting")
				return

			}
		}

	}
}
