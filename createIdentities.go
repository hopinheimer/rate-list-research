package main

import (
	"os"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"encoding/hex"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type KeyRing struct {
	Priv string `json:"priv"`
	Pub string `json:"pub"`
	Id peer.ID `json:"id"` 
}

func createIdentities(filename string) {

	// create a key pair for the peer's identity
	priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if (err != nil) {
		panic(err)
	}

	rawPrivKey, _ := priv.Raw()
	rawPubKey, _ := pub.Raw()
	rawPrivKeyString := hex.EncodeToString(rawPrivKey)
	rawPubKeyString := hex.EncodeToString(rawPubKey)
	
	// fmt.Println(rawPrivKeyString)
	// fmt.Println(rawPubKeyString)

	// calculate the id from the pubkey
	id, err := peer.IDFromPublicKey(pub)
	if (err != nil) {
		panic(err)
	}

	// fmt.Println(id)

	key_ring := KeyRing{
		Priv: rawPrivKeyString,
		Pub: rawPubKeyString,
		Id: id,
	}

	file, err := json.Marshal(key_ring)
	if (err != nil) {
		panic(err)
	}

	_ = ioutil.WriteFile("./identities/" + filename + ".json", file, 0644)	
}

func main() {
	for _, name := range os.Args[1:] {
		fmt.Println(name)
		createIdentities(name)
	}
}
