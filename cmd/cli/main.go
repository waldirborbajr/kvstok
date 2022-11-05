package main

import (
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"log"
	"os"

	"github.com/waldirborbajr/kvstok/cmd"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
)

var hasPub = true
var hasPriv = true
var privateKey rsa.PrivateKey
var err error

func main() {

	home := kvpath.GetKVHomeDir()

	pub := home + "/.config/kvstok.pub"
	priv := home + "/.config/kvstok.priv"

	if _, err := os.Stat(pub); err != nil {
		hasPub = false
	}

	if _, err := os.Stat(priv); err != nil {
		hasPriv = false
	}

	// Generete new PRIV/PUB RSA Key
	if !hasPub && !hasPriv {
		privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			log.Fatal("Error GenerateKey: ", err.Error())
		}
	}

	publicKey := privateKey.PublicKey

	_ = ioutil.WriteFile(pub, []byte(publicKey), 0o644)
	_ = ioutil.WriteFile(priv, []byte(privateKey), 0o644)

	cmd.Execute()
}
