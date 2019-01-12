package tests

import (
	"testing"

	"github.com/ageapps/gambercoin/pkg/utils"
)

const testString = "Hello!"

func TestHashValue(t *testing.T) {

	t.Log("Testing HashValue struct")

	hashValue := utils.MakeHashString(testString)

	if len(hashValue) != 32 {
		t.Errorf("Size not correct %v", hashValue.String())
	}

	t.Logf("Hash string created: %v", hashValue)

	hashValue2, err := utils.GetHash(hashValue.String())
	if err != nil {
		t.Errorf("Error creating hash %v %v", hashValue.String(), err)
	}
	t.Logf("Hash 2 created: %v", hashValue2.String())

	if hashValue.String() != hashValue2.String() {
		t.Errorf("Hashes should be the same %v %v", hashValue.String(), hashValue2.String())
	}
	if !hashValue.Equals(hashValue2.String()) {
		t.Errorf("Hashes should be equals %v %v", hashValue.String(), hashValue2.String())
	}
}
