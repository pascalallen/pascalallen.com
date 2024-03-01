package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

type Crypto string

func Generate() Crypto {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}

	crypto := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(crypto, bytes)

	return Crypto(crypto)
}
