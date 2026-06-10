package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"testing_go/auth"
	"testing_go/models"

	"google.golang.org/genai"
)

// ─────────────────────────────────────────────────────────────────────────────
//  ANNOTATION OCR — MULTI-PROVIDER VISION
// ─────────────────────────────────────────────────────────────────────────────

var annotationOCRPrompt = `Kamu adalah asisten pembaca dokumen akademik.
Di hadapanmu adalah foto halaman skripsi/tesis yang sudah dicoret-coret atau diberi anotasi oleh dosen pembimbing.
Tugasmu adalah membaca SEMUA catatan, coretan, tulisan tangan, garis bawah, dan anotasi yang ada di halaman ini.
Kembalikan daftar terstruktur dari setiap poin koreksi yang kamu temukan, dalam Bahasa Indonesia.
Format output:
- [Lokasi/halaman jika terlihat]: Deskripsi singkat isi koreksi

Jika tidak ada anotasi yang terbaca, tulis: "(Tidak ada anotasi yang terbaca di gambar ini)"`

// processAnnotationImage sends a saved image file to the user's selected Vision AI provider and returns OCR text.
// Supports Gemini, NVIDIA, OpenAI, and Anthropic.
func processAnnotationImage(imagePath string, user *models.User) (string, error) {
	provider := "gemini"
	model := ""
	apiKey := ""

	if user != nil {
		if user.PreferredModel != "" && user.PreferredModel != "default" {
			parts := strings.Split(user.PreferredModel, ":")
			if len(parts) == 2 {
				provider = strings.ToLower(parts[0])
				model = parts[1]
			}
		}

		switch provider {
		case "openai":
			apiKey = auth.DecryptAPIKey(user.OpenAIKey)
		case "nvidia":
			apiKey = auth.DecryptAPIKey(user.NvidiaKey)
		case "gemini":
			apiKey = auth.DecryptAPIKey(user.GeminiKey)
		case "anthropic":
			apiKey = auth.DecryptAPIKey(user.AnthropicKey)
		}
	}

	// Fallback to environment variables if keys are not set at the user level
	if provider == "nvidia" && apiKey == "" {
		apiKey = os.Getenv("NVIDIA_API_KEY")
	}
	if provider == "openai" && apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if provider == "gemini" && apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if provider == "anthropic" && apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	if apiKey == "" {
		return fmt.Sprintf("(OCR tidak tersedia: %s API key belum diatur di AI Gateway maupun di server)", strings.ToUpper(provider)), nil
	}

	// Read image bytes
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read annotation image: %v", err)
	}

	// Determine MIME type from extension
	ext := strings.ToLower(filepath.Ext(imagePath))
	mimeType := "image/jpeg"
	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".webp":
		mimeType = "image/webp"
	case ".gif":
		mimeType = "image/gif"
	}

	fmt.Printf("\033[35m[ANNOTATION OCR] Sending %s (%.2f KB) to %s Vision model: %s...\033[0m\n",
		filepath.Base(imagePath), float64(len(imgData))/1024, strings.ToUpper(provider), model)

	// Route based on provider
	switch provider {
	case "gemini":
		return callGeminiVision(apiKey, model, imgData, mimeType)
	case "nvidia":
		return callOpenAIVision("https://integrate.api.nvidia.com/v1/chat/completions", apiKey, model, imgData, mimeType)
	case "openai":
		return callOpenAIVision("https://api.openai.com/v1/chat/completions", apiKey, model, imgData, mimeType)
	case "anthropic":
		return callAnthropicVision(apiKey, model, imgData, mimeType)
	default:
		return callGeminiVision(apiKey, model, imgData, mimeType)
	}
}

func callGeminiVision(apiKey, model string, imgData []byte, mimeType string) (string, error) {
	if model == "" {
		model = os.Getenv("GEMINI_ANNOTATION_MODEL")
		if model == "" {
			model = "gemini-2.0-flash"
		}
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return fmt.Sprintf("(OCR gagal: %v)", err), nil
	}

	content := &genai.Content{
		Parts: []*genai.Part{
			{Text: annotationOCRPrompt},
			{
				InlineData: &genai.Blob{
					MIMEType: mimeType,
					Data:     imgData,
				},
			},
		},
	}

	resp, err := client.Models.GenerateContent(ctx, model, []*genai.Content{content}, nil)
	if err != nil {
		fmt.Printf("\033[31m[ANNOTATION OCR] Gemini Vision failed: %v\033[0m\n", err)
		return fmt.Sprintf("(OCR gagal: %v)", err), nil
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "(OCR: Tidak ada respons dari Gemini)", nil
	}

	result := resp.Candidates[0].Content.Parts[0].Text
	fmt.Printf("\033[32m[ANNOTATION OCR] Done — %d chars extracted\033[0m\n", len(result))
	return result, nil
}

func callOpenAIVision(endpointUrl, apiKey, model string, imgData []byte, mimeType string) (string, error) {
	if model == "" {
		if strings.Contains(endpointUrl, "nvidia") {
			model = "meta/llama-3.2-90b-vision-instruct"
		} else {
			model = "gpt-4o-mini"
		}
	}

	base64Img := base64.StdEncoding.EncodeToString(imgData)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Img)

	reqPayload := map[string]interface{}{
		"model": model,
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "system",
				"content": annotationOCRPrompt,
			},
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Analisis gambar anotasi dosen ini.",
					},
					map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]interface{}{
							"url": dataURI,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpointUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("(OCR gagal menghubungi API: %v)", err), nil
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("(OCR gagal, API mengembalikan status %d: %s)", resp.StatusCode, string(bodyBytes)), nil
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "(OCR: Tidak ada respons dari model vision)", nil
	}

	result := response.Choices[0].Message.Content
	fmt.Printf("\033[32m[ANNOTATION OCR] Done — %d chars extracted\033[0m\n", len(result))
	return result, nil
}

func callAnthropicVision(apiKey, model string, imgData []byte, mimeType string) (string, error) {
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	base64Img := base64.StdEncoding.EncodeToString(imgData)

	reqPayload := map[string]interface{}{
		"model":      model,
		"max_tokens": 4096,
		"system":     annotationOCRPrompt,
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "image",
						"source": map[string]interface{}{
							"type":       "base64",
							"media_type": mimeType,
							"data":       base64Img,
						},
					},
					map[string]interface{}{
						"type": "text",
						"text": "Analisis gambar anotasi dosen ini.",
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("(OCR gagal menghubungi Anthropic: %v)", err), nil
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("(OCR gagal, Anthropic mengembalikan status %d: %s)", resp.StatusCode, string(bodyBytes)), nil
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", err
	}

	if len(response.Content) == 0 {
		return "(OCR: Tidak ada respons dari Claude)", nil
	}

	result := response.Content[0].Text
	fmt.Printf("\033[32m[ANNOTATION OCR] Done — %d chars extracted\033[0m\n", len(result))
	return result, nil
}
