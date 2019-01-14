package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
)

// PeerAddress struct
type PeerAddress struct {
	IP   net.IP
	Port int64
}

// PeerAddresses struct
type PeerAddresses struct {
	Addresses []PeerAddress
	mux       sync.Mutex
}

func EmptyAdresses() *PeerAddresses {
	return &PeerAddresses{
		Addresses: []PeerAddress{},
	}
}

// GetPeerAddress returns a PeerAdress
// from an string
func GetPeerAddress(value string) (PeerAddress, error) {
	var address PeerAddress
	return address, address.Set(value)
}

// String method
func (address *PeerAddress) String() string {
	return fmt.Sprint(address.IP.String(), ":", address.Port)
}

// Set PeerAddress from string
func (address *PeerAddress) Set(value string) error {
	ipPortStr := strings.Split(value, ":")
	if len(ipPortStr) != 2 {
		return errors.New(value + " format not valid")
	}
	if parsedIP := net.ParseIP(ipPortStr[0]); parsedIP == nil {
		return errors.New(value + " IP was not parsed correctly")
	} else if parsedPort, err := strconv.ParseInt(ipPortStr[1], 10, 0); err != nil {
		return err
	} else {
		ad := PeerAddress{IP: parsedIP, Port: parsedPort}
		*address = ad
	}
	return nil
}

func (peers *PeerAddresses) String() string {
	peers.mux.Lock()
	defer peers.mux.Unlock()

	var s []string
	for _, peer := range peers.Addresses {
		s = append(s, peer.String())
	}
	return strings.Join(s, ",")
}

// GetAdresses func
func (peers *PeerAddresses) GetAdresses() []PeerAddress {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	return peers.Addresses
}

func (peers *PeerAddresses) appendPeer(address *PeerAddress) {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	peers.Addresses = append(peers.Addresses, *address)
}

// RemovePeer from addresses
func (peers *PeerAddresses) RemovePeer(address string) {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	newArr := []PeerAddress{}
	for _, savedAddress := range peers.Addresses {
		add := savedAddress
		if add.String() == address {
			continue
		}
		newArr = append(newArr, add)
	}
	peers.Addresses = newArr
}

// AppendPeers func
func (peers *PeerAddresses) AppendPeers(addresses *PeerAddresses) {
	for _, address := range addresses.GetAdresses() {
		peers.appendPeer(&address)
	}
}

// Add PeerAddreses from string, return if it was added
func (peers *PeerAddresses) Add(value string) (new bool, err error) {
	initial := len(peers.GetAdresses())
	err = peers.Set(value)
	new = (len(peers.GetAdresses()) - initial) > 0
	return new, err
}

// Set PeerAddreses from string
func (peers *PeerAddresses) Set(value string) error {

	addresses := strings.Split(value, ",")
	for _, item := range addresses {
		var address PeerAddress
		if err := address.Set(item); err != nil {
			return err
		} else if !strings.Contains(peers.String(), item) {
			peers.appendPeer(&address)
		}
	}
	return nil
}

// GetRandomPeer func
func (peers *PeerAddresses) GetRandomPeer(usedPeers map[string]bool) *PeerAddress {
	peers.mux.Lock()
	defer peers.mux.Unlock()

	peerNr := len(peers.Addresses)
	if len(usedPeers) >= peerNr {
		return nil
	}
	for {
		index := rand.Int() % peerNr
		peerAddress := peers.Addresses[index].String()
		if _, ok := usedPeers[peerAddress]; !ok {
			return &peers.Addresses[index]
		}
	}
}
