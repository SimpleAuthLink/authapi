package token

import (
	"crypto/ed25519"
	"crypto/sha256"
)

type AppID string

func (id *AppID) String() string {
	return string(*id)
}

func (id *AppID) SetString(data string) *AppID {
	newID := AppID(data)
	if !new(App).Unmarshal(newID.Bytes()).Valid() {
		return nil
	}
	*id = newID
	return id
}

func (id *AppID) Bytes() []byte {
	return []byte(*id)
}

func (id *AppID) SetBytes(data []byte) *AppID {
	return id.SetString(string(data))
}

func (id *AppID) PrivKey() ed25519.PrivateKey {
	bID := id.Bytes()
	if len(bID) == 0 {
		return nil
	}
	hID := sha256.Sum256(bID)
	return ed25519.NewKeyFromSeed(hID[:])
}

func (id *AppID) Sign(data []byte) []byte {
	privKey := id.PrivKey()
	if len(privKey) == 0 {
		return nil
	}
	hData := sha256.Sum256(data)
	return ed25519.Sign(privKey, hData[:])
}

func (id *AppID) Verify(data, sig []byte) bool {
	privKey := id.PrivKey()
	if privKey == nil {
		return false
	}
	pubKey := privKey.Public().(ed25519.PublicKey)
	hData := sha256.Sum256(data)
	return ed25519.Verify(pubKey, hData[:], sig)
}
