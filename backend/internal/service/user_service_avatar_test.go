package service

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAvatarURL(t *testing.T) {
	t.Run("empty clears avatar", func(t *testing.T) {
		got, err := normalizeAvatarURL("")
		require.NoError(t, err)
		require.Equal(t, "", got)
	})

	t.Run("valid jpeg data url passes", func(t *testing.T) {
		payload := base64.StdEncoding.EncodeToString([]byte("jpeg-avatar"))
		got, err := normalizeAvatarURL("data:image/jpeg;base64," + payload)
		require.NoError(t, err)
		require.Contains(t, got, payload)
	})

	t.Run("invalid type rejected", func(t *testing.T) {
		_, err := normalizeAvatarURL("data:image/gif;base64,R0lGODlhAQABAIAAAAUEBA==")
		require.Error(t, err)
	})

	t.Run("oversized payload rejected", func(t *testing.T) {
		largePayload := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", maxAvatarImageBytes+1)))
		_, err := normalizeAvatarURL("data:image/png;base64," + largePayload)
		require.Error(t, err)
	})
}
