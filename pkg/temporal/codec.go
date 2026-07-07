package temporal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

const (
	metadataEncodingEncrypted = "binary/encrypted"
	metadataEncryptionKeyID   = "encryption-key-id"
)

// Codec is a Temporal PayloadCodec that encrypts payloads with AES-256-GCM, so
// workflow/activity inputs and results are encrypted at rest in Temporal (and
// decryptable by a codec server for the Web UI). Encryption is applied to each
// payload individually, preserving Temporal's metadata routing.
type Codec struct {
	keyID string
	key   []byte // 32 bytes (AES-256)
}

// NewCodec builds a codec from a 32-byte key. keyID is recorded on each payload
// so a future key rotation can select the right key on decode.
func NewCodec(keyID string, key []byte) (*Codec, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes (AES-256), got %d", len(key))
	}
	if keyID == "" {
		keyID = "default"
	}
	return &Codec{keyID: keyID, key: key}, nil
}

// Encode encrypts each payload.
func (c *Codec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	out := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		plain, err := p.Marshal()
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		enc, err := c.encrypt(plain)
		if err != nil {
			return nil, err
		}
		out[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				"encoding":              []byte(metadataEncodingEncrypted),
				metadataEncryptionKeyID: []byte(c.keyID),
			},
			Data: enc,
		}
	}
	return out, nil
}

// Decode decrypts payloads this codec encrypted, and passes others through.
func (c *Codec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	out := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		if string(p.Metadata["encoding"]) != metadataEncodingEncrypted {
			out[i] = p
			continue
		}
		plain, err := c.decrypt(p.Data)
		if err != nil {
			return nil, err
		}
		out[i] = &commonpb.Payload{}
		if err := out[i].Unmarshal(plain); err != nil {
			return nil, fmt.Errorf("unmarshal decrypted payload: %w", err)
		}
	}
	return out, nil
}

func (c *Codec) gcm() (cipher.AEAD, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (c *Codec) encrypt(plain []byte) ([]byte, error) {
	gcm, err := c.gcm()
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plain, nil), nil
}

func (c *Codec) decrypt(data []byte) ([]byte, error) {
	gcm, err := c.gcm()
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(data) < ns {
		return nil, fmt.Errorf("ciphertext too short")
	}
	return gcm.Open(nil, data[:ns], data[ns:], nil)
}

// NewEncryptionDataConverter wraps the default converter so all payloads are
// encrypted with the given codec.
func NewEncryptionDataConverter(codec *Codec) converter.DataConverter {
	return converter.NewCodecDataConverter(converter.GetDefaultDataConverter(), codec)
}
