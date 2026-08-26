package logic

import (
	"encoding/base64"
	"testing"
)

// TestDecodeImage 验证 base64 解码成功且内容正确。
func TestDecodeImage(t *testing.T) {
	raw := []byte("fake-image-bytes")
	b64 := base64.StdEncoding.EncodeToString(raw)
	data, err := decodeImage(b64)
	if err != nil {
		t.Fatalf("decodeImage failed: %v", err)
	}
	if string(data) != string(raw) {
		t.Fatalf("decode mismatch: got %s want %s", data, raw)
	}
}

// TestDecodeImageDataURI 验证带 data URI 前缀的 base64 也能正确解码。
func TestDecodeImageDataURI(t *testing.T) {
	raw := []byte("png-bytes")
	b64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw)
	data, err := decodeImage(b64)
	if err != nil {
		t.Fatalf("decodeImage failed: %v", err)
	}
	if string(data) != string(raw) {
		t.Fatalf("decode mismatch: got %s want %s", data, raw)
	}
}

// TestDecodeImageInvalid 验证非法 base64 返回错误。
func TestDecodeImageInvalid(t *testing.T) {
	if _, err := decodeImage("@@@not-base64@@@"); err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

// TestDecodeImageTooLarge 验证超过大小上限的图片被拒绝。
func TestDecodeImageTooLarge(t *testing.T) {
	big := make([]byte, maxCertImageBytes+1)
	b64 := base64.StdEncoding.EncodeToString(big)
	if _, err := decodeImage(b64); err == nil {
		t.Fatal("expected error for oversized image")
	}
}

// TestGuessImageExt 验证扩展名推断逻辑。
func TestGuessImageExt(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"data:image/png;base64,iVBOR...", ".png"},
		{"data:image/jpeg;base64,/9j/4AAQ", ".jpg"},
		{"iVBORw0KGgoAAAANSUhEUg", ".png"}, // png 魔数
		{"/9j/4AAQSkZJRgABAQE", ".jpg"},    // jpeg 魔数
		{"unknown-content", ".jpg"},         // 未知默认 jpg
	}
	for _, c := range cases {
		if got := guessImageExt(c.in); got != c.want {
			t.Errorf("guessImageExt(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
