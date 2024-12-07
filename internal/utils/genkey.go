package utils

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"errors"
	"fmt"

	"golang.org/x/crypto/ssh"
)

func GenerateKeyPair(typ string) (pub string, priv string, err error) {
	var privateKey interface{}

	switch typ {
	case "ssh-rsa":
		fallthrough
	case "rsa":
		privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return "", "", err
		}

	case "ssh-ed25519":
		fallthrough
	case "ed25519":
		_, privateKey, err = ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return "", "", err
		}

	default:
		return "", "", errors.New("unsupported key type")
	}

	// Encode private key to PKCS#8 ASN.1 PEM
	privBlock, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return "", "", err
	}
	priv = string(pem.EncodeToMemory(privBlock))

	// Encode public key to PKIX ASN.1 PEM
	privSSH, err := ssh.ParsePrivateKey([]byte(priv))
	if err != nil {
		return "", "", err
	}
	pubBytes := ssh.MarshalAuthorizedKey(privSSH.PublicKey())
	pub = string(pubBytes)

	return pub, priv, nil
}

func PublicKeyFromPrivateKey(priv string) (pub string, err error) {
	s, err := ssh.ParsePrivateKey([]byte(priv))
	if err != nil {
		return "", err
	}

	pubBytes := ssh.MarshalAuthorizedKey(s.PublicKey())
	pub = string(pubBytes)

	return pub, nil
}

func CheckKeyPair(pub string, priv string) (typ string, err error) {
	s, err := ssh.ParsePrivateKey([]byte(priv))
	if err != nil {
		return "", err
	}
	p, _, _, rest, err := ssh.ParseAuthorizedKey([]byte(pub))
	if len(rest) > 0 {
		return "", fmt.Errorf("trailing data after single public key")
	}
	if err != nil {
		return "", err
	}
	if p.Type() != s.PublicKey().Type() {
		return "", fmt.Errorf("public key type does not match private key type")
	}
	if !bytes.Equal(p.Marshal(), s.PublicKey().Marshal()) {
		return "", fmt.Errorf("public key does not match private key")
	}
	return s.PublicKey().Type(), nil
}
