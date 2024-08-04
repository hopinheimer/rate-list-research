package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"flag"

	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/core/crypto"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/core/peerstore"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		bootnode string
		create bool
	)

	flag.StringVar(&bootnode, "bootnode", "", "bootnode file name")
	flag.BoolVar(&create, "create", false, "create a new bootnode (wont have any affect if bootnode is not specified")
	
	var dht *kaddht.IpfsDHT
	// define a function to create a DHT node because libp2p accepts only a routing constructor. Additionally, the function initializes a variable in the outer scope
	createDHTNode := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		dht, err = kaddht.New(ctx, h)
		return dht, err
	}
	
	// create a key pair for the peer's identity
	var priv crypto.PrivKey
	var err error
	// TODO: use a library to define CLI options
	if bootnode != "" {
		if create {
			priv, _, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
			err = StoreIdentity(priv, bootnode)
		} else {
			priv, err = LoadIdentity(bootnode)
		}
	} else {
		priv, _, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
	}

	if (err != nil) {
		panic(err)
	}

	host, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/7878"), // be reachable at all ipv4 address of the local machine
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport), // streams will be created with the default transport (TCP)
		libp2p.Security(tls.ID, tls.New),
		libp2p.Routing(createDHTNode),
		// TODO: ideally we should also include a peerstore that is loaded from the disk after a restart
		// right now a empty peerstore is created on every restar
	)
	defer host.Close()

	if err != nil {
		panic(err)
	}

	// connect to the bootstrap nodes if no identity file was provided
	// an identity file is provided for nodes that need to have deterministic node ids
	// currently these nodes are expected to be bootstrapped nodes
	// TODO: use a different mechanism than the above
	if bootnode != "" {
		err = bootstrapConnect(ctx, host, BOOTSTRAP_PEERS)
		if err != nil {
			panic(err)
		}
	}
	
	// just tell the DHT that you have bootstrapped
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
		fmt.Println("Found peer with multiaddress: ", peer.Addrs)
		// add the discovered peers to the DHT
		host.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.PermanentAddrTTL)
	}
	
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT)

	<-stop
	fmt.Println("Received signal, shutting down...")
}
