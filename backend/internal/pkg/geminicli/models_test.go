package geminicli

import "testing"

func TestDefaultModels_ContainsImageModels(t *testing.T) {
	t.Parallel()

	byID := make(map[string]Model, len(DefaultModels))
	for _, model := range DefaultModels {
		byID[model.ID] = model
	}

	required := []string{
		"gemini-2.5-flash-image",
		"gemini-3-pro-image-preview",
	}

	for _, id := range required {
		if _, ok := byID[id]; !ok {
			t.Fatalf("expected curated Gemini model %q to exist", id)
		}
	}
}

func TestDefaultOAuthFallbackModels_ExcludeImageModels(t *testing.T) {
	t.Parallel()

	for _, model := range DefaultOAuthFallbackModels {
		if model.ID == "gemini-2.5-flash-image" || model.ID == "gemini-3-pro-image-preview" {
			t.Fatalf("expected oauth fallback model list to exclude image model %q", model.ID)
		}
	}
}
