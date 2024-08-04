package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	ma "github.com/multiformats/go-multiaddr"
)

// Source: https://github.com/libp2p/go-libp2p/blob/master/examples/routed-echo/bootstrap.go
var (
	BOOTSTRAP_PEERS = convertPeers([]string{
		"/ip4/10.0.0.2/tcp/7878/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	})
)

// Source: https://github.com/libp2p/go-libp2p/blob/master/examples/routed-echo/bootstrap.go
func convertPeers(peers []string) []peer.AddrInfo {
	pinfos := make([]peer.AddrInfo, len(peers))
	for i, addr := range peers {
		maddr := ma.StringCast(addr)
		p, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.Fatalln(err)
		}
		pinfos[i] = *p
	}
	return pinfos
}

// Source: https://github.com/libp2p/go-libp2p/blob/master/examples/routed-echo/bootstrap.go
// This code is borrowed from the go-ipfs bootstrap process
func bootstrapConnect(ctx context.Context, ph host.Host, peers []peer.AddrInfo) error {
	if len(peers) < 1 {
		return errors.New("not enough bootstrap peers")
	}

	errs := make(chan error, len(peers))
	var wg sync.WaitGroup
	for _, p := range peers {

		wg.Add(1)
		go func(p peer.AddrInfo) {
			defer wg.Done()
			defer log.Println(ctx, "bootstrapDial", ph.ID(), p.Addrs)
			log.Printf("%s bootstrapping to %s", ph.ID(), p.Addrs)

			ph.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL) // connect adds the peers to peerstore but only temporarily hence we add it permanently
			if err := ph.Connect(ctx, p); err != nil {
				log.Println(ctx, "bootstrapDialFailed", p.Addrs)
				log.Printf("failed to bootstrap with %v: %s", p.Addrs, err)
				errs <- err
				return
			}

			log.Println(ctx, "bootstrapDialSuccess", p.ID)
			log.Printf("bootstrapped with %v", p.ID)
		}(p)
	}
	wg.Wait()

	// our failure condition is when no connection attempt succeeded.
	// So drain the errs channel, counting the results.
	close(errs)
	count := 0
	var err error
	for err = range errs {
		if err != nil {
			count++
		}
	}
	if count == len(peers) {
		return fmt.Errorf("failed to bootstrap. %s", err)
	}
	return nil
}
