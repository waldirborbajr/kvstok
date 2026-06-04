package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nutsdb/nutsdb"
	"github.com/waldirborbajr/kvstok/internal/database"
)

func main() {
	tmp := filepath.Join(os.TempDir(), "kvstok_cli_test")
	os.RemoveAll(tmp)
	os.MkdirAll(tmp, 0700)
	os.Setenv("HOME", tmp)

	s, err := database.Init("")
	if err != nil {
		panic(err)
	}
	if err := s.DB().Update(func(tx *nutsdb.Tx) error { return tx.NewBucket(nutsdb.DataStructureBTree, database.Bucket) }); err != nil {
		panic(err)
	}
	if err := s.SetMasterPassword("cli-password"); err != nil {
		panic(err)
	}
	if err := database.Close(); err != nil {
		panic(err)
	}

	st, err := database.GetStore()
	if err != nil {
		panic(err)
	}
	if err := st.Close(); err != nil {
		panic(err)
	}

	st2, err := database.Init("")
	if err != nil {
		panic(err)
	}
	fmt.Println("open2 ok")
	if err := database.Close(); err != nil {
		panic(err)
	}
}
