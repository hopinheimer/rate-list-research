package main

import (
	"os"
	"io"
	"bytes"
	"log"
	"encoding/json"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type KeyRing struct {
  priv crypto.PrivKey
  pub crypto.PubKey
  id peer.ID
}

func main() {

	// create a key pair for the peer's identity
	priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if (err != nil) {
		panic(err)
	}

	// calculate the id from the pubkey
	id, err := peer.IDFromPublicKey(pub)
	if (err != nil) {
		panic(err)
	}

	key_ring := &KeyRing{
		priv: priv,
		pub: pub,
		id: id,
	}

	var Marshal = func(v interface{}) (io.Reader, error) {
		b, err := json.MarshalIndent(v, "", "\t")
		if err != nil {
			return nil, err
		}
		
		return bytes.NewReader(b), nil
	}

	var Save = func (path string, v interface{}) error {
  		f, err := os.Create(path)
  		if err != nil {
    		return err
  		}
  		defer f.Close()
  		
		r, err := Marshal(v)
  		if err != nil {
    		return err
  		}
  
		_, err = io.Copy(f, r)
  		return err
	}

	if  err := Save("./identities/" + os.Args[1] + ".json", key_ring); err != nil {
		log.Fatalln(err)
	}
}
