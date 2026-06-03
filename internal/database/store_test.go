package database

import (
	"testing"

	"github.com/nutsdb/nutsdb"
	"github.com/stretchr/testify/require"
)

func createTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := NewStore(t.TempDir())
	require.NoError(t, err)

	err = store.DB().Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureBTree, Bucket)
	})
	require.NoError(t, err)

	return store
}

func TestStorePutGetRoundTrip(t *testing.T) {
	store := createTestStore(t)
	defer store.Close()

	require.NoError(t, store.SetMasterPassword("secret123"))
	require.NoError(t, store.Put("example", "value123", 0, []string{"tag1", "tag2"}))

	got, err := store.Get("example")
	require.NoError(t, err)
	require.Equal(t, "value123", got)

	value, entry, err := store.GetRaw("example")
	require.NoError(t, err)
	require.Equal(t, "value123", value)
	require.NotNil(t, entry)
	require.Equal(t, uint32(0), entry.TTL)
	require.ElementsMatch(t, []string{"tag1", "tag2"}, entry.Tags)
}

func TestStorePutEmptyValueFails(t *testing.T) {
	store := createTestStore(t)
	defer store.Close()

	require.NoError(t, store.SetMasterPassword("secret123"))
	require.Error(t, store.Put("example", "", 0, nil))
}
