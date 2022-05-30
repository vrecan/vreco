# Overview
I've recently been very interested in getting back to the basics with server side rendering instead of building everything as a SPA (Single Page App). I stumbled upon [HotWire](https://hotwired.dev/) & [HTMX](https://htmx.org/)

I decided to build out a simple version of a live broadcast chat app with no state leveraging golang and HTMX. [Live Chat here](/live_chat). Anyone on the page at the same time will see messages sent over SSE (Server Side Events). To test it out, open the page in multiple tabs or devices and you should be able to communicate through the live chat page.


# How does it work? 

Leveraging go echo web server with HTML templates. Lets start with the go server. All code is [available here.](https://github.com/vrecan/vreco)

```go
	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
		LogLevel:  log.ERROR,
	}))

	e.HTTPErrorHandler = customHTTPErrorHandler
	err := routes.Setup(e)
	if err != nil {
		panic(fmt.Sprintln("failed to setup routes: ", err))
	}

	go func() {
		err := e.Start(":8080")
		if err != nil {
			e.Logger.Warn(err)
		}
	}()
```

You will notice that we override the http error handler, that allows us to get more useful errors.

```go
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	host := c.Request().Host
	URI := c.Request().RequestURI
	qs := c.QueryString()

	c.Logger().Error(err, fmt.Sprintf(" on: %s%s%s error code: %d", host, URI, qs, code))
	if code == 404 {
		c.Redirect(http.StatusTemporaryRedirect, "/404")
	}
	c.String(code, fmt.Sprintf("error code: %d", code))
}
```

Our setup function looks something like this.

```go	
	bc = broadcast.NewBroadcast()

	templates["live_chat.html"] = template.Must(template.New("").Funcs(functionMap).ParseFiles(
		"templates/pages/live_chat.html",
		"templates/base.html",
		"templates/partials/chat_input.html"))

	root.GET("live_chat", func(c echo.Context) error {
		return c.Render(http.StatusOK, "live_chat.html", map[string]interface{}{})
	})
```

We leverage a base html template to allow for simple navigation, header and footer of all our content.
You may have noticed the New Broadcast... we need a simple mechanism to broadcast to all listeners, the [implementation is here](https://github.com/vrecan/vreco/tree/main/broadcast)

The live_chat.html file looks like this: 
```html
{{define "title"}}
Live Chat
{{end}}

{{define "body"}}
<div hx-sse="connect:/chatroom" class="card text-center border bg-base-100
  shadow-xl p-8">
  Chatroom is open for business....
  <div hx-sse="swap:message" hx-swap="beforeend" class="card-body"> </div>
</div>
<div>
  <label class="block text-sm font-bold mb-2" for="username">
    Send a message
  </label>
  <div id="sendmsg" class="">
    {{template "chat_input.html" .}}
  </div>
</div>
{{end}}
```

Leveraging HTMX we are able to create an SSE connection with this simple attribute. `hx-sse="connect:/chatroom"` creates our connection while `hx-sse="swap:message" hx-swap="beforeend` tells the div component to swap every time a message event occurs. The beforeend swap tells it to append it's contents instead of replacing it.

```html
{{define "chat_msg.html"}}
data:  <p class="text-left border-dashed border-2 p-1">{{.msg}}</p>
{{end}}
```
The chat message content is required to be formatted with `data: `, this is part of the SSE standard, it also must be followed by `\n\n`.


Now that we have the HTML partials we also need to setup our routes.

```go
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
```
# Conclusion

HTMX is pretty powerful, it was very easy to use but there are clear limitations where you will still need to pull in javascript to create the user experience you really want. It does simplify a lot of things that would otherwise be difficult or hard to follow and works great for simple websites.



