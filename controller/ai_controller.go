package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"testing_go/auth"
	"testing_go/koneksi"
	"testing_go/middleware"
	"testing_go/models"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

// ─────────────────────────────────────────────────────────────────────────────
//  JSON SANITIZER — LLM Output Cleanup
// ─────────────────────────────────────────────────────────────────────────────

// sanitizeJSON processes raw LLM output and escapes illegal control characters
// inside JSON string values. It uses a state-machine approach to track whether
// the reader is currently inside a JSON string literal (between double quotes).
//
// Characters escaped:
//
//	\n → \\n
//	\r → \\r
//	\t → \\t
//	Other control chars (< 0x20) → \\uXXXX
//
// Characters outside string literals are passed through unchanged.
func sanitizeJSON(input string) string {
	var result strings.Builder
	inString := false
	escaped := false
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if inString {
			if escaped {
				// Previous char was backslash — this char is part of escape sequence
				result.WriteRune(r)
				escaped = false
				continue
			}
			if r == '\\' {
				result.WriteRune(r)
				escaped = true
				continue
			}
			if r == '"' {
				// End of string
				inString = false
				result.WriteRune(r)
				continue
			}
			// Inside string, check for illegal control characters
			if r == '\n' {
				result.WriteString("\\n")
			} else if r == '\r' {
				result.WriteString("\\r")
			} else if r == '\t' {
				result.WriteString("\\t")
			} else if r < 0x20 {
				result.WriteString(fmt.Sprintf("\\u%04x", r))
			} else {
				result.WriteRune(r)
			}
		} else {
			// Outside string
			if r == '"' {
				inString = true
			}
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ─────────────────────────────────────────────────────────────────────────────
//  PROMPT CONSTANTS
// ─────────────────────────────────────────────────────────────────────────────

const personaDosenPrompt = `Kamu adalah Dosen Pembimbing yang teliti dan suportif. 
Tugasmu adalah menganalisis Draft Paper Mahasiswa berdasarkan Transkrip Bimbingan (Instruksi Dosen).

PRINSIP UTAMA:
1. Dilarang berhalusinasi atau memberikan ide baru yang tidak ada di transkrip.
2. Instruksi 100% berasal dari teks transkrip rekaman.
3. Bandingkan draf mahasiswa dengan poin-poin dalam transkrip.
4. Hasilkan daftar tugas revisi yang spesifik.

KATEGORI FEEDBACK:
- HOC (Higher Order Concerns): Fokus pada substansi seperti struktur, argumen, metodologi, dan kesesuaian judul.
- LOC (Lower Order Concerns): Fokus pada teknis seperti penulisan, typo, format sitasi, dan tata bahasa.

TATA CARA OUTPUT:
Kamu WAJIB mengembalikan output dalam format JSON dengan struktur:
{
  "feedbacks": [
    {"content": "...", "category": "HOC"},
    {"content": "...", "category": "LOC"}
  ]
}`

const systemPromptTemplate = `Peran Utama Kamu adalah asisten pendukung dosen. Tugas utamamu bukan memberikan saran mandiri atau ide baru secara acak. Kamu berfungsi sebagai jembatan yang memperluas dan mengimplementasikan feedback yang telah diberikan oleh dosen kepada mahasiswa.

ALUR KERJA:
1. Input Feedback: Berikut adalah poin-poin feedback resmi dari dosen:
[INJECT_FETCHED_FEEDBACK_ITEMS_HERE]

2. Konteks Asli: Berikut adalah transkrip asli dari sesi bimbingan tersebut sebagai referensi tambahan:
[INJECT_ORIGINAL_TRANSCRIPT_HERE]

3. Bantuan Kerja: Gunakan feedback di atas sebagai batasan utamamu. Jika mahasiswa bertanya, jawablah dengan persona 'Pakar' yang relevan dengan topik feedback tersebut.

BATASAN:
- Dilarang memberi saran yang bertentangan dengan feedback dosen.
- Jika mahasiswa meminta bantuan di luar cakupan feedback, ingatkan mereka untuk konsultasi lagi dengan dosen.`

const feedbackPlaceholder = "[INJECT_FETCHED_FEEDBACK_ITEMS_HERE]"
const transcriptPlaceholder = "[INJECT_ORIGINAL_TRANSCRIPT_HERE]"

// ─────────────────────────────────────────────────────────────────────────────
//  GROQ API LOGIC (AUDIO TO TEXT) — with chunking for long recordings
// ─────────────────────────────────────────────────────────────────────────────

const (
	maxChunkBytes     int64         = 20 * 1024 * 1024 // 20 MB
	groqTimeout                     = 300 * time.Second
	maxAIResponseSize int64         = 10 * 1024 * 1024 // 10 MB
	maxHTTPBodySize   int64         = 50 * 1024 * 1024 // 50 MB
)

// transcribeChunk sends a single raw audio byte slice to Groq Whisper and returns the transcript text.
func transcribeChunk(apiKey string, audioData []byte, filename string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, bytes.NewReader(audioData)); err != nil {
		return "", err
	}

	writer.WriteField("model", "whisper-large-v3")
	writer.WriteField("response_format", "text")
	writer.WriteField("language", "id")
	writer.Close()

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/audio/transcriptions", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: groqTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Groq API error (%d): %s", resp.StatusCode, string(respBody))
	}
	return string(respBody), nil
}

// transcribeAudio reads the audio file, optionally splits it into ≤20 MB byte chunks,
// transcribes each chunk via Groq Whisper, and returns the stitched full transcript.
func transcribeAudio(userApiKey string, audioPath string) (string, error) {
	apiKey := userApiKey
	if apiKey == "" {
		apiKey = os.Getenv("GROQ_API_KEY")
	}
	if apiKey == "" {
		fmt.Println("\033[31m[GROQ STT] Warning: GROQ_API_KEY is not set. Audio transcription is disabled.\033[0m")
		return "Transkripsi dinonaktifkan: Groq API key belum dikonfigurasi di AI Gateway maupun di server.", nil
	}

	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %v", err)
	}

	ext := filepath.Ext(audioPath)
	totalSize := int64(len(audioData))

	fmt.Printf("\033[36m[GROQ STT] File: %s | Size: %.2f MB\033[0m\n",
		filepath.Base(audioPath), float64(totalSize)/(1024*1024))

	// ── Fast path: file fits in a single request ─────────────────────────────
	if totalSize <= maxChunkBytes {
		fmt.Printf("\033[36m[GROQ STT] Single-chunk mode — sending directly to Whisper...\033[0m\n")
		return transcribeChunk(apiKey, audioData, filepath.Base(audioPath))
	}

	// ── Chunked path: split file into ≤20 MB byte slices ────────────────────
	totalChunks := int((totalSize + maxChunkBytes - 1) / maxChunkBytes) // ceiling division
	fmt.Printf("\033[33m[GROQ STT] File exceeds 20 MB — splitting into %d chunks...\033[0m\n", totalChunks)

	var transcripts []string
	var offset int64
	chunkNum := 0

	for offset < totalSize {
		end := offset + maxChunkBytes
		if end > totalSize {
			end = totalSize
		}

		chunk := audioData[offset:end]
		chunkNum++
		chunkSizeMB := float64(len(chunk)) / (1024 * 1024)
		chunkFilename := fmt.Sprintf("chunk_%d_of_%d%s", chunkNum, totalChunks, ext)

		fmt.Printf("\033[36m[GROQ STT] Chunk %d/%d (%.2f MB) — transcribing...\033[0m\n",
			chunkNum, totalChunks, chunkSizeMB)

		text, err := transcribeChunk(apiKey, chunk, chunkFilename)
		if err != nil {
			return "", fmt.Errorf("[GROQ STT] chunk %d/%d failed: %v", chunkNum, totalChunks, err)
		}

		trimmed := strings.TrimSpace(text)
		if trimmed != "" {
			transcripts = append(transcripts, trimmed)
		}
		fmt.Printf("\033[32m[GROQ STT] Chunk %d/%d done — %d chars transcribed\033[0m\n",
			chunkNum, totalChunks, len(trimmed))

		offset = end
	}

	fullTranscript := strings.Join(transcripts, " ")
	fmt.Printf("\033[32m[GROQ STT] All %d chunks done — total transcript: %d chars\033[0m\n",
		totalChunks, len(fullTranscript))

	return fullTranscript, nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  NVIDIA NIM API LOGIC
// ─────────────────────────────────────────────────────────────────────────────

type NVIDIARequest struct {
	Model          string          `json:"model"`
	Messages       []NVIDIAMessage `json:"messages"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

type NVIDIAMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type NVIDIAResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func callNVIDIA(apiKey, model, systemPrompt, userPrompt string, isJSON bool) (string, error) {
	if apiKey == "" {
		apiKey = os.Getenv("NVIDIA_API_KEY")
	}
	if apiKey == "" {
		return "", errors.New("NVIDIA_API_KEY is not set")
	}

	if model == "" {
		model = os.Getenv("NVIDIA_DEFAULT_MODEL")
		if model == "" {
			model = "meta/llama-3.1-70b-instruct"
		}
	}

	url := "https://integrate.api.nvidia.com/v1/chat/completions"
	
	reqBody := NVIDIARequest{
		Model: model,
		Messages: []NVIDIAMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	
	if isJSON {
		reqBody.ResponseFormat = &struct {
			Type string `json:"type"`
		}{Type: "json_object"}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal NVIDIA request: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create NVIDIA request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxHTTPBodySize))

	if resp.StatusCode != http.StatusOK {
		if (resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest) && model != "meta/llama-3.1-70b-instruct" {
			return callNVIDIA(apiKey, "meta/llama-3.1-70b-instruct", systemPrompt, userPrompt, isJSON)
		}

		if isJSON && reqBody.ResponseFormat != nil {
			reqBody.ResponseFormat = nil
			retryJsonData, err := json.Marshal(reqBody)
			if err != nil {
				return "", fmt.Errorf("failed to marshal NVIDIA retry request: %w", err)
			}
			retryReq, err := http.NewRequest("POST", url, bytes.NewBuffer(retryJsonData))
			if err != nil {
				return "", fmt.Errorf("failed to create NVIDIA retry request: %w", err)
			}
			retryReq.Header.Set("Content-Type", "application/json")
			retryReq.Header.Set("Authorization", "Bearer "+apiKey)
			
			retryResp, err := client.Do(retryReq)
			if err == nil {
				defer retryResp.Body.Close()
				retryBody, _ := io.ReadAll(io.LimitReader(retryResp.Body, maxHTTPBodySize))
				if retryResp.StatusCode == http.StatusOK {
					var retryNvidiaResp NVIDIAResponse
					if err := json.Unmarshal(retryBody, &retryNvidiaResp); err == nil && len(retryNvidiaResp.Choices) > 0 {
						return retryNvidiaResp.Choices[0].Message.Content, nil
					}
				}
				body = retryBody
				resp.StatusCode = retryResp.StatusCode
			}
		}
		return "", fmt.Errorf("NVIDIA API error (%d): %s", resp.StatusCode, string(body))
	}

	var nvidiaResp NVIDIAResponse
	if err := json.Unmarshal(body, &nvidiaResp); err != nil {
		return "", err
	}

	if len(nvidiaResp.Choices) == 0 {
		return "", errors.New("NVIDIA returned no choices")
	}

	return nvidiaResp.Choices[0].Message.Content, nil
}

func callAnthropic(apiKey, model, systemPrompt, userPrompt string) (string, error) {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return "", errors.New("ANTHROPIC_API_KEY is not set")
	}

	if model == "" {
		model = os.Getenv("ANTHROPIC_DEFAULT_MODEL")
		if model == "" {
			model = "claude-3-5-sonnet-20240620"
		}
	}

	url := "https://api.anthropic.com/v1/messages"
	
	reqBody := struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		System    string `json:"system"`
		Messages  []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}{
		Model:     model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "user", Content: userPrompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Anthropic request: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create Anthropic request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxHTTPBodySize))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic API error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Content) == 0 {
		return "", errors.New("Anthropic returned no content")
	}

	return result.Content[0].Text, nil
}

func callAI(user *models.User, systemPrompt, userPrompt string, isJSON bool) (string, error) {
	provider := strings.ToLower(os.Getenv("AI_PROVIDER"))
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

	switch provider {
	case "gemini":
		return callGemini(apiKey, model, systemPrompt, userPrompt, isJSON)
	case "anthropic":
		return callAnthropic(apiKey, model, systemPrompt, userPrompt)
	case "openai":
		return callOpenAI(apiKey, model, systemPrompt, userPrompt, isJSON)
	case "nvidia":
		return callNVIDIA(apiKey, model, systemPrompt, userPrompt, isJSON)
	default:
		return callNVIDIA("", "", systemPrompt, userPrompt, isJSON)
	}
}

// Helper to use NVIDIA-style call for generic OpenAI compatible APIs
func callOpenAI(apiKey, model, systemPrompt, userPrompt string, isJSON bool) (string, error) {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY is not set")
	}
	if model == "" {
		model = os.Getenv("OPENAI_DEFAULT_MODEL")
		if model == "" {
			model = "gpt-4o"
		}
	}

	url := "https://api.openai.com/v1/chat/completions"
	
	reqBody := NVIDIARequest{
		Model: model,
		Messages: []NVIDIAMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	
	if isJSON {
		reqBody.ResponseFormat = &struct {
			Type string `json:"type"`
		}{Type: "json_object"}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OpenAI request: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create OpenAI request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxHTTPBodySize))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, string(body))
	}

	var nvidiaResp NVIDIAResponse
	if err := json.Unmarshal(body, &nvidiaResp); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(nvidiaResp.Choices) == 0 {
		return "", errors.New("OpenAI returned no choices")
	}

	return nvidiaResp.Choices[0].Message.Content, nil
}

func callGemini(apiKey, model, systemPrompt, userPrompt string, isJSON bool) (string, error) {
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		return "", errors.New("GEMINI_API_KEY is not set")
	}

	if model == "" {
		model = os.Getenv("GEMINI_DEFAULT_MODEL")
		if model == "" {
			model = "gemini-2.0-flash"
		}
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %v", err)
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
	}

	if isJSON {
		config.ResponseMIMEType = "application/json"
	}

	content := &genai.Content{
		Parts: []*genai.Part{{Text: userPrompt}},
	}

	resp, err := client.Models.GenerateContent(ctx, model, []*genai.Content{content}, config)
	if err != nil {
		if isJSON {
			config.ResponseMIMEType = ""
			resp, err = client.Models.GenerateContent(ctx, model, []*genai.Content{content}, config)
		}
		if err != nil {
			return "", fmt.Errorf("Gemini API error: %v", err)
		}
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("Gemini returned no content")
	}

	return resp.Candidates[0].Content.Parts[0].Text, nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  ANALYSIS LOGIC (GROQ + NVIDIA)
// ─────────────────────────────────────────────────────────────────────────────

type FeedbackResponse struct {
	Content  string `json:"content"`
	Category string `json:"category"`
}

// buildAnalysisPrompt constructs the user prompt adapting to whether a transcript exists.
func buildAnalysisPrompt(transcript, paperText, prevFeedback string) string {
	var parts []string

	if transcript != "" {
		parts = append(parts, fmt.Sprintf(
			"Berikut adalah transkrip audio bimbingan dosen:\n\"%s\"", transcript))
	} else {
		parts = append(parts,
			"Tidak ada rekaman audio bimbingan. Analisis hanya berdasarkan draf paper dan anotasi visual dosen.")
	}

	parts = append(parts, fmt.Sprintf("\n\nDan ini adalah paper mahasiswa:\n\n%s", paperText))

	if prevFeedback != "" {
		parts = append(parts, fmt.Sprintf(
			"\n\nKONTEKS REVISI (Feedback Sesi Sebelumnya):\n%s\n\nTugas tambahanmu: Cek apakah mahasiswa sudah memperbaiki poin-poin di atas dalam draf baru ini. Jika belum, sertakan kembali dalam daftar feedback.", prevFeedback))
	}

	parts = append(parts, "\n\nBerikan analisis revisi (HOC/LOC) berdasarkan informasi yang tersedia.")
	return strings.Join(parts, "")
}

func AnalyzeAudioAndPaper(userID uint64, audioPath, paperText, prevFeedback string) ([]models.FeedbackItem, string, error) {
	fmt.Printf("\033[34m[AI ANALYSIS] Starting paper analysis for User ID: %d...\033[0m\n", userID)

	// Fetch user for AI Gateway settings
	var user models.User
	koneksi.DB.First(&user, userID)

	var transcript string

	if audioPath != "" {
		// Audio present: transcribe via Groq Whisper
		var err error
		userGroqKey := ""
		if user.GroqKey != "" {
			userGroqKey = auth.DecryptAPIKey(user.GroqKey)
		}
		transcript, err = transcribeAudio(userGroqKey, audioPath)
		if err != nil {
			fmt.Printf("Warning: Transcription failed: %v\n", err)
			transcript = "Transkripsi Audio Gagal: " + err.Error()
		}
	} else {
		// No audio: use annotation text + paper text only
		transcript = ""
	}

	// Build adaptive prompt
	systemPrompt := personaDosenPrompt
	userPrompt := buildAnalysisPrompt(transcript, paperText, prevFeedback)

	// Determine active provider/model for analysis log
	provider := strings.ToLower(os.Getenv("AI_PROVIDER"))
	model := "default"
	if user.PreferredModel != "" && user.PreferredModel != "default" {
		parts := strings.Split(user.PreferredModel, ":")
		if len(parts) == 2 {
			provider = strings.ToLower(parts[0])
			model = parts[1]
		}
	}
	fmt.Printf("\033[34m[AI ANALYSIS] Calling %s model: %s...\033[0m\n", strings.ToUpper(provider), model)

	rawResponse, err := callAI(&user, systemPrompt, userPrompt, true)
	if err != nil {
		fmt.Printf("\033[31m[AI ANALYSIS] Failed: %v\033[0m\n", err)
		return nil, transcript, err
	}

	fmt.Printf("\033[32m[AI ANALYSIS] Done — Received raw response (%d chars)\033[0m\n", len(rawResponse))

	// Sanitize JSON before parsing
	cleanJSON := sanitizeJSON(rawResponse)
	cleanJSON = extractJSONBounds(cleanJSON)
	cleanJSON = strings.TrimSpace(cleanJSON)

	var aiResponse struct {
		Feedbacks []FeedbackResponse `json:"feedbacks"`
	}

	if err := json.Unmarshal([]byte(cleanJSON), &aiResponse); err != nil {
		start := strings.Index(cleanJSON, "{")
		end := strings.LastIndex(cleanJSON, "}")
		if start != -1 && end != -1 && end > start {
			if err2 := json.Unmarshal([]byte(cleanJSON[start:end+1]), &aiResponse); err2 == nil {
				goto PROCESS
			}
		}
		return nil, transcript, fmt.Errorf("failed to parse AI response: %w. Raw: %s", err, rawResponse)
	}

PROCESS:
	var items []models.FeedbackItem
	for _, f := range aiResponse.Feedbacks {
		category := models.CategoryMinor
		catUpper := strings.ToUpper(f.Category)
		if catUpper == "HOC" || catUpper == "MAJOR" {
			category = models.CategoryMajor
		}
		items = append(items, models.FeedbackItem{
			Content:  f.Content,
			Category: category,
			Status:   models.StatusPending,
		})
	}

	return items, transcript, nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  CONVERSATIONAL ASSISTANCE
// ─────────────────────────────────────────────────────────────────────────────

func GenerateRevisionAssistance(logID uint64, studentQuery string, modelOverride string) (string, error) {
	var log models.ConsultationLog
	if err := koneksi.DB.Preload("FeedbackItems").Preload("Student.User").First(&log, logID).Error; err != nil {
		return "", fmt.Errorf("database error: %w", err)
	}

	if len(log.FeedbackItems) == 0 {
		return "", errors.New("GUARDED: Belum ada feedback resmi.")
	}

	var feedbackLines []string
	for i, item := range log.FeedbackItems {
		feedbackLines = append(feedbackLines, fmt.Sprintf("%d. [%s] %s", i+1, item.Category, item.Content))
	}
	formattedFeedback := strings.Join(feedbackLines, "\n")

	finalSystemPrompt := strings.ReplaceAll(systemPromptTemplate, feedbackPlaceholder, formattedFeedback)
	finalSystemPrompt = strings.ReplaceAll(finalSystemPrompt, transcriptPlaceholder, log.TranscriptText)

	// Apply model override if provided
	user := log.Student.User
	if modelOverride != "" && modelOverride != "default" {
		user.PreferredModel = modelOverride
		user.IsGatewayActive = true // Force active if a specific model is chosen
	}

	return callAI(user, finalSystemPrompt, studentQuery, false)
}

// ─────────────────────────────────────────────────────────────────────────────
//  HTTP HANDLERS
// ─────────────────────────────────────────────────────────────────────────────

func AIAssistHandler(c *gin.Context) {
	var req struct {
		LogID uint64 `json:"log_id" binding:"required"`
		Query string `json:"query"  binding:"required"`
		Model string `json:"model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	response, err := GenerateRevisionAssistance(req.LogID, req.Query, req.Model)
	if err != nil {
		if strings.HasPrefix(err.Error(), "GUARDED:") {
			c.JSON(http.StatusForbidden, gin.H{"status": "guarded", "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "ai_response": response})
}

func GetAIModels(c *gin.Context) {
	provider := c.Query("provider")

	if provider != "nvidia" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only nvidia provider supports dynamic models currently"})
		return
	}

	apiKey := c.Query("api_key")
	if apiKey == "" {
		apiKey = c.GetHeader("X-API-Key")
	}
	if apiKey == "" {
		if user := middleware.CurrentUser(c); user != nil && user.NvidiaKey != "" {
			apiKey = auth.DecryptAPIKey(user.NvidiaKey)
		}
	}
	if apiKey == "" {
		apiKey = auth.AuthorizationToken(c.GetHeader("Authorization"))
	}

	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key is required"})
		return
	}

	req, _ := http.NewRequest("GET", "https://integrate.api.nvidia.com/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": string(body)})
		return
	}

	var response struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode models response"})
		return
	}

	var models []string
	for _, m := range response.Data {
		models = append(models, m.ID)
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

// ─────────────────────────────────────────────────────────────────────────────
//  PROVIDER-AGNOSTIC MODEL FILTERING
// ─────────────────────────────────────────────────────────────────────────────

// ListFilteredModels returns models filtered by capability requirements.
// Query params:
//   - capability: comma-separated list (e.g., "vision,json"). Default: all.
//   - provider: optional filter by provider name.
func ListFilteredModels(c *gin.Context) {
	user := middleware.CurrentUser(c)

	// Parse capability filter
	capParam := c.DefaultQuery("capability", "")
	var requiredCaps []ModelCapability
	if capParam != "" {
		for _, s := range strings.Split(capParam, ",") {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				requiredCaps = append(requiredCaps, ModelCapability(trimmed))
			}
		}
	}

	// Parse provider filter
	providerFilter := c.Query("provider")

	result := FilterModelsByCapability(user, requiredCaps)

	// Apply provider filter if specified
	if providerFilter != "" {
		var filtered []ModelInfo
		for _, m := range result {
			if m.Provider == providerFilter {
				filtered = append(filtered, m)
			}
		}
		result = filtered
	}

	c.JSON(http.StatusOK, gin.H{"models": result})
}
