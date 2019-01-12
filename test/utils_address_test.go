package tests

import (
	"testing"

	"github.com/ageapps/gambercoin/pkg/utils"
)

const testAddress = "10.1.2.3"
const testAddress2 = "10.1.2.3:1111"
const testAddress3 = "10.1.2.3:2222"

func TestAdress(t *testing.T) {
	t.Log("Testing peer address struct")

	address, err := utils.GetPeerAddress(testAddress)
	if err == nil {
		t.Errorf("Test should fail, no port specified in %v", testAddress)
	}

	err = address.Set(testAddress2)
	if err != nil {
		t.Errorf("Test address not converted %v %v", testAddress2, err)
	}
	if address.String() != testAddress2 {
		t.Errorf("Addresses should match %v %v", address.String(), testAddress2)
	}
}

func TestAdresses(t *testing.T) {
	t.Log("Testing peer addresses struct")

	addresses := utils.EmptyAdresses()
	err := addresses.Set(testAddress2)
	err = addresses.Set(testAddress3)
	if err != nil {
		t.Errorf("Test address not converted %v %v", testAddress2, err)
	}
	if len(addresses.GetAdresses()) != 2 {
		t.Errorf("Addresses not saved %v", addresses.GetAdresses())
	}
	if addresses.GetAdresses()[0].String() != testAddress2 || addresses.GetAdresses()[1].String() != testAddress3 {
		t.Errorf("Addresses not matching %v", addresses.GetAdresses())
	}
	addresses2 := utils.EmptyAdresses()
	addresses2.AppendPeers(addresses)

	if addresses2.String() != (testAddress2 + "," + testAddress3) {
		t.Errorf("Addresses not matching %v", addresses.GetAdresses())
	}
	used := make(map[string]bool)
	random := addresses2.GetRandomPeer(used)

	if random == nil {
		t.Errorf("Random peer not working")
	}

	used[testAddress2] = true

	random = addresses2.GetRandomPeer(used)
	if random.String() != testAddress3 {
		t.Errorf("Random peer %v not correct", random)
	}
}
