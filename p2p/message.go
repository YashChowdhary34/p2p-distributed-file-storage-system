package p2p

import "net"

// RPC represens any arbitrary data tha is being sent over each transport between two nodes in the network
type RPC struct {
	From					net.Addr
	payload				[]byte
}