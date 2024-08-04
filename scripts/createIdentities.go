package main

import (
	"os"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func main() {
	for _, arg := range os.Args[1:] {
		priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
		if (err != nil) {
			panic(err)
		}

		err = StoreIdentity(priv, arg)
		if (err != nil) {
			panic(err)
		}
	}
}


