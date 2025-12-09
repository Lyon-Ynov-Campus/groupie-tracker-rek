package spotify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// pour stocker le token et credentials
var token string
var tokenExpire time.Time
var clientID string
var clientSecret string
var redirectURI = "http://localhost:8080/callback"

// TokenResponse c'est ce que spotify renvoie
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// SetCredentials pour stocker les credentials
func SetCredentials(id, secret string) {
	clientID = id
	clientSecret = secret
}

// GetAuthURL genere l'URL pour que l'utilisateur se connecte
func GetAuthURL() string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "user-read-private user-read-email")
	return "https://accounts.spotify.com/authorize?" + params.Encode()
}

// ExchangeCode echange le code contre un token
func ExchangeCode(code string) error {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("erreur spotify: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return err
	}

	//stocker le token
	token = tokenResp.AccessToken
	tokenExpire = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return nil
}

// GetAccessToken renvoie le token stocké
func GetAccessToken() string {
	return token
}

// IsAuthenticated verifie si on a un token valide
func IsAuthenticated() bool {
	return token != "" && time.Now().Before(tokenExpire)
}
