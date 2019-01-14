package tests

import (
	"testing"

	"github.com/ageapps/gambercoin/pkg/router"
)

const destination = "nodeA"
const destination2 = "nodeB"

func TestRouter(t *testing.T) {

	t.Log("Testing Router struct")

	router := router.NewRouter()

	if router.GetTableSize() != 0 {
		t.Error("router not empty")
	}

	added := router.AddEntry(destination, testAddress, false)

	if added {
		t.Error("Address should not be added " + testAddress)
	}
	added = router.AddEntry(destination, testAddress2, false)

	if !added || router.GetTableSize() != 1 {
		t.Error("Address not added correctly " + testAddress2)
	}
	added = router.AddEntry(destination, testAddress2, false)
	if added || router.GetTableSize() != 1 {
		t.Error("Address should not be added " + testAddress2)
	}

	added = router.AddEntry(destination, testAddress3, false)
	if !added || router.GetTableSize() != 1 {
		t.Error("Address should be updated " + testAddress3)
	}
	added = router.AddEntry(destination, testAddress2, true)
	if added || router.GetTableSize() != 1 {
		t.Error("Address should not be updated " + testAddress3)
	}

	address, found := router.GetAddress("")
	if found || address != nil {
		t.Error("Address should not be returned")
	}

	address, found = router.GetAddress(destination)
	if !found || address.String() != testAddress3 {
		t.Errorf("Address should match %v", address)
	}
	router.AddEntry(destination2, testAddress2, false)

	used := make(map[string]int)
	random := router.GetRandomDestination(used)

	if random == "" {
		t.Errorf("Random destination not working")
	}

	used[destination2] = 0

	random = router.GetRandomDestination(used)
	if random != destination {
		t.Error(router.GetTable())

		t.Errorf("Random destination %v not correct, should be %v", random, destination)
	}

}
