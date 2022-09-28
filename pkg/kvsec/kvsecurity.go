package kvsec

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const md5FileName = ".md5sum"

func GenMD5Sum(kvFileStoreName string) string {
	file, err := os.Open(kvFileStoreName)
	if err != nil {
		fmt.Printf("Error reading data store file %s", err.Error())
		os.Exit(-1)
	}

	defer file.Close()

	hash := md5.New()

	if _, err := io.Copy(hash, file); err != nil {
		fmt.Printf("Error %s", err.Error())
	}

	md5sum := hex.EncodeToString(hash.Sum(nil))

	return md5sum
}

func SaveMD5(md5sum string) {
	file, err := os.Create(md5FileName)
	if err != nil {
		fmt.Printf("Error creating file %s", err.Error())
		os.Exit(-1)
	}

	fmt.Printf("MD5File %s", string(md5sum))

	_, err = file.WriteString(string(md5sum))
	if err != nil {
		fmt.Printf("Errow wrinting file %s", err.Error())
		os.Exit(-1)
	}
}

func CompareMD5(kvFileStoreName string) {
	file, err := os.Open(md5FileName)
	if err != nil {
		fmt.Printf("Error opening md5sum file %s", err.Error())
		os.Exit(-1)
	}

	line := bufio.NewScanner(file)
	line.Split(bufio.ScanLines)

	var contentLine string
	for line.Scan() {
		contentLine = line.Text()
	}

	md5sum := GenMD5Sum(kvFileStoreName)

	if md5sum != contentLine {
		SaveMD5(md5sum)
	}
}
