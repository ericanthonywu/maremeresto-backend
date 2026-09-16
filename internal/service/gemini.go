package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/dto"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/model"
)

type GeminiClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewGeminiClient(apiKey, model string) *GeminiClient {
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &GeminiClient{
		apiKey: strings.TrimSpace(apiKey),
		model:  model,
		httpClient: &http.Client{
			Timeout: 35 * time.Second,
		},
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	ResponseMimeType string `json:"responseMimeType,omitempty"`
}

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiCandidate struct {
	Content struct {
		Parts []geminiPart `json:"parts"`
	} `json:"content"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

type aiOutputJSON struct {
	ExecutiveSummary      string   `json:"executive_summary"`
	DetailSummary         string   `json:"detail_summary"`
	ActionableSuggestions []string `json:"actionable_suggestions"`
}

// GenerateFeedbackSummary calls Google Gemini API to analyze ratings and produce
// an executive summary, category breakdown, and instant actionable improvements.
func (g *GeminiClient) GenerateFeedbackSummary(
	ctx context.Context,
	branchName string,
	analytics *dto.FeedbackAnalytics,
	recentReviews []model.OrderFeedbackAdminItem,
) (*dto.FeedbackAISummaryResponse, error) {
	if g.apiKey == "" {
		return &dto.FeedbackAISummaryResponse{
			Configured: false,
			Model:      g.model,
			Error:      "Google Gemini API Key belum disetel pada konfigurasi backend. Silakan tambahkan GEMINI_API_KEY di file .env backend.",
		}, nil
	}

	prompt := buildGeminiPrompt(branchName, analytics, recentReviews)

	// Attempt generation with configured model, with graceful fallback if model is unavailable
	modelsToTry := []string{g.model}
	if !strings.Contains(g.model, "2.0-flash") && g.model != "gemini-2.0-flash" {
		modelsToTry = append(modelsToTry, "gemini-2.0-flash")
	}
	if !strings.Contains(g.model, "1.5-flash") && g.model != "gemini-1.5-flash" {
		modelsToTry = append(modelsToTry, "gemini-1.5-flash")
	}

	var rawResponse string
	var successfulModel string
	var lastErr error

	for _, modelName := range modelsToTry {
		respText, err := g.callGeminiAPI(ctx, modelName, prompt)
		if err == nil && strings.TrimSpace(respText) != "" {
			rawResponse = respText
			successfulModel = modelName
			break
		}
		lastErr = err
	}

	if rawResponse == "" {
		errMsg := "Gagal menghubungi Google Gemini API."
		if lastErr != nil {
			errMsg = fmt.Sprintf("Gagal menghubungi Google Gemini API: %v", lastErr)
		}
		return &dto.FeedbackAISummaryResponse{
			Configured:  true,
			Model:       g.model,
			Error:       errMsg,
			GeneratedAt: time.Now().Format(time.RFC3339),
		}, nil
	}

	parsed := parseGeminiOutput(rawResponse)
	return &dto.FeedbackAISummaryResponse{
		Configured:            true,
		Model:                 successfulModel,
		ExecutiveSummary:      parsed.ExecutiveSummary,
		DetailSummary:         parsed.DetailSummary,
		ActionableSuggestions: parsed.ActionableSuggestions,
		RawAnalysis:           rawResponse,
		GeneratedAt:           time.Now().Format(time.RFC3339),
	}, nil
}

func (g *GeminiClient) callGeminiAPI(ctx context.Context, modelName, prompt string) (string, error) {
	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		modelName,
		g.apiKey,
	)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.2,
			MaxOutputTokens:  2048,
			ResponseMimeType: "application/json",
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var gResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &gResp); err != nil {
		return "", fmt.Errorf("decode error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	if gResp.Error != nil {
		return "", fmt.Errorf("gemini api error (%d): %s", gResp.Error.Code, gResp.Error.Message)
	}

	if len(gResp.Candidates) == 0 || len(gResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty candidate response from gemini")
	}

	return gResp.Candidates[0].Content.Parts[0].Text, nil
}

func buildGeminiPrompt(branchName string, analytics *dto.FeedbackAnalytics, reviews []model.OrderFeedbackAdminItem) string {
	var sb strings.Builder

	sb.WriteString("Anda adalah konsultan operasional restoran F&B dan pengalaman pelanggan (CX) profesional untuk Mareme Group (outlet kuliner di Surakarta/Solo).\n")
	sb.WriteString("Analisis data rating, ulasan menu, feedback resto, dan feedback aplikasi pelanggan berikut ini.\n\n")

	sb.WriteString(fmt.Sprintf("=== LINGKUP OUTLET: %s ===\n", branchName))
	sb.WriteString(fmt.Sprintf("- Total Ulasan: %d\n", analytics.TotalReviews))
	sb.WriteString(fmt.Sprintf("- Rating Rata-rata Keseluruhan: %.1f / 5.0\n", analytics.AvgOverallRating))
	sb.WriteString(fmt.Sprintf("- Rating Rata-rata Resto: %.1f / 5.0\n", analytics.AvgRestoRating))
	sb.WriteString(fmt.Sprintf("- Rating Rata-rata Aplikasi: %.1f / 5.0\n", analytics.AvgAppRating))
	sb.WriteString(fmt.Sprintf("- Tingkat Kepuasan Pelanggan: %.1f%%\n", analytics.SatisfactionRate))
	sb.WriteString(fmt.Sprintf("- Ulasan Positif (Bintang 4-5): %d | Ulasan Evaluatif (Bintang 1-3): %d\n\n", analytics.PositiveCount, analytics.ConstructiveCount))

	// Rating breakdown
	sb.WriteString("=== DISTRIBUSI BINTANG ===\n")
	for _, r := range analytics.RatingBreakdown {
		sb.WriteString(fmt.Sprintf("- Bintang %d: %d ulasan (%.1f%%)\n", r.Rating, r.Count, r.Percentage))
	}
	sb.WriteString("\n")

	// Menu items performance
	if len(analytics.TopMenuItems) > 0 {
		sb.WriteString("=== MENU TERBAIK (TOP RATED) ===\n")
		for _, m := range analytics.TopMenuItems {
			sb.WriteString(fmt.Sprintf("- %s: Rating %.1f (dari %d review, %d positif)\n", m.ItemName, m.AvgRating, m.TotalReviews, m.PositiveCount))
		}
		sb.WriteString("\n")
	}

	if len(analytics.NeedsAttentionItems) > 0 {
		sb.WriteString("=== MENU MEMERLUKAN PERHATIAN/PERBAIKAN ===\n")
		for _, m := range analytics.NeedsAttentionItems {
			reasons := strings.Join(m.SampleReasons, "; ")
			if reasons == "" {
				reasons = "belum ada catatan spesifik"
			}
			sb.WriteString(fmt.Sprintf("- %s: Rating %.1f (keluhan/evaluasi: %d) -> Alasan: %s\n", m.ItemName, m.AvgRating, m.NegativeCount, reasons))
		}
		sb.WriteString("\n")
	}

	// Common tags
	if len(analytics.CommonTags) > 0 {
		sb.WriteString("=== KATA KUNCI / REASON CHIPS PALING SERING MUNCUL ===\n")
		for _, t := range analytics.CommonTags {
			sentiment := "Positif"
			if !t.IsPositive {
				sentiment = "Perlu Perhatian"
			}
			sb.WriteString(fmt.Sprintf("- [%s] \"%s\" (%dx)\n", sentiment, t.Tag, t.Count))
		}
		sb.WriteString("\n")
	}

	// Recent reviews samples
	sb.WriteString("=== SAMPEL ULASAN TERBARU PELANGGAN ===\n")
	limit := 15
	if len(reviews) < limit {
		limit = len(reviews)
	}
	for i := 0; i < limit; i++ {
		rev := reviews[i]
		sb.WriteString(fmt.Sprintf("%d. Order #%s (%s): Rating %d★ | Resto: %v★ (Alasan: %s) | App: %v★ (Alasan: %s)",
			i+1, rev.OrderNumber, rev.BranchName, rev.Rating,
			derefInt(rev.RestoRating), derefStr(rev.RestoReason),
			derefInt(rev.AppRating), derefStr(rev.AppReason)))
		if rev.Comment != nil && *rev.Comment != "" {
			sb.WriteString(fmt.Sprintf(" | Catatan: \"%s\"", *rev.Comment))
		}
		if len(rev.ItemsFeedback) > 0 {
			itemsStr := make([]string, 0)
			for _, it := range rev.ItemsFeedback {
				rStr := ""
				if it.Reason != nil && *it.Reason != "" {
					rStr = fmt.Sprintf(" (\"%s\")", *it.Reason)
				}
				itemsStr = append(itemsStr, fmt.Sprintf("%s: %d★%s", it.ItemName, it.Rating, rStr))
			}
			sb.WriteString(fmt.Sprintf(" | Menu Items: [%s]", strings.Join(itemsStr, ", ")))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`
TUGAS ANDA:
Berikan analisis mendalam, tajam, dan langsung dapat ditindaklanjuti (actionable) untuk owner dan tim resto.
Jawab HANYA dalam format JSON valid dengan struktur berikut:
{
  "executive_summary": "Ringkasan eksekutif 2-3 paragraf ringkas mengenai kesehatan sentimen pelanggan, tren kepuasan, dan reputasi resto serta aplikasi saat ini.",
  "detail_summary": "Rincian evaluasi terstruktur dalam format markdown yang mencakup 3 aspek utama: 1) Kualitas Makanan & Rasa (apa yang disukai dan keluhan spesifik menu), 2) Operasional & Pengemasan (suhu, kemasan, kecepatan pelayanan), 3) Pengalaman Aplikasi & Pemesanan (kemudahan bayar, kelancaran pelacakan).",
  "actionable_suggestions": [
    "Prioritas 1 (Dapur/Rasa): Rekomendasi konkret langkah perbaikan segera pada menu tertentu atau standar resep/porsi.",
    "Prioritas 2 (Operasional/Kemasan): Rekomendasi konkret mengenai pengemasan, suhu makanan, atau waktu penyiapan.",
    "Prioritas 3 (Aplikasi & Layanan): Rekomendasi konkret perbaikan pada alur pemesanan atau interaksi staf.",
    "Prioritas 4 (Quick Win): Tindakan instan yang bisa langsung dieksekusi hari ini juga untuk meningkatkan skor rating."
  ]
}
`)

	return sb.String()
}

func parseGeminiOutput(raw string) aiOutputJSON {
	var out aiOutputJSON
	cleaned := strings.TrimSpace(raw)

	// Remove markdown code fences if Gemini wrapped it
	if strings.HasPrefix(cleaned, "```") {
		lines := strings.Split(cleaned, "\n")
		if len(lines) >= 2 {
			if strings.HasPrefix(lines[0], "```") {
				lines = lines[1:]
			}
			if len(lines) > 0 && strings.HasPrefix(lines[len(lines)-1], "```") {
				lines = lines[:len(lines)-1]
			}
			cleaned = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}

	if err := json.Unmarshal([]byte(cleaned), &out); err == nil {
		if out.ExecutiveSummary != "" || len(out.ActionableSuggestions) > 0 {
			return out
		}
	}

	// Fallback if unstructured
	return aiOutputJSON{
		ExecutiveSummary: cleaned,
		DetailSummary:    "Analisis detail telah disertakan pada ringkasan di atas.",
		ActionableSuggestions: []string{
			"Pantau konsistensi resep dan suhu penyajian menu teratas.",
			"Pastikan kemasan pesanan takeaway/delivery tertutup rapat.",
			"Lakukan follow up pada ulasan dengan bintang di bawah 4 untuk kepuasan pelanggan.",
		},
	}
}

func derefInt(i *int) string {
	if i == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *i)
}

func derefStr(s *string) string {
	if s == nil || *s == "" {
		return "-"
	}
	return *s
}
