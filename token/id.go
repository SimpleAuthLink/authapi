package token

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
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
	rawSign := ed25519.Sign(privKey, hData[:])
	// encode to hex
	sign := make([]byte, hex.EncodedLen(len(rawSign)))
	hex.Encode(sign, rawSign)
	return sign
}

func (id *AppID) Verify(msg, sig []byte) bool {
	privKey := id.PrivKey()
	if privKey == nil {
		return false
	}
	// decode sign from hex
	rawSign := make([]byte, hex.DecodedLen(len(sig)))
	if _, err := hex.Decode(rawSign, sig); err != nil {
		return false
	}
	pubKey := privKey.Public().(ed25519.PublicKey)
	hMsg := sha256.Sum256(msg)
	return ed25519.Verify(pubKey, hMsg[:], rawSign)
}

func (id *AppID) NewToken(secret, email string) []byte {
	app := new(App).SetID(id)
	if app == nil {
		return nil
	}
	exp := NewExpiration(app.SessionDuration)
	msg := signMsg(id.Bytes(), []byte(secret), []byte(email), exp.Bytes())
	sig := id.Sign(msg)
	return fmtToken(exp.Marshal(), sig)
}

func (id *AppID) VerifyToken(token []byte, secret, email string) bool {
	if len(token) == 0 {
		return false
	}
	parts := bytes.Split(token, []byte{tokenSeparator})
	if len(parts) != 2 {
		return false
	}
	dExp := new(Expiration).Unmarshal(parts[0])
	if dExp == nil || !dExp.Valid() {
		return false
	}
	sig := parts[1]
	msg := signMsg(id.Bytes(), []byte(secret), []byte(email), dExp.Bytes())
	return id.Verify(msg, sig)
}

func signMsg(id, secret, email, exp []byte) []byte {
	return bytes.Join([][]byte{id, email, exp, secret}, nil)
}

func fmtToken(exp, sig []byte) []byte {
	t := append(exp, tokenSeparator)
	return append(t, sig...)
}
