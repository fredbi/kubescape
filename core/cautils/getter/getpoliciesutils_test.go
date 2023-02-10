package getter

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultPath(t *testing.T) {
	t.Parallel()

	const name = "mine"

	pth := GetDefaultPath(name)
	require.Equal(t, name, filepath.Base(pth))
	require.Equal(t, ".kubescape", filepath.Base(filepath.Dir(pth)))
}

func TestSaveInFile(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp(".", "test")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	policy := map[string]interface{}{
		"key":    "value",
		"number": 1.00,
	}

	t.Run("should save data as JSON (target folder exists)", func(t *testing.T) {
		target := filepath.Join(dir, "target.json")
		require.NoError(t, SaveInFile(policy, target))

		buf, err := os.ReadFile(target)
		require.NoError(t, err)
		var retrieved interface{}
		require.NoError(t, jsoniter.Unmarshal(buf, &retrieved))

		require.EqualValues(t, policy, retrieved)
	})

	t.Run("should save data as JSON (new target folder)", func(t *testing.T) {
		target := filepath.Join(dir, "subdir", "target.json")
		require.NoError(t, SaveInFile(policy, target))

		buf, err := os.ReadFile(target)
		require.NoError(t, err)
		var retrieved interface{}
		require.NoError(t, jsoniter.Unmarshal(buf, &retrieved))

		require.EqualValues(t, policy, retrieved)
	})

	t.Run("should error", func(t *testing.T) {
		badPolicy := map[string]interface{}{
			"key":    "value",
			"number": 1.00,
			"err":    func() {},
		}
		target := filepath.Join(dir, "error.json")
		require.Error(t, SaveInFile(badPolicy, target))
	})
}

func TestHttpMethods(t *testing.T) {
	client := http.DefaultClient
	hdrs := map[string]string{"key": "value"}

	srv := mockAPIServer(t)
	t.Cleanup(srv.Close)

	t.Run("HttpGetter should GET", func(t *testing.T) {
		resp, err := HttpGetter(client, srv.URL(pathTestGet), hdrs)
		require.NoError(t, err)
		require.EqualValues(t, "body-get", resp)
	})
}
