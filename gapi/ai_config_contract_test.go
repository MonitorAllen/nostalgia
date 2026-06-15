package gapi

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAIConfigRuntimeMutationContracts(t *testing.T) {
	protoSource, err := os.ReadFile("../proto/rpc_polish_text.proto")
	require.NoError(t, err)
	serviceSource, err := os.ReadFile("../proto/service_nostalgia.proto")
	require.NoError(t, err)
	querySource, err := os.ReadFile("../db/query/ai_provider_config.sql")
	require.NoError(t, err)
	migrationSource, err := os.ReadFile("../db/migration/000011_add_ai_provider_configs.up.sql")
	require.NoError(t, err)

	require.Contains(t, string(protoSource), "message UpdateAIConfigRequest")
	require.Contains(t, string(protoSource), "string api_key =")
	require.Contains(t, string(protoSource), "bool clear_api_key =")
	require.Contains(t, string(serviceSource), "rpc UpdateAIConfig")
	require.Contains(t, string(serviceSource), `patch: "/v1/ai/config"`)
	require.Contains(t, string(querySource), "GetAIProviderConfig")
	require.Contains(t, string(querySource), "UpsertAIProviderConfig")
	require.Contains(t, string(migrationSource), "CREATE TABLE ai_provider_configs")
	require.Contains(t, string(migrationSource), "api_key_ciphertext")
}
