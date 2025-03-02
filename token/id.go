package token

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

func (id *AppID) PrivKey(secret Secret) ed25519.PrivateKey {
	if id == nil {
		return nil
	}
	if !secret.Valid() {
		return nil
	}
	bID := id.Bytes()
	if len(bID) == 0 {
		return nil
	}
	hFn := hmac.New(sha256.New, secret.Bytes())
	hID := hFn.Sum(bID)
	return ed25519.NewKeyFromSeed(hID[:32])
}

func (id *AppID) Sign(secret Secret, msg []byte) []byte {
	if id == nil || len(msg) == 0 {
		return nil
	}
	privKey := id.PrivKey(secret)
	if len(privKey) == 0 {
		return nil
	}
	data := append(msg, hedgedNonce(privKey[:], msg)...)
	rawSign := ed25519.Sign(privKey, data[:])
	// encode to base64
	sign := make([]byte, base64.RawStdEncoding.EncodedLen(len(rawSign)))
	base64.RawStdEncoding.Encode(sign, rawSign)
	return sign
}

func (id *AppID) Verify(secret Secret, msg, sig []byte) bool {
	if id == nil || len(msg) == 0 || len(sig) == 0 {
		return false
	}
	privKey := id.PrivKey(secret)
	if privKey == nil {
		return false
	}
	// decode sign from base64
	rawSign := make([]byte, base64.RawStdEncoding.DecodedLen(len(sig)))
	if _, err := base64.RawStdEncoding.Decode(rawSign, sig); err != nil {
		return false
	}
	data := append(msg, hedgedNonce(privKey[:], msg)...)
	pubKey := privKey.Public().(ed25519.PublicKey)
	return ed25519.Verify(pubKey, data, rawSign)
}

func (id *AppID) GenerateToken(secret Secret, email string) Token {
	if id == nil {
		return nil
	}
	app := new(App).SetID(id)
	if app == nil {
		return nil
	}
	exp := new(Expiration).SetDuration(app.SessionDuration)
	msg := id.Message(email, *exp)
	sig := id.Sign(secret, msg)
	if len(sig) == 0 {
		return nil
	}
	return *new(Token).SetExpiration(*exp).SetSignature(sig)
}

func (id *AppID) Message(email string, exp Expiration) []byte {
	if id == nil || len(email) == 0 || !exp.Valid() {
		return nil
	}
	hmsg := sha256.Sum256(append(append(id.Bytes(), []byte(email)...), exp.Bytes()...))
	return hmsg[:]
}

func (id *AppID) VerifyToken(token Token, secret Secret, email string) bool {
	if id == nil {
		return false
	}
	exp := token.Expiration()
	if exp == nil || !exp.Valid() {
		return false
	}
	sig := token.Signature()
	if len(sig) == 0 {
		return false
	}
	msg := id.Message(email, *exp)
	if len(msg) == 0 {
		return false
	}
	return id.Verify(secret, msg, sig)
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
