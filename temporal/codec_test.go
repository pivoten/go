package temporal

import (
	"bytes"
	"testing"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

func TestNewCodecKeyLength(t *testing.T) {
	if _, err := NewCodec("k", make([]byte, 16)); err == nil {
		t.Error("expected error for 16-byte key")
	}
	if _, err := NewCodec("k", make([]byte, 32)); err != nil {
		t.Errorf("32-byte key should be accepted: %v", err)
	}
}

func TestCodecRoundTripAndHidesPlaintext(t *testing.T) {
	codec, err := NewCodec("test-1", bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	dc := converter.GetDefaultDataConverter()

	secret := "PIVOTEN\\Administrator reset-password"
	p, err := dc.ToPayload(secret)
	if err != nil {
		t.Fatal(err)
	}

	enc, err := codec.Encode([]*commonpb.Payload{p})
	if err != nil {
		t.Fatal(err)
	}
	if string(enc[0].Metadata["encoding"]) != metadataEncodingEncrypted {
		t.Fatal("payload not marked encrypted")
	}
	if string(enc[0].Metadata[metadataEncryptionKeyID]) != "test-1" {
		t.Error("key id not recorded")
	}
	if bytes.Contains(enc[0].Data, []byte("Administrator")) {
		t.Fatal("plaintext leaked into ciphertext")
	}

	dec, err := codec.Decode(enc)
	if err != nil {
		t.Fatal(err)
	}
	var out string
	if err := dc.FromPayload(dec[0], &out); err != nil {
		t.Fatal(err)
	}
	if out != secret {
		t.Errorf("round-trip mismatch: got %q, want %q", out, secret)
	}
}

func TestCodecDecodePassesThroughUnencrypted(t *testing.T) {
	codec, _ := NewCodec("test-1", bytes.Repeat([]byte{7}, 32))
	plain := &commonpb.Payload{
		Metadata: map[string][]byte{"encoding": []byte("json/plain")},
		Data:     []byte(`"hi"`),
	}
	out, err := codec.Decode([]*commonpb.Payload{plain})
	if err != nil {
		t.Fatal(err)
	}
	if string(out[0].Data) != `"hi"` {
		t.Error("non-encrypted payload should pass through unchanged")
	}
}
