package utils

import (
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	tests := []struct {
		keyType string
	}{
		{"rsa"},
		{"ed25519"},
	}

	for _, tt := range tests {
		t.Run(tt.keyType, func(t *testing.T) {
			pub, priv, err := GenerateKeyPair(tt.keyType)
			if err != nil {
				t.Fatalf("GenerateKeyPair(%s) error: %v", tt.keyType, err)
			}
			if pub == "" || priv == "" {
				t.Fatalf("GenerateKeyPair(%s) returned empty keys", tt.keyType)
			}

			typ, err := CheckKeyPair(pub, priv)
			if err != nil {
				t.Fatalf("CheckKeyPair(%s) error: %v", tt.keyType, err)
			}
			if typ != tt.keyType && !(tt.keyType == "rsa" && typ == "ssh-rsa") && !(tt.keyType == "ed25519" && typ == "ssh-ed25519") {
				t.Fatalf("CheckKeyPair(%s) returned wrong type: %s", tt.keyType, typ)
			}
		})
	}
}
