package database

import (
	"testing"

	"github.com/xujiajun/nutsdb"
)

func TestXxx(t *testing.T) {
	_, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir("./"),
		nutsdb.WithSegmentSize(512*512),
	)

	if err != nil {
		t.Fatal("Error")
	}

}
