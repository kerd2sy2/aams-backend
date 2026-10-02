package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2/google"
)

func main() {
	ctx := context.Background()
	saPath := "d:\\AAMS\\aams-8a73f-firebase-adminsdk-fbsvc-aa4581dbbd.json"
	saBytes, err := ioutil.ReadFile(saPath)
	if err != nil {
		fmt.Println("Error reading SA file:", err)
		return
	}

	creds, err := google.CredentialsFromJSON(ctx, saBytes, "https://www.googleapis.com/auth/cloud-vision", "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		fmt.Println("Error creating credentials:", err)
		return
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		fmt.Println("Error getting token:", err)
		return
	}
	fmt.Println("Successfully got OAuth token! Testing Cloud Vision OCR on sample images...")

	// Test on sample images in D:\AAMS\لوحات_الدبابات
	files, err := ioutil.ReadDir("d:\\AAMS\\لوحات_الدبابات")
	if err != nil {
		fmt.Println("Error reading dir:", err)
		return
	}

	for i, f := range files {
		if i >= 5 {
			break
		}
		imgPath := filepath.Join("d:\\AAMS\\لوحات_الدبابات", f.Name())
		imgData, err := ioutil.ReadFile(imgPath)
		if err != nil {
			continue
		}
		imgB64 := base64.StdEncoding.EncodeToString(imgData)

		reqBody := map[string]interface{}{
			"requests": []interface{}{
				map[string]interface{}{
					"image": map[string]interface{}{
						"content": imgB64,
					},
					"features": []interface{}{
						map[string]interface{}{
							"type": "TEXT_DETECTION",
						},
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "https://vision.googleapis.com/v1/images:annotate", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Vision request error:", err)
			continue
		}
		defer resp.Body.Close()
		respBytes, _ := ioutil.ReadAll(resp.Body)

		var visionResp struct {
			Responses []struct {
				FullTextAnnotation struct {
					Text string `json:"text"`
				} `json:"fullTextAnnotation"`
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			} `json:"responses"`
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.Unmarshal(respBytes, &visionResp)

		if visionResp.Error.Message != "" {
			fmt.Printf("[%s] Vision API Error: %s\n", f.Name(), visionResp.Error.Message)
		} else if len(visionResp.Responses) > 0 {
			if visionResp.Responses[0].Error.Message != "" {
				fmt.Printf("[%s] Error: %s\n", f.Name(), visionResp.Responses[0].Error.Message)
			} else {
				text := strings.TrimSpace(visionResp.Responses[0].FullTextAnnotation.Text)
				fmt.Printf("=== [%s] OCR DETECTED ===\n%s\n=========================\n\n", f.Name(), text)
			}
		}
	}
}
