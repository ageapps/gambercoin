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

	added := router.SetEntry(destination, testAddress)

	if added {
		t.Error("Address should not be added " + testAddress)
	}
	added = router.SetEntry(destination, testAddress2)

	if !added || router.GetTableSize() != 1 {
		t.Error("Address not added correctly " + testAddress2)
	}
	added = router.SetEntry(destination, testAddress2)
	if added || router.GetTableSize() != 1 {
		t.Error("Address should not be added " + testAddress2)
	}

	added = router.SetEntry(destination, testAddress3)
	if !added || router.GetTableSize() != 1 {
		t.Error("Address should be updated " + testAddress3)
	}

	address, found := router.GetAddress("")
	if found || address != nil {
		t.Error("Address should not be returned")
	}

	address, found = router.GetAddress(destination)
	if !found || address.String() != testAddress3 {
		t.Errorf("Address should match %v", address)
	}
	router.SetEntry(destination2, testAddress2)

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
