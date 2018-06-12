package qbchain

import (
	"testing"

	"github.com/izqui/helpers"
)

func TestGenerateNewKeypair(t *testing.T) {
	keypair := GenerateNewKeypair()

	if len(keypair.Public) > 80 || len(keypair.Public) < 1 {
		t.Error("Failed to generated Public key")
	}
	if len(keypair.Private) > 80 || len(keypair.Private) < 1 {
		t.Error("Failed to generated Private key")
	}
}

func TestSignAndSignatureVerify(t *testing.T) {
	for i := 0; i < 5; i++ {
		keypair := GenerateNewKeypair()

		data := helpers.ArrayOfBytes(i, 'a')
		hash := helpers.SHA256(data)

		signature, err := keypair.Sign(hash)

		if err != nil {
			t.Error("base58 error")
		} else if !SignatureVerify(keypair.Public, signature, hash) {
			t.Error("Signing and verifying error", len(keypair.Public))
		}
	}
}
