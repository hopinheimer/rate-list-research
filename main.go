package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
	"github.com/libp2p/go-libp2p/core/peerstore"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var dht *kaddht.IpfsDHT
	// define a function to create a DHT node because libp2p accepts only a routing constructor. Additionally, the function initializes a variable in the outer scope
	createDHTNode := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		dht, err = kaddht.New(ctx, h)
		return dht, err
	}
	
	// create a key pair for the peer's identity
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)

	host, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"), // be reachable at all ipv4 address of the local machine on a random port
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport), // streams will be created with the default transport (TCP)
		libp2p.Security(tls.ID, tls.New),
		libp2p.Routing(createDHTNode)
	)
	defer host.close()
	if err != nil {
		panic(err)
	}

	// TODO: connect to bootstrap peers here

	err = dht.Bootstrap(ctx)
	if err != nil {
		panic(err)
	}

	// routing discovery implements discovering peers through provider records
	// TODO: analyse if using the discovery method makes sense, since for a particular 
	// stream protocol only one subset of nodes will be handling the provider records
	routingDiscovery := drouting.NewRoutingDiscovery(dht)
	// util functions implementing timeouts and error handling over the discovery
	dutil.Advertise(ctx, routingDiscovery, "ratedlist/0.0.1")
	peers, err := dutil.FindPeers(ctx, routingDiscovery, "ratedlist/0.0.1")
	if err != nil {
		panic(err)
	}
	
	// TODO: discover peers on trigger (whenever peer count is dropping)
	for _, peer := range peers {
		// add the discovered peers to the DHT
		ph.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.PermanentAddrTTL)
	}
	
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT)

	<-stop
	fmt.Println("Received signal, shutting down...")
}
