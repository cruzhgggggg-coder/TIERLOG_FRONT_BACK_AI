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

	"testing_go/koneksi"
	"testing_go/models"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

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
//  GROQ API LOGIC (AUDIO TO TEXT)
// ─────────────────────────────────────────────────────────────────────────────

func transcribeAudio(audioPath string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", errors.New("GROQ_API_KEY is not set in .env")
	}

	file, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}
	
	writer.WriteField("model", "whisper-large-v3")
	writer.WriteField("response_format", "text")
	writer.WriteField("language", "id") // Assuming Indonesian
	writer.Close()

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/audio/transcriptions", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 120 * time.Second}
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

func callNVIDIA(systemPrompt, userPrompt string, isJSON bool) (string, error) {
	apiKey := os.Getenv("NVIDIA_API_KEY")
	if apiKey == "" {
		return "", errors.New("NVIDIA_API_KEY is not set")
	}

	fmt.Printf("\033[34m[NVIDIA NIM] Initiating request (Model: meta/llama-3.1-70b-instruct)...\033[0m\n")

	url := "https://integrate.api.nvidia.com/v1/chat/completions"
	
	reqBody := NVIDIARequest{
		Model: "meta/llama-3.1-70b-instruct",
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

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("\033[31m[NVIDIA NIM] Request Failed: %v\033[0m\n", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("\033[32m[NVIDIA NIM] Response Received! Status: %d, Size: %d bytes\033[0m\n", resp.StatusCode, len(body))

	if resp.StatusCode != http.StatusOK {
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

func callGemini(systemPrompt, userPrompt string, isJSON bool) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", errors.New("GEMINI_API_KEY is not set")
	}

	fmt.Printf("\033[35m[GEMINI AI] Initiating request (Model: gemini-2.0-flash)...\033[0m\n")

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
	})
	if err != nil {
		fmt.Printf("\033[31m[GEMINI AI] Failed to create client: %v\033[0m\n", err)
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

	// Use a single Content struct with a single Part
	content := &genai.Content{
		Parts: []*genai.Part{{Text: userPrompt}},
	}

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash", []*genai.Content{content}, config)
	if err != nil {
		// Second attempt with simplified configuration if MIME type causes issues
		if isJSON {
			config.ResponseMIMEType = ""
			resp, err = client.Models.GenerateContent(ctx, "gemini-2.0-flash", []*genai.Content{content}, config)
		}
		
		if err != nil {
			fmt.Printf("\033[31m[GEMINI AI] Request Failed: %v\033[0m\n", err)
			return "", fmt.Errorf("Gemini API error: %v", err)
		}
	}

	fmt.Printf("\033[32m[GEMINI AI] Response Received! Candidates: %d\033[0m\n", len(resp.Candidates))

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("Gemini returned no content")
	}

	return resp.Candidates[0].Content.Parts[0].Text, nil
}

func callAI(systemPrompt, userPrompt string, isJSON bool) (string, error) {
	provider := strings.ToLower(os.Getenv("AI_PROVIDER"))
	if provider == "gemini" {
		return callGemini(systemPrompt, userPrompt, isJSON)
	}
	// Default to NVIDIA for backward compatibility or if set explicitly
	return callNVIDIA(systemPrompt, userPrompt, isJSON)
}

// ─────────────────────────────────────────────────────────────────────────────
//  ANALYSIS LOGIC (GROQ + NVIDIA)
// ─────────────────────────────────────────────────────────────────────────────

type FeedbackResponse struct {
	Content  string `json:"content"`
	Category string `json:"category"`
}

func AnalyzeAudioAndPaper(audioPath, paperText, prevFeedback string) ([]models.FeedbackItem, string, error) {
	// 1. Convert Audio to Text using Groq Whisper
	transcript, err := transcribeAudio(audioPath)
	if err != nil {
		// Log the error but continue to try analyzing just the paper if possible
		fmt.Printf("Warning: Transcription failed: %v\n", err)
		transcript = "Transkripsi Audio Gagal: " + err.Error()
	}

	// 2. Analyze the Transcript and Paper with AI
	systemPrompt := personaDosenPrompt
	
	consistencyContext := ""
	if prevFeedback != "" {
		consistencyContext = fmt.Sprintf("\n\nKONTEKS REVISI (Feedback Sesi Sebelumnya):\n%s\n\nTugas tambahanmu: Cek apakah mahasiswa sudah memperbaiki poin-poin di atas dalam draf baru ini. Jika belum, sertakan kembali dalam daftar feedback.", prevFeedback)
	}

	userPrompt := fmt.Sprintf("Berikut adalah transkrip audio bimbingan dosen:\n\"%s\"\n\nDan ini adalah paper mahasiswa:\n\n%s%s\n\nBerikan analisis revisi (HOC/LOC) berdasarkan transkrip tersebut.", transcript, paperText, consistencyContext)

	rawResponse, err := callAI(systemPrompt, userPrompt, true)
	if err != nil {
		return nil, transcript, err
	}

	// Clean up markdown code blocks if present
	cleanJSON := strings.TrimSpace(rawResponse)
	if strings.HasPrefix(cleanJSON, "```json") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
		cleanJSON = strings.TrimSuffix(cleanJSON, "```")
	} else if strings.HasPrefix(cleanJSON, "```") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```")
		cleanJSON = strings.TrimSuffix(cleanJSON, "```")
	}
	cleanJSON = strings.TrimSpace(cleanJSON)

	var aiResponse struct {
		Feedbacks []FeedbackResponse `json:"feedbacks"`
	}

	if err := json.Unmarshal([]byte(cleanJSON), &aiResponse); err != nil {
		// Fallback: search for { ... } in case of extra text
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

func GenerateRevisionAssistance(logID uint64, studentQuery string) (string, error) {
	var log models.ConsultationLog
	if err := koneksi.DB.Preload("FeedbackItems").First(&log, logID).Error; err != nil {
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

	return callAI(finalSystemPrompt, studentQuery, false)
}

// ─────────────────────────────────────────────────────────────────────────────
//  HTTP HANDLERS
// ─────────────────────────────────────────────────────────────────────────────

func AIAssistHandler(c *gin.Context) {
	var req struct {
		LogID uint64 `json:"log_id" binding:"required"`
		Query string `json:"query"  binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	response, err := GenerateRevisionAssistance(req.LogID, req.Query)
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
