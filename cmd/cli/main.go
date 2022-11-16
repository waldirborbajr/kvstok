package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/waldirborbajr/kvstok/cmd"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	security "github.com/waldirborbajr/kvstok/internal/secutiry"
)

var hasPub = true
var hasPriv = true

func main() {

	home := kvpath.GetKVHomeDir()

	pub := home + "/.config/kvstok/kvstok.pub"
	priv := home + "/.config/kvstok/kvstok.priv"

	if _, err := os.Stat(pub); err != nil {
		hasPub = false
	}

	if _, err := os.Stat(priv); err != nil {
		hasPriv = false
	}

	// Generete new PRIV/PUB RSA Key
	if !hasPub && !hasPriv {

		fmt.Println("Generating RSA priv/pub key pairing")
		privateKey, publicKey := security.RSA_GenerateKey(4096)

		_ = ioutil.WriteFile(pub, []byte(security.PublicKeyToBytes(publicKey)), 0600)
		_ = ioutil.WriteFile(priv, []byte(security.PrivateKeyToBytes(privateKey)), 0600)
	}

	cmd.Execute()
}
