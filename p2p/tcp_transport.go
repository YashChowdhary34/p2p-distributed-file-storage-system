package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
)

// public/priority functions on top and private/low-priority functions on the bottom

// TCPPeer represents the remote node over a TCP estabilished connection
type TCPPeer struct {
	// conn is the underlying connection of the peer
	conn					net.Conn

	//if we dial and retrieve a conn => outbound == true
	//if we accept and retrieve a conn => outbound == false
	outbound 			bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn: conn,
		outbound: outbound,
	}
}

//Close implements the Peer interface
func (p *TCPPeer) close() error {
	return p.conn.Close()
}

type TCPTransportOpts struct {
	ListenAddr 		string
	HandshakeFunc	HandshakeFunc
	Decoder				Decoder
	OnPeer 				func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener       net.Listener
	rpcch					chan RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch: make(chan RPC),
	}
}

// consume implements the Transport interface, which will return read-only channel for reading the incoming messages received from another peer in the network
// we can read from the channel not write to the channel
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// close implements the transport interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if (err != nil) {
		return err
	}

	go t.startAcceptLoop()

	log.Printf("TCP transport listening on port: %s\n", t.ListenAddr)
	
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if (err != nil) {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		
		fmt.Printf("New incoming connection %+v\n", conn) // +v also includes field names in output

		go t.handleConn(conn)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error

	defer func() {
		fmt.Printf("Dropping peer connection: %s", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)

	if err = t.HandshakeFunc(peer); err != nil {
		return 
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	//read loop
	rpc := RPC{}
	for {
		if err := t.Decoder.Decode(conn, &rpc); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}

		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}

	
}