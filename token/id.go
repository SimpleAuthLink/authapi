package token

import (
	"bytes"
	"crypto/ed25519"
	"crypto/hmac"
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

func (id *AppID) PrivKey(secret []byte) ed25519.PrivateKey {
	hFn := hmac.New(sha256.New, secret)

	bID := id.Bytes()
	if len(bID) == 0 {
		return nil
	}
	// hID := sha256.Sum256(bID)
	hID := hFn.Sum(bID)
	return ed25519.NewKeyFromSeed(hID[:32])
}

func (id *AppID) Sign(secret, msg []byte) []byte {
	privKey := id.PrivKey(secret)
	if len(privKey) == 0 {
		return nil
	}
	hmsg := sha256.Sum256(msg)
	data := append(msg, hedgedNonce(privKey[:], hmsg[:])...)
	rawSign := ed25519.Sign(privKey, data[:])
	// encode to hex
	sign := make([]byte, hex.EncodedLen(len(rawSign)))
	hex.Encode(sign, rawSign)
	return sign
}

func (id *AppID) Verify(secret, msg, sig []byte) bool {
	privKey := id.PrivKey(secret)
	if privKey == nil {
		return false
	}
	// decode sign from hex
	rawSign := make([]byte, hex.DecodedLen(len(sig)))
	if _, err := hex.Decode(rawSign, sig); err != nil {
		return false
	}
	hmsg := sha256.Sum256(msg)
	data := append(msg, hedgedNonce(privKey[:], hmsg[:])...)
	pubKey := privKey.Public().(ed25519.PublicKey)
	return ed25519.Verify(pubKey, data, rawSign)
}

func (id *AppID) NewToken(secret []byte, email string) []byte {
	app := new(App).SetID(id)
	if app == nil {
		return nil
	}
	exp := NewExpiration(app.SessionDuration)
	msg := signMsg(id.Bytes(), []byte(email), exp.Bytes())
	sig := id.Sign(secret, msg)
	return fmtToken(exp.Marshal(), sig)
}

func (id *AppID) VerifyToken(secret, token []byte, email string) bool {
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
	msg := signMsg(id.Bytes(), []byte(email), dExp.Bytes())
	return id.Verify(secret, msg, sig)
}

func signMsg(id, email, exp []byte) []byte {
	res := append(id, email...)
	return append(res, exp...)
}

func fmtToken(exp, sig []byte) []byte {
	t := append(exp, tokenSeparator)
	return append(t, sig...)
}

func hedgedNonce(inputs ...[]byte) []byte {
	if len(inputs) == 0 || len(inputs[0]) == 0 {
		return nil
	}
	hFn := hmac.New(sha256.New, inputs[0])
	for _, in := range inputs[1:] {
		hFn.Write(in)
	}
	return hFn.Sum(nil)
}
