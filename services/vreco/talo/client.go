package talo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	apiURL = "https://api.trytalo.com"
)

// Client calls the Talo Player Auth API.
type Client struct {
	baseURL   string
	accessKey string
	client    *http.Client
}

// NewClient creates a Talo client. accessKey is read from TALO_ACCESS_KEY env if empty.
func NewClient(accessKey string) *Client {
	if accessKey == "" {
		accessKey = os.Getenv("TALO_ACCESS_KEY")
	}
	return &Client{
		baseURL:   apiURL,
		accessKey: accessKey,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// HasAccessKey returns true if the client has an access key configured.
func (c *Client) HasAccessKey() bool {
	return c.accessKey != ""
}

// LoginRequest is the request body for login.
type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// LoginResponse is the response from a successful login.
type LoginResponse struct {
	Alias               Alias   `json:"alias"`
	SessionToken        string  `json:"sessionToken"`
	SocketToken         string  `json:"socketToken"`
	VerificationRequired bool    `json:"verificationRequired,omitempty"`
	AliasID             int     `json:"aliasId,omitempty"`
}

// Alias represents a player alias from Talo.
type Alias struct {
	ID         int    `json:"id"`
	Service    string `json:"service"`
	Identifier string `json:"identifier"`
	Player     Player `json:"player"`
}

// Player represents a player from Talo.
type Player struct {
	ID string `json:"id"`
}

// LoginResult represents the outcome of a login attempt.
type LoginResult struct {
	OK                   bool
	VerificationRequired bool
	Alias                *Alias
	SessionToken         string
	ErrorCode            string
	Message              string
}

// Login authenticates a player with Talo.
func (c *Client) Login(identifier, password string) (*LoginResult, error) {
	body, err := json.Marshal(LoginRequest{Identifier: identifier, Password: password})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/v1/players/auth/login", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	c.setCommonHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		var loginResp LoginResponse
		if err := json.Unmarshal(respBody, &loginResp); err != nil {
			return nil, fmt.Errorf("parse login response: %w", err)
		}
		return &LoginResult{
			OK:                   !loginResp.VerificationRequired,
			VerificationRequired: loginResp.VerificationRequired,
			Alias:                &loginResp.Alias,
			SessionToken:         loginResp.SessionToken,
		}, nil
	}

	var errBody struct {
		ErrorCode string `json:"errorCode"`
		Message   string `json:"message"`
	}
	_ = json.Unmarshal(respBody, &errBody)

	return &LoginResult{
		OK:        false,
		ErrorCode: errBody.ErrorCode,
		Message:   errBody.Message,
	}, nil
}

// DeleteAccount deletes the player account. Requires session headers.
func (c *Client) DeleteAccount(sessionToken string, aliasID int, playerID, currentPassword string) error {
	body, err := json.Marshal(map[string]string{"currentPassword": currentPassword})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, c.baseURL+"/v1/players/auth/", bytes.NewReader(body))
	if err != nil {
		return err
	}

	c.setCommonHeaders(req)
	req.Header.Set("x-talo-session", sessionToken)
	req.Header.Set("x-talo-alias", fmt.Sprintf("%d", aliasID))
	req.Header.Set("x-talo-player", playerID)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	var errBody struct {
		ErrorCode string `json:"errorCode"`
		Message   string `json:"message"`
	}
	_ = json.Unmarshal(respBody, &errBody)
	return fmt.Errorf("talo delete account: %s (code: %s)", errBody.Message, errBody.ErrorCode)
}

func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.accessKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}
