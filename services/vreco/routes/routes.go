package routes

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
	"vreco/broadcast"
	vMiddleware "vreco/routes/middleware"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/russross/blackfriday/v2"
)

var bc *broadcast.BroadCast

// Define the template registry struct
type TemplateRegistry struct {
	templates map[string]*template.Template
}

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	template, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	// if we are loading a partial base will be missing
	base := template.Lookup("base.html")
	if base == nil {
		return template.ExecuteTemplate(w, name, data)
	}
	return template.ExecuteTemplate(w, "base.html", data)

}

func markDowner(args ...interface{}) template.HTML {
	s := blackfriday.Run([]byte(fmt.Sprintf("%s", args...)))
	fmt.Println(string(s))
	return template.HTML(s)
}

func Setup(e *echo.Echo) error {
	if bc == nil {
		bc = broadcast.NewBroadcast()
	}
	SetupStaticAssets(e)

	blogs, err := GenerateBlogHtml("posts/")
	if err != nil {
		return err
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
	templates["live_chat.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles(
		"templates/pages/live_chat.html",
		"templates/base.html",
		"templates/partials/chat_input.html"))
	templates["blog.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles(
		"templates/pages/blog.html",
		"templates/base.html"))
	templates["post.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles(
		"templates/pages/post.html",
		"templates/base.html"))
	templates["blog_card.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles(
		"templates/partials/blog_card.html"))
	templates["about.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles("templates/pages/about.html", "templates/base.html"))
	templates["clicked.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles("templates/partials/clicked.html"))
	templates["chat_msg.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles("templates/partials/chat_msg.html"))
	templates["chat_input.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles("templates/partials/chat_input.html"))

	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

	root := e.Group("/", vMiddleware.CacheControl(0))

	root.GET("health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	root.GET("", func(c echo.Context) error {
		return c.Render(http.StatusOK, "home.html", map[string]interface{}{})
	})
	root.GET("blog", func(c echo.Context) error {
		params := make(map[string]string, 0)
		for _, k := range c.ParamNames() {
			for _, value := range c.ParamValues() {
				params[k] = value
			}
		}
		PID, exist := params["id"]
		var ID int
		if !exist {
			ID = 0
		} else {
			ID, err = strconv.Atoi(PID)
		}
		if err != nil {
			return err
		}
		blog, err := getBlogByID(ID, blogs)
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, "blog.html", map[string]interface{}{
			"blog": blog,
		})
	})
	root.GET("blog/post/:name", func(c echo.Context) error {
		var name string
		for range c.ParamNames() {
			for _, value := range c.ParamValues() {
				name = value
			}
		}
		blog, err := getBlogByName(name, blogs)
		if err != nil {
			return c.Render(http.StatusOK, "404.html", map[string]interface{}{})
		}
		return c.Render(http.StatusOK, "post.html", map[string]interface{}{
			"blog": blog,
			"name": name,
		})

	})

	root.GET("blog/card", func(c echo.Context) error {
		QID := c.QueryParam("id")
		var ID int
		nextID := new(int)
		if QID == "" {
			return fmt.Errorf("invalid card id")
		}
		ID, err = strconv.Atoi(QID)
		if err != nil {
			return err
		}
		blog, err := getBlogByID(ID, blogs)
		if err != nil {
			return err
		}

		if ID+1 <= len(blogs)-1 {
			*nextID = ID + 1
			return c.Render(http.StatusOK, "blog_card.html", map[string]interface{}{
				"blog":   blog,
				"nextID": nextID,
			})
		}
		//invalid nextID so don't render it
		return c.Render(http.StatusOK, "blog_card.html", map[string]interface{}{
			"blog": blog,
		})

	})
	root.GET("live_chat", func(c echo.Context) error {
		return c.Render(http.StatusOK, "live_chat.html", map[string]interface{}{})
	})
	root.GET("404", func(c echo.Context) error {
		return c.Render(http.StatusOK, "404.html", map[string]interface{}{})
	})
	root.GET("about", func(c echo.Context) error {
		return c.Render(http.StatusOK, "about.html", map[string]interface{}{})
	})
	root.POST("clicked", func(c echo.Context) error {
		return c.Render(http.StatusOK, "clicked.html", map[string]interface{}{})
	})

	root.GET("chatroom", func(c echo.Context) error {
		handler := handleSSE(c, e.Renderer)
		handler(c.Response().Writer, c.Request())
		return nil
	})

	e.POST("sendChat", func(c echo.Context) error {
		msg := c.FormValue("msg")

		if bc != nil && msg != "" {
			errs := bc.Send(msg)
			for id, err := range errs {
				e.Logger.Errorf("listener: %s %s", id, err)
			}
		}
		return c.Render(http.StatusOK, "chat_input.html", map[string]interface{}{})
	})
	return nil
}

func SetupStaticAssets(e *echo.Echo) {
	e.Use(vMiddleware.CacheControl(0), middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "static",
		Browse: false,
	}))
}

func getBlogByName(name string, blogs Blogs) (blog *Blog, err error) {
	for _, b := range blogs {
		if b.Meta.Title == name {
			return &b, nil
		}
	}
	return nil, fmt.Errorf("no blog found")
}

func getBlogByID(id int, blogs Blogs) (blog *Blog, err error) {
	for i, b := range blogs {
		if id == i {
			return &b, nil
		}
	}
	return nil, fmt.Errorf("no blog found")
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

func GenerateBlogHtml(relativePath string) (blogs Blogs, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return blogs, err
	}
	blogs = make([]Blog, 0)
	path := filepath.Join(cwd, relativePath)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return blogs, err
	}

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			blog, err := readBlogFolder(filepath.Join(path, fileInfo.Name()))
			if err != nil {
				return blogs, err
			}
			blogs = append(blogs, blog)
			continue
		}
	}
	sort.Sort(sort.Reverse(blogs))
	return blogs, err

}

type BlogMeta struct {
	Categories  []string
	Description string
	Tags        []string
	Date        time.Time
	Title       string
}

type Blog struct {
	Meta     BlogMeta
	Contents []byte
}

type Blogs []Blog

func (b Blogs) Len() int {
	return len(b)
}

func (b Blogs) Less(i, j int) bool {
	return b[i].Meta.Date.Before(b[j].Meta.Date)
}

func (b Blogs) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func readBlogFolder(path string) (blog Blog, err error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return blog, err
	}

	for _, fileInfo := range files {
		if fileInfo.Name() == "index.md" {
			contents, err := readFileWithLimit(filepath.Join(path, fileInfo.Name()), 5242880)
			if err != nil {
				return blog, err
			}
			blog.Contents = contents
		}
		if fileInfo.Name() == "meta.toml" {
			meta := &BlogMeta{}
			contents, err := readFileWithLimit(filepath.Join(path, fileInfo.Name()), 5242880)
			if err != nil {
				return blog, err
			}
			_, err = toml.Decode(string(contents), meta)
			if err != nil {
				return blog, err
			}
			blog.Meta = *meta

		}
	}
	return blog, err

}

func readFileWithLimit(path string, limit int64) (contents []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(io.LimitReader(file, limit)) //max size 5mb
	return bytes, err
}
