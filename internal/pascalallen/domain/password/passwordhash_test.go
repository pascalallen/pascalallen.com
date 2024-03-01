package password

import (
	"reflect"
	"testing"
)

func TestThatCreateReturnsInstanceOfPasswordHash(t *testing.T) {
	p := Create("pa$$w0rd")

	if reflect.TypeOf(p).Name() != "PasswordHash" {
		t.Fatal("test failed attempting to call method: Create")
	}
}

func TestThatCompareReturnsTrueOnValidPasswordString(t *testing.T) {
	p := "pa$$w0rd"
	ph := Create(p)

	if ph.Compare(p) == false {
		t.Fatal("test failed attempting to call method: Compare")
	}
}

func TestThatCompareReturnsFalseOnInvalidPasswordString(t *testing.T) {
	p := "pa$$word"
	ph := Create(p)

	if ph.Compare("invalid_pa$$w0rd") == true {
		t.Fatal("test failed attempting to call method: Compare")
	}
}
