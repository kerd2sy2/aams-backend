package service

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ──────────────────────────────────────────────────────────────────────────────
// FCM V1 Push Notification Service (Firebase Cloud Messaging)
// Sends push notifications directly to Android devices via FCM V1 API,
// meaning they arrive even when the app is closed or the screen is locked.
// ──────────────────────────────────────────────────────────────────────────────

type fcmServiceAccount struct {
	ProjectID   string `json:"project_id"`
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
	TokenURI    string `json:"token_uri"`
}

type fcmSenderClient struct {
	sa fcmServiceAccount
}

// Global FCM client instance
var globalFCMClient *fcmSenderClient

// InitFCM loads the Firebase service account and initializes the FCM client.
// Searches for the JSON file in common locations if path is empty.
func InitFCM(serviceAccountPath string) {
	candidates := []string{serviceAccountPath}
	if serviceAccountPath == "" {
		candidates = []string{}
	}
	candidates = append(candidates,
		"aams-8a73f-firebase-adminsdk-fbsvc-aa4581dbbd.json",
		"../aams-8a73f-firebase-adminsdk-fbsvc-aa4581dbbd.json",
		"fcm-service-account.json",
		"../fcm-service-account.json",
	)

	for _, path := range candidates {
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var sa fcmServiceAccount
		if err := json.Unmarshal(data, &sa); err != nil {
			continue
		}
		if sa.ProjectID == "" || sa.PrivateKey == "" {
			continue
		}
		globalFCMClient = &fcmSenderClient{sa: sa}
		fmt.Printf("[FCM] Initialized with project: %s\n", sa.ProjectID)
		return
	}
	fmt.Println("[FCM] Service account not found — falling back to Expo Push API")
}

func ensureAbsoluteURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}
	if !strings.HasPrefix(rawURL, "/") {
		rawURL = "/" + rawURL
	}
	return "https://api.kerd2sy.com" + rawURL
}

// SendFCMBroadcast sends FCM push notifications to a list of FCM/Expo tokens in background.
// For Expo tokens (ExponentPushToken[...]), it falls back to Expo Push API.
// For native FCM tokens, it sends directly via FCM V1 API.
func SendFCMBroadcast(tokens []string, title, body string, extraData map[string]string) {
	if len(tokens) == 0 {
		return
	}

	if extraData != nil && extraData["image_url"] != "" {
		extraData["image_url"] = ensureAbsoluteURL(extraData["image_url"])
	}

	var fcmTokens []string
	var expoTokens []string

	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if strings.HasPrefix(tok, "ExponentPushToken[") || strings.HasPrefix(tok, "ExpoPushToken[") {
			expoTokens = append(expoTokens, tok)
		} else if len(tok) > 20 {
			// Native FCM token
			fcmTokens = append(fcmTokens, tok)
		}
	}

	// Send to FCM native tokens via FCM V1 API
	if len(fcmTokens) > 0 && globalFCMClient != nil {
		go globalFCMClient.sendBatch(fcmTokens, title, body, extraData)
	}

	// Send to Expo tokens via Expo Push API (which internally uses FCM)
	if len(expoTokens) > 0 {
		imgURL := ""
		if extraData != nil {
			imgURL = extraData["image_url"]
		}
		sendExpoPushNotifications(expoTokens, title, body, imgURL)
	}
}

// sendBatch sends FCM notifications to a batch of native FCM tokens
func (f *fcmSenderClient) sendBatch(tokens []string, title, body string, extraData map[string]string) {
	accessToken, err := f.getOAuthToken()
	if err != nil {
		fmt.Printf("[FCM] Failed to get access token: %v\n", err)
		return
	}

	for _, token := range tokens {
		if err := f.sendOne(accessToken, token, title, body, extraData); err != nil {
			fmt.Printf("[FCM] Send error for token %s...: %v\n", token[:min(10, len(token))], err)
		}
	}
}

func (f *fcmSenderClient) sendOne(accessToken, token, title, body string, extraData map[string]string) error {
	data := map[string]string{"type": "BROADCAST"}
	for k, v := range extraData {
		data[k] = v
	}

	img := ensureAbsoluteURL(extraData["image_url"])
	if img != "" {
		data["image_url"] = img
	}

	notificationMap := map[string]interface{}{
		"title": title,
		"body":  body,
	}
	if img != "" {
		notificationMap["image"] = img
	}

	androidNotification := map[string]interface{}{
		"channel_id":            "aams_broadcasts",
		"sound":                 "default",
		"default_sound":         true,
		"visibility":            "PUBLIC",
		"notification_priority": "PRIORITY_MAX",
	}
	if img != "" {
		androidNotification["image"] = img
	}

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token":        token,
			"notification": notificationMap,
			"android": map[string]interface{}{
				"priority":     "high",
				"notification": androidNotification,
			},
			"data": data,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", f.sa.ProjectID)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("HTTP %d: %v", resp.StatusCode, errBody)
	}
	return nil
}

// getOAuthToken generates a short-lived Google OAuth2 access token using a service account JWT
func (f *fcmSenderClient) getOAuthToken() (string, error) {
	now := time.Now()
	tokenURI := f.sa.TokenURI
	if tokenURI == "" {
		tokenURI = "https://oauth2.googleapis.com/token"
	}

	// Build JWT
	headerJSON, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	claimsJSON, _ := json.Marshal(map[string]interface{}{
		"iss":   f.sa.ClientEmail,
		"scope": "https://www.googleapis.com/auth/firebase.messaging",
		"aud":   tokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	})

	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	claims := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := header + "." + claims

	// Parse RSA private key
	rsaKey, err := parseRSAKey(f.sa.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("parse RSA key: %w", err)
	}

	// Sign with RS256
	hash := sha256.Sum256([]byte(signingInput))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}
	sig := base64.RawURLEncoding.EncodeToString(sigBytes)
	jwtToken := signingInput + "." + sig

	// Exchange JWT for access token
	resp, err := http.PostForm(tokenURI, url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {jwtToken},
	})
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode token: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("%s: %s", result.Error, result.ErrorDesc)
	}
	return result.AccessToken, nil
}

func parseRSAKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("invalid PEM block")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return rsaKey, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
