package database

import (
	"testing"
	"time"

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

func TestStoreGetExpiredKey(t *testing.T) {
	store := createTestStore(t)
	defer store.Close()

	require.NoError(t, store.SetMasterPassword("secret123"))
	require.NoError(t, store.Put("expiring", "value", 1, nil))

	time.Sleep(2 * time.Second)

	_, err := store.Get("expiring")
	require.ErrorIs(t, err, ErrKeyNotFound)
}

func TestStoreGetRawExpiredKeyDeletes(t *testing.T) {
	store := createTestStore(t)
	defer store.Close()

	require.NoError(t, store.SetMasterPassword("secret123"))
	require.NoError(t, store.Put("expiring", "value", 1, nil))

	time.Sleep(2 * time.Second)

	_, _, err := store.GetRaw("expiring")
	require.ErrorIs(t, err, ErrKeyNotFound)

	_, err = store.Get("expiring")
	require.ErrorIs(t, err, ErrKeyNotFound)
}

func TestStoreWrongMasterPasswordFails(t *testing.T) {
	store := createTestStore(t)
	defer store.Close()

	require.NoError(t, store.SetMasterPassword("correct-password"))
	require.NoError(t, store.Put("secret-key", "secret-value", 0, nil))

	storePath := store.dbPath
	require.NoError(t, store.Close())

	otherStore, err := NewStore(storePath)
	require.NoError(t, err)
	defer otherStore.Close()
	require.NoError(t, otherStore.LoadMasterSalt())
	require.NoError(t, otherStore.SetMasterPassword("wrong-password"))

	_, err = otherStore.Get("secret-key")
	require.Error(t, err)
	require.Contains(t, err.Error(), "decryption failed")
}

func TestStoreChangeMasterPassword(t *testing.T) {
	store := createTestStore(t)
	defer store.Close()

	require.NoError(t, store.SetMasterPassword("old-password"))
	require.NoError(t, store.Put("secret-key", "secret-value", 0, []string{"important"}))

	require.NoError(t, store.ChangeMasterPassword("old-password", "new-password"))

	storePath := store.dbPath
	require.NoError(t, store.Close())

	newStore, err := NewStore(storePath)
	require.NoError(t, err)
	defer newStore.Close()

	require.NoError(t, newStore.LoadMasterSalt())
	require.NoError(t, newStore.SetMasterPassword("new-password"))

	value, err := newStore.Get("secret-key")
	require.NoError(t, err)
	require.Equal(t, "secret-value", value)

	require.NoError(t, newStore.Close())

	oldStore, err := NewStore(storePath)
	require.NoError(t, err)
	defer oldStore.Close()

	require.NoError(t, oldStore.LoadMasterSalt())
	require.NoError(t, oldStore.SetMasterPassword("old-password"))

	_, err = oldStore.Get("secret-key")
	require.Error(t, err)
	require.Contains(t, err.Error(), "decryption failed")
}
