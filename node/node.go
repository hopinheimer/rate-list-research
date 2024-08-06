package node

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	dcrypto "github.com/chirag-parmar/rated-list/crypto"
	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/routing"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	P2PTopic = "ratedlist/0.0.1"
)

type Node struct {
	dht   *kaddht.IpfsDHT
	dhost host.Host
	// IsBootstrapped maintains is the flag the node is bootstrapped or yet to be bootstrapped
	IsBootstraped bool
	// bnode store bootnode name
	bnode            *string
	routingDiscovery *drouting.DiscoveryRouting
	dctx             *context.Context
	lock             sync.Mutex
	bootstrapPeers   []peer.AddrInfo
	privKey          crypto.PrivKey
}

func New(ctx context.Context, bootnode string, create bool) (*Node, error) {

	dnode := &Node{}
	dnode.dctx = &ctx

	dnode.bootstrapPeers = BOOTSTRAP_PEERS

	createDHTNode := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		dnode.dht, err = kaddht.New(ctx, h)
		return dnode.dht, err
	}

	var err error

	if bootnode != "" {
		if create {
			dnode.privKey, _, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
			err = dcrypto.StoreIdentity(dnode.privKey, bootnode)
		} else {
			dnode.privKey, err = dcrypto.LoadIdentity(bootnode)
		}
		dnode.bnode = &bootnode
	} else {
		dnode.privKey, _, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
	}

	host, err := libp2p.New(
		libp2p.Identity(dnode.privKey),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/7878"),
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.Security(tls.ID, tls.New),
		libp2p.Routing(createDHTNode),
	)

	if err != nil {
		log.Fatal("critical failure in node creation")
	}

	dnode.dhost = host
	dnode.dctx = &ctx

	return dnode, nil
}

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

func (n *Node) Start() {

	var err error

	if n.bnode == nil {
		n.Bootstrap()
	}

	routingDiscovery := drouting.NewRoutingDiscovery(n.dht)

	dutil.Advertise(*n.dctx, routingDiscovery, P2PTopic)
	n.bootstrapPeers, err = dutil.FindPeers(*n.dctx, routingDiscovery, P2PTopic)
	if err != nil {
		panic(err)
	}

	for _, peer := range n.bootstrapPeers {
		fmt.Println("Found peer with multiaddress: ", peer.Addrs)

		n.dhost.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.PermanentAddrTTL)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT)

	<-stop
	fmt.Println("Received signal, shutting down...")
}

func (n *Node) Bootstrap() error {
	peers := n.bootstrapPeers
	if len(peers) < 1 {
		return errors.New("not enough bootstrap peers")
	}

	errs := make(chan error, len(peers))
	var wg sync.WaitGroup
	for _, p := range peers {

		wg.Add(1)
		go func(p peer.AddrInfo) {
			defer wg.Done()
			defer log.Println(n.dctx, "bootstrapDial", n.dhost.ID(), p.Addrs)
			log.Printf("%s bootstrapping to %s", n.dhost.ID(), p.Addrs)

			n.dhost.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL) // connect adds the peers to peerstore but only temporarily hence we add it permanently
			if err := n.dhost.Connect(*n.dctx, p); err != nil {
				log.Println(n.dctx, "bootstrapDialFailed", p.Addrs)
				log.Printf("failed to bootstrap with %v: %s", p.Addrs, err)
				errs <- err
				return
			}

			log.Println(n.dctx, "bootstrapDialSuccess", p.ID)
			log.Printf("bootstrapped with %v", p.ID)
		}(p)
	}
	wg.Wait()

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

func (n *Node) InformDHTBootstrap() error {
	n.dht.Bootstrap(*n.dctx)
	return nil
}
