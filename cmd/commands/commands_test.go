package commands

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/nutsdb/nutsdb"
	"github.com/stretchr/testify/require"
	"github.com/waldirborbajr/kvstok/internal/database"
)

func TestAddGetDelCommands(t *testing.T) {
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	tmpHome := t.TempDir()
	require.NoError(t, os.Setenv("HOME", tmpHome))

	store, err := database.Init("")
	require.NoError(t, err)

	require.NoError(t, store.DB().Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureBTree, database.Bucket)
	}))
	require.NoError(t, store.SetMasterPassword("cli-password"))
	require.NoError(t, store.Close())

	err = runAdd(nil, []string{"cli-key", "cli-value"})
	require.NoError(t, err)
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	err = runGet(nil, []string{"cli-key"})
	require.NoError(t, err)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "cli-value")

	store, err = database.Init("")
	require.NoError(t, err)
	defer database.Close()

	DelCmd.Run(nil, []string{"cli-key"})

	_, err = store.Get("cli-key")
	require.ErrorIs(t, err, database.ErrKeyNotFound)
}
