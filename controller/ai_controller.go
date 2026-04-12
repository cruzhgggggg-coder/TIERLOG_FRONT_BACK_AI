package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"testing_go/koneksi"
	"testing_go/models"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

// ─────────────────────────────────────────────────────────────────────────────
//  SYSTEM PROMPT TEMPLATE
//  Exact text mandated by the lecturer. The placeholder string
//  "[INJECT_FETCHED_FEEDBACK_ITEMS_HERE]" is replaced at runtime using
//  strings.ReplaceAll so that the literal "100%" in the final paragraph
//  is never misinterpreted as a fmt.Sprintf verb.
// ─────────────────────────────────────────────────────────────────────────────

const feedbackPlaceholder = "[INJECT_FETCHED_FEEDBACK_ITEMS_HERE]"

const systemPromptTemplate = `Peran Utama Kamu adalah asisten pendukung dosen. Tugas utamamu bukan memberikan saran mandiri atau ide baru secara acak. Kamu berfungsi sebagai jembatan yang memperluas dan mengimplementasikan feedback yang telah diberikan oleh dosen kepada mahasiswa.

Alur Kerja (Wajib Urut)

Tahap Konsultasi: AI tidak diperbolehkan memberikan bantuan teknis atau saran materi sebelum mahasiswa memasukkan feedback resmi dari dosen.

Input Feedback: Berikut adalah poin-poin feedback resmi dari dosen yang harus dikerjakan mahasiswa:
[INJECT_FETCHED_FEEDBACK_ITEMS_HERE]

Generasi Persona: Setelah menerima feedback tersebut, kamu harus membentuk 'Persona Ahli' yang spesifik sesuai dengan arahan dosen. Misalnya, jika dosen meminta perbaikan pada metodologi, kamu berubah menjadi persona 'Pakar Metodologi Penelitian'.

Bantuan Kerja: Setelah persona terbentuk, kamu baru diperbolehkan membantu mahasiswa mengerjakan tugasnya dengan tetap berpegang teguh pada batasan feedback dosen tersebut.

Batasan Tindakan

Dilarang Memberi Saran Mandiri: Jangan memberikan instruksi yang bertentangan atau di luar lingkup feedback dosen.

Fokus Kelanjutan: Fokusmu hanya melanjutkan, merinci, dan membantu eksekusi dari apa yang sudah dikomentari dosen.

Verifikasi: Jika mahasiswa meminta bantuan di luar feedback yang ada, ingatkan mahasiswa untuk melakukan konsultasi ulang dengan dosen terlebih dahulu.

Tujuan Akhir Membantu mahasiswa menyelesaikan tugas dengan hasil yang selaras 100% dengan ekspektasi dan arahan dosen pembimbing.`

// ─────────────────────────────────────────────────────────────────────────────
//  CORE BUSINESS LOGIC
// ─────────────────────────────────────────────────────────────────────────────

// GenerateRevisionAssistance is the primary AI logic function.
//
// Workflow:
//  1. Query the DB for all FeedbackItem records associated with logID.
//  2. Guard: refuse to call the AI if no official feedback has been recorded
//     yet (Tahap Konsultasi rule — lecturer's strict constraint).
//  3. Format each FeedbackItem into a numbered bullet list.
//  4. Inject the formatted list into the System Prompt by replacing the
//     placeholder string.
//  5. Call the Gemini API with the constructed System Instruction and the
//     student's query as the user turn.
//  6. Return the AI-generated response text.
func GenerateRevisionAssistance(logID uint64, studentQuery string) (string, error) {

	// ── Step 1: Fetch feedback items from DB ─────────────────────────────────
	var feedbackItems []models.FeedbackItem
	if err := koneksi.DB.Where("log_id = ?", logID).Find(&feedbackItems).Error; err != nil {
		return "", fmt.Errorf("database error while fetching feedback: %w", err)
	}

	// ── Step 2: Guard — enforce "Tahap Konsultasi" rule ─────────────────────
	// The AI must NOT answer before official lecturer feedback is recorded.
	if len(feedbackItems) == 0 {
		return "", errors.New(
			"GUARDED: Belum ada feedback resmi dari dosen untuk sesi konsultasi ini. " +
				"AI tidak dapat memberikan bantuan sebelum feedback dosen diinputkan ke dalam sistem.",
		)
	}

	// ── Step 3: Format feedback items into a numbered bullet list ────────────
	// Each item shows its sequential number, category, status, and content.
	// Example:
	//   1. [Major | Pending] Perbaiki strukturisasi Bab 2 metodologi penelitian.
	//   2. [Minor | Fixed]   Tambahkan referensi pada paragraf pembuka.
	var lines []string
	for i, item := range feedbackItems {
		lines = append(lines, fmt.Sprintf(
			"%d. [%s | %s] %s",
			i+1,
			string(item.Category), // "Minor" or "Major"
			string(item.Status),   // "Fixed" or "Pending"
			item.Content,
		))
	}
	formattedFeedback := strings.Join(lines, "\n")

	// ── Step 4: Inject feedback into the System Prompt ───────────────────────
	// We use strings.ReplaceAll instead of fmt.Sprintf so that the literal
	// "100%" in the Tujuan Akhir paragraph is never treated as a format verb.
	finalSystemPrompt := strings.ReplaceAll(
		systemPromptTemplate,
		feedbackPlaceholder,
		formattedFeedback,
	)

	// ── Step 5: Call the Gemini API ──────────────────────────────────────────
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		// DEV MODE: no API key configured — return the assembled prompt so the
		// developer can verify injection is correct without hitting the API.
		devResponse := fmt.Sprintf(
			"[DEV MODE — GEMINI_API_KEY is not set]\n\n"+
				"[SYSTEM PROMPT INJECTED]\n%s\n\n"+
				"[STUDENT QUERY]\n%s",
			finalSystemPrompt,
			studentQuery,
		)
		return devResponse, nil
	}

	ctx := context.Background()

	// Initialise the Gemini client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("failed to initialise Gemini client: %w", err)
	}

	// Build and send the request.
	// The finalSystemPrompt is passed as a SystemInstruction so Gemini treats
	// it as the authoritative constraint for the entire conversation.
	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		genai.Text(studentQuery),
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(finalSystemPrompt, genai.RoleUser),
		},
	)
	if err != nil {
		return "", fmt.Errorf("Gemini API call failed: %w", err)
	}

	// ── Step 6: Extract and return the response text ─────────────────────────
	if result == nil || len(result.Candidates) == 0 {
		return "", errors.New("Gemini returned an empty response — no candidates")
	}

	return result.Text(), nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  HTTP HANDLER
// ─────────────────────────────────────────────────────────────────────────────

// AIAssistHandler godoc
//
//	@Summary      Ask AI for revision assistance
//	@Description  Sends the student query to Gemini, guarded by lecturer feedback.
//	              Returns 403 if no official feedback has been recorded yet.
//	@Tags         AI
//	@Accept       json
//	@Produce      json
//	@Param        body  body  object  true  "Request body"
//	@Success      200   {object}  object
//	@Failure      400   {object}  object
//	@Failure      403   {object}  object
//	@Failure      500   {object}  object
//	@Router       /api/ai/assist [post]
//
// POST /api/ai/assist
//
// Request body (JSON):
//
//	{
//	  "log_id": 1,
//	  "query":  "Bagaimana cara memperbaiki struktur Bab 2 saya?"
//	}
func AIAssistHandler(c *gin.Context) {
	var req struct {
		LogID uint64 `json:"log_id" binding:"required"`
		Query string `json:"query"  binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Request tidak valid. Field 'log_id' (angka) dan 'query' (string) wajib diisi.",
		})
		return
	}

	aiResponse, err := GenerateRevisionAssistance(req.LogID, req.Query)
	if err != nil {
		// Return 403 Forbidden for guarded refusals (no feedback recorded yet)
		if strings.HasPrefix(err.Error(), "GUARDED:") {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "guarded",
				"message": err.Error(),
			})
			return
		}
		// Return 500 for all other server / API errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "AI processing error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"log_id":      req.LogID,
		"ai_response": aiResponse,
	})
}
