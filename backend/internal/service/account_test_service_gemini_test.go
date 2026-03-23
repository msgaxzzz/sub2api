//go:build unit

package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCreateGeminiTestPayload_ImageModel(t *testing.T) {
	t.Parallel()

	payload := createGeminiTestPayload("gemini-2.5-flash-image", "draw a tiny robot")

	var parsed struct {
		Contents []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"contents"`
		GenerationConfig struct {
			ResponseModalities []string `json:"responseModalities"`
			ImageConfig        struct {
				AspectRatio string `json:"aspectRatio"`
			} `json:"imageConfig"`
		} `json:"generationConfig"`
	}

	require.NoError(t, json.Unmarshal(payload, &parsed))
	require.Len(t, parsed.Contents, 1)
	require.Len(t, parsed.Contents[0].Parts, 1)
	require.Equal(t, "draw a tiny robot", parsed.Contents[0].Parts[0].Text)
	require.Equal(t, []string{"TEXT", "IMAGE"}, parsed.GenerationConfig.ResponseModalities)
	require.Equal(t, "1:1", parsed.GenerationConfig.ImageConfig.AspectRatio)
}

func TestProcessGeminiStream_EmitsImageEvent(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	ctx, recorder := newSoraTestContext()
	svc := &AccountTestService{}

	stream := strings.NewReader("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"ok\"},{\"inlineData\":{\"mimeType\":\"image/png\",\"data\":\"QUJD\"}}]}}]}\n\ndata: [DONE]\n\n")

	err := svc.processGeminiStream(ctx, stream)
	require.NoError(t, err)

	body := recorder.Body.String()
	require.Contains(t, body, "\"type\":\"content\"")
	require.Contains(t, body, "\"text\":\"ok\"")
	require.Contains(t, body, "\"type\":\"image\"")
	require.Contains(t, body, "\"image_url\":\"data:image/png;base64,QUJD\"")
	require.Contains(t, body, "\"mime_type\":\"image/png\"")
}

func TestProcessGeminiResponse_EmitsImageEvent(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	ctx, recorder := newSoraTestContext()
	svc := &AccountTestService{}

	resp := strings.NewReader("{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"ok\"},{\"inlineData\":{\"mimeType\":\"image/png\",\"data\":\"QUJD\"}}]}}]}")

	err := svc.processGeminiResponse(ctx, resp)
	require.NoError(t, err)

	body := recorder.Body.String()
	require.Contains(t, body, "\"type\":\"content\"")
	require.Contains(t, body, "\"text\":\"ok\"")
	require.Contains(t, body, "\"type\":\"image\"")
	require.Contains(t, body, "\"image_url\":\"data:image/png;base64,QUJD\"")
	require.Contains(t, body, "\"mime_type\":\"image/png\"")
	require.Contains(t, body, "\"type\":\"test_complete\"")
}

func TestNormalizeGeminiNativeTestModelID_LegacyAliases(t *testing.T) {
	t.Parallel()

	require.Equal(t, "gemini-3.1-flash-image-preview", normalizeGeminiNativeTestModelID("gemini-3.1-flash-image"))
	require.Equal(t, "gemini-3-pro-image-preview", normalizeGeminiNativeTestModelID("gemini-3-pro-image"))
	require.Equal(t, "gemini-2.5-flash-image", normalizeGeminiNativeTestModelID("gemini-2.5-flash-image"))
}

func TestShouldForceGeminiCodeAssistStream(t *testing.T) {
	t.Parallel()

	require.True(t, shouldForceGeminiCodeAssistStream("gemini-2.5-pro"))
	require.False(t, shouldForceGeminiCodeAssistStream("gemini-2.5-flash-image"))
	require.False(t, shouldForceGeminiCodeAssistStream("gemini-3.1-flash-image-preview"))
}

func TestGeminiImagePreviewFallbackModel(t *testing.T) {
	t.Parallel()

	modelID, ok := geminiImagePreviewFallbackModel("gemini-2.5-flash-image")
	require.True(t, ok)
	require.Equal(t, "gemini-2.5-flash-image-preview", modelID)

	modelID, ok = geminiImagePreviewFallbackModel("gemini-3.1-flash-image")
	require.True(t, ok)
	require.Equal(t, "gemini-3.1-flash-image-preview", modelID)

	_, ok = geminiImagePreviewFallbackModel("gemini-2.5-flash")
	require.False(t, ok)
}

func TestTestGeminiAccountConnection_CodeAssistImageModelFailsFast(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	ctx, recorder := newSoraTestContext()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	ctx.Request = req

	svc := &AccountTestService{}
	account := &Account{
		ID:       1,
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"project_id":    "project-1",
			"access_token":  "token",
			"refresh_token": "refresh",
		},
	}

	err := svc.testGeminiAccountConnection(ctx, account, "gemini-2.5-flash-image", "draw a tiny robot")
	require.NoError(t, err)

	body := recorder.Body.String()
	require.Contains(t, body, "Gemini OAuth mode: Code Assist")
	require.Contains(t, body, "do not support native image model tests here")
}
