package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/suer/first-task-cli/internal/config"
)

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
	signInURL      = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithIdp"
	refreshURL     = "https://securetoken.googleapis.com/v1/token"
	expiryMargin   = time.Minute
)

// Login runs the Google OAuth loopback flow and exchanges the result for a Firebase session.
func Login(ctx context.Context, cfg config.Config, openURL func(string)) (config.Credentials, error) {
	verifier, err := randomString(32)
	if err != nil {
		return config.Credentials{}, err
	}
	state, err := randomString(16)
	if err != nil {
		return config.Credentials{}, err
	}
	challenge := sha256.Sum256([]byte(verifier))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return config.Credentials{}, err
	}
	defer func() { _ = listener.Close() }()
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port)

	type result struct {
		code string
		err  error
	}
	resultCh := make(chan result, 1)
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch {
		case q.Get("error") != "":
			resultCh <- result{err: fmt.Errorf("authorization failed: %s", q.Get("error"))}
		case q.Get("state") != state:
			resultCh <- result{err: errors.New("state mismatch")}
		case q.Get("code") == "":
			http.NotFound(w, r)
			return
		default:
			resultCh <- result{code: q.Get("code")}
		}
		_, _ = fmt.Fprintln(w, "Login complete. You can close this tab.")
	})}
	go func() { _ = server.Serve(listener) }()
	defer func() { _ = server.Shutdown(context.Background()) }()

	authURL := googleAuthURL + "?" + url.Values{
		"client_id":             {cfg.ClientID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"scope":                 {"openid email profile"},
		"state":                 {state},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(challenge[:])},
		"code_challenge_method": {"S256"},
	}.Encode()
	openURL(authURL)

	var res result
	select {
	case res = <-resultCh:
	case <-ctx.Done():
		return config.Credentials{}, ctx.Err()
	case <-time.After(5 * time.Minute):
		return config.Credentials{}, errors.New("login timed out")
	}
	if res.err != nil {
		return config.Credentials{}, res.err
	}

	var tokenResp struct {
		IDToken string `json:"id_token"`
	}
	if err := postForm(ctx, googleTokenURL, url.Values{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"code":          {res.code},
		"code_verifier": {verifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}, &tokenResp); err != nil {
		return config.Credentials{}, fmt.Errorf("exchange code: %w", err)
	}

	var signIn struct {
		IDToken      string `json:"idToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresIn    string `json:"expiresIn"`
		LocalID      string `json:"localId"`
		Email        string `json:"email"`
	}
	body := map[string]any{
		"postBody":            "id_token=" + url.QueryEscape(tokenResp.IDToken) + "&providerId=google.com",
		"requestUri":          "http://localhost",
		"returnIdpCredential": true,
		"returnSecureToken":   true,
	}
	if err := postJSON(ctx, signInURL+"?key="+url.QueryEscape(cfg.APIKey), body, &signIn); err != nil {
		return config.Credentials{}, fmt.Errorf("sign in to Firebase: %w", err)
	}

	cred := config.Credentials{
		UID:          signIn.LocalID,
		Email:        signIn.Email,
		IDToken:      signIn.IDToken,
		RefreshToken: signIn.RefreshToken,
		ExpiresAt:    expiry(signIn.ExpiresIn),
	}
	return cred, config.SaveCredentials(cred)
}

// Session returns valid credentials, refreshing the ID token when it is about to expire.
func Session(ctx context.Context, cfg config.Config) (config.Credentials, error) {
	cred, err := config.LoadCredentials()
	if err != nil {
		return cred, err
	}
	if time.Now().Add(expiryMargin).Before(cred.ExpiresAt) {
		return cred, nil
	}

	var resp struct {
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    string `json:"expires_in"`
		UserID       string `json:"user_id"`
	}
	if err := postForm(ctx, refreshURL+"?key="+url.QueryEscape(cfg.APIKey), url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {cred.RefreshToken},
	}, &resp); err != nil {
		return cred, fmt.Errorf("refresh token (try `firsttask login`): %w", err)
	}
	cred.IDToken = resp.IDToken
	cred.RefreshToken = resp.RefreshToken
	cred.ExpiresAt = expiry(resp.ExpiresIn)
	return cred, config.SaveCredentials(cred)
}

func expiry(expiresIn string) time.Time {
	sec, err := strconv.Atoi(expiresIn)
	if err != nil {
		sec = 3600
	}
	return time.Now().Add(time.Duration(sec) * time.Second)
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func postForm(ctx context.Context, endpoint string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return do(req, out)
}

func postJSON(ctx context.Context, endpoint string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return do(req, out)
}

func do(req *http.Request, out any) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	return json.Unmarshal(data, out)
}
