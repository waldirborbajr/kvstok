package database

import (
	"github.com/xujiajun/nutsdb"
)

const (
	DBName = ".6B7673" // -> .kvs
	Bucket = "kvstok"
)

// Reference of database
var DB *nutsdb.DB
