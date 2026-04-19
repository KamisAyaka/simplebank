package util

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig_LoadsRefreshTokenDuration(t *testing.T) {
	dir := t.TempDir()
	content := `DB_DRIVER=postgres
DB_SOURCE=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable
SERVER_ADDRESS=0.0.0.0:8080
TOKEN_SYMMETRIC_KEY=12345678901234567890123456789012
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=24h
`

	configPath := filepath.Join(dir, "app.env")
	err := os.WriteFile(configPath, []byte(content), 0o600)
	require.NoError(t, err)

	config, err := LoadConfig(dir)
	require.NoError(t, err)
	require.Equal(t, 15*time.Minute, config.AccessTokenDuration)
	require.Equal(t, 24*time.Hour, config.RefreshTokenDuration)
}
