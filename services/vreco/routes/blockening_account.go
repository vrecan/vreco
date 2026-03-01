package routes

import (
	"net/http"
	"os"

	"vreco/talo"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

const (
	sessionName          = "blockening_account"
	sessionKeyTalo       = "talo_session"
	sessionKeyAliasID   = "talo_alias_id"
	sessionKeyPlayerID  = "talo_player_id"
	sessionKeyIdentifier = "talo_identifier"

	maxIdentifierLen = 256
	maxPasswordLen   = 256
)

// SetupBlockeningAccount registers the Blockening account management routes.
// Requires TALO_ACCESS_KEY and SESSION_SECRET env vars. If either is missing,
// registers a fallback route that shows "not configured" so the URL always works.
func SetupBlockeningAccount(e *echo.Echo) error {
	accessKey := os.Getenv("TALO_ACCESS_KEY")
	sessionSecret := os.Getenv("SESSION_SECRET")

	accessKeySet := accessKey != ""
	sessionSecretSet := sessionSecret != ""
	sessionSecretLenOk := len(sessionSecret) >= 32

	e.Logger.Infof("[blockening_account] TALO_ACCESS_KEY set=%v, SESSION_SECRET set=%v len=%d (need>=32)", accessKeySet, sessionSecretSet, len(sessionSecret))

	if !accessKeySet || !sessionSecretSet || !sessionSecretLenOk {
		e.Logger.Warnf("[blockening_account] Using fallback route (not configured): accessKey=%v sessionSecret=%v sessionSecretLenOk=%v", accessKeySet, sessionSecretSet, sessionSecretLenOk)
		e.GET("/games/blockening/account", handleAccountNotConfigured)
		return nil
	}

	e.Logger.Info("[blockening_account] Registering full account routes")
	taloClient := talo.NewClient(accessKey)
	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = true
	store.Options.SameSite = http.SameSiteLaxMode
	store.MaxAge(24 * 60 * 60) // 24 hours

	account := e.Group("/games/blockening/account",
		session.Middleware(store),
		echomw.CSRFWithConfig(echomw.CSRFConfig{
			TokenLookup:    "form:csrf",
			CookieSecure:   true,
			CookieHTTPOnly: true,
			CookieSameSite: http.SameSiteLaxMode,
		}))
	account.GET("", func(c echo.Context) error {
		return handleAccountPage(c, taloClient)
	})
	account.POST("/login", func(c echo.Context) error {
		return handleLogin(c, taloClient)
	})
	account.POST("/logout", func(c echo.Context) error {
		return handleLogout(c)
	})
	account.POST("/delete", func(c echo.Context) error {
		return handleDelete(c, taloClient)
	})

	return nil
}

func handleAccountNotConfigured(c echo.Context) error {
	c.Logger().Info("[blockening_account] Serving not-configured page (env vars missing or invalid)")
	return c.Render(http.StatusOK, "blockening_account.html", map[string]interface{}{
		"LoggedIn":       false,
		"NotConfigured": true,
	})
}

func getSessionAliasID(sess *sessions.Session) int {
	if v, ok := sess.Values[sessionKeyAliasID].(int); ok {
		return v
	}
	if v, ok := sess.Values[sessionKeyAliasID].(float64); ok {
		return int(v)
	}
	return 0
}

func getSession(c echo.Context) (*sessions.Session, error) {
	sess, err := session.Get(sessionName, c)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func isLoggedIn(c echo.Context) bool {
	sess, err := getSession(c)
	if err != nil {
		return false
	}
	_, ok := sess.Values[sessionKeyTalo].(string)
	return ok
}

func handleAccountPage(c echo.Context, client *talo.Client) error {
	loggedIn := isLoggedIn(c)
	c.Logger().Infof("[blockening_account] Account page: loggedIn=%v", loggedIn)

	data := map[string]interface{}{
		"LoggedIn":   loggedIn,
		"Flash":      c.QueryParam("flash"),
		"FlashError": c.QueryParam("error"),
	}
	if csrf, ok := c.Get("csrf").(string); ok {
		data["CSRFToken"] = csrf
	}
	if data["LoggedIn"].(bool) {
		sess, _ := getSession(c)
		if sess != nil {
			data["Identifier"] = sess.Values[sessionKeyIdentifier]
		}
	}
	return c.Render(http.StatusOK, "blockening_account.html", data)
}

func handleLogin(c echo.Context, client *talo.Client) error {
	identifier := c.FormValue("identifier")
	password := c.FormValue("password")
	if identifier == "" || password == "" {
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=missing_credentials")
	}
	if len(identifier) > maxIdentifierLen || len(password) > maxPasswordLen {
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=invalid_credentials")
	}

	result, err := client.Login(identifier, password)
	if err != nil {
		c.Logger().Error(err)
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=login_failed")
	}

	if result.VerificationRequired {
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=verification_required")
	}

	if !result.OK || result.Alias == nil {
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=invalid_credentials")
	}

	sess, err := getSession(c)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=session_failed")
	}
	sess.Values[sessionKeyTalo] = result.SessionToken
	sess.Values[sessionKeyAliasID] = result.Alias.ID
	sess.Values[sessionKeyPlayerID] = result.Alias.Player.ID
	sess.Values[sessionKeyIdentifier] = result.Alias.Identifier
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Error(err)
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=session_save_failed")
	}

	return c.Redirect(http.StatusSeeOther, "/games/blockening/account")
}

func handleLogout(c echo.Context) error {
	sess, err := getSession(c)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account")
	}
	for k := range sess.Values {
		delete(sess.Values, k)
	}
	sess.Options.MaxAge = -1
	_ = sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusSeeOther, "/games/blockening/account")
}

func handleDelete(c echo.Context, client *talo.Client) error {
	if !isLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=not_logged_in")
	}

	currentPassword := c.FormValue("currentPassword")
	if currentPassword == "" {
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=password_required")
	}
	if len(currentPassword) > maxPasswordLen {
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=delete_failed")
	}

	sess, err := getSession(c)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=session_failed")
	}

	sessionToken, _ := sess.Values[sessionKeyTalo].(string)
	playerID, _ := sess.Values[sessionKeyPlayerID].(string)
	aliasID := getSessionAliasID(sess)

	if sessionToken == "" || aliasID == 0 || playerID == "" {
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=session_invalid")
	}

	err = client.DeleteAccount(sessionToken, aliasID, playerID, currentPassword)
	if err != nil {
		c.Logger().Error(err)
		return c.Redirect(http.StatusSeeOther, "/games/blockening/account?error=delete_failed")
	}

	for k := range sess.Values {
		delete(sess.Values, k)
	}
	sess.Options.MaxAge = -1
	_ = sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusSeeOther, "/games/blockening/account?flash=account_deleted")
}
