package crypto

import (
	"reflect"
	"testing"
)

func TestThatGenerateReturnsInstanceOfCrypto(t *testing.T) {
	c := Generate()

	if reflect.TypeOf(c).Name() != "Crypto" {
		t.Fatal("test failed attempting to call method: Generate")
	}
}
