package service

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShouldRetryOpenAIPoolModeOnSameAccount(t *testing.T) {
	account := &Account{
		Type: AccountTypeAPIKey,
		Credentials: map[string]any{
			"pool_mode": true,
		},
	}

	t.Run("permanent 401 is not retried on same account", func(t *testing.T) {
		body := []byte(`{"error":{"message":"Your OpenAI account has been deactivated, please check your email for more information.","code":"account_deactivated"}}`)

		retryable := shouldRetryOpenAIPoolModeOnSameAccount(account, http.StatusUnauthorized, "Your OpenAI account has been deactivated, please check your email for more information.", body)

		require.False(t, retryable)
	})

	t.Run("non-permanent 401 remains retryable in pool mode", func(t *testing.T) {
		body := []byte(`{"error":{"message":"Authentication required","code":"invalid_api_key"}}`)

		retryable := shouldRetryOpenAIPoolModeOnSameAccount(account, http.StatusUnauthorized, "Authentication required", body)

		require.True(t, retryable)
	})

	t.Run("429 remains retryable in pool mode", func(t *testing.T) {
		retryable := shouldRetryOpenAIPoolModeOnSameAccount(account, http.StatusTooManyRequests, "Rate limited", nil)

		require.True(t, retryable)
	})
}
