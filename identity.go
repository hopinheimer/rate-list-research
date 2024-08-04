package main

import (
	"os"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func LoadIdentity(filename string) (crypto.PrivKey, error) {
	privBytes, err := os.ReadFile("./identities/" + filename + ".priv")
	if err != nil {
		return nil, err
	}
		
	priv, err := crypto.UnmarshalPrivateKey(privBytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func StoreIdentity(privKey crypto.PrivKey, filename string) (error) {

	privb, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return err
	}
	os.WriteFile("./identities/" + filename + ".priv", privb, 0666)

	
	// calculate the id from the pubkey
	id, err := peer.IDFromPrivateKey(privKey)
	if (err != nil) {
		return err
	}

	// fmt.Println(id)

	os.WriteFile("./identities/" + filename + ".id", []byte(id.String()), 0644)	
	return nil
}

