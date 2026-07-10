package imagecore

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

// 2×2 with known channel values for exact bit stream checks.
func extractSample() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	// Row-major: (0,0), (1,0), (0,1), (1,1)
	img.SetRGBA(0, 0, color.RGBA{0b00000001, 0b00000010, 0b00000100, 0b00001000}) // R0=1 G1=1 B2=1 A3=1
	img.SetRGBA(1, 0, color.RGBA{0b00000000, 0b00000000, 0b00000000, 0b00000000})
	img.SetRGBA(0, 1, color.RGBA{0b11111111, 0b00000000, 0b10101010, 0b01010101})
	img.SetRGBA(1, 1, color.RGBA{0b10000000, 0b01000000, 0b00100000, 0b00010000})
	return img
}

func baseReq() ExtractRequest {
	return ExtractRequest{
		MaskA:        0,
		MaskR:        0,
		MaskG:        0,
		MaskB:        0,
		ChannelOrder: []Channel{ChannelR, ChannelG, ChannelB},
		Traverse:     TraverseRow,
		BitOrder:     BitOrderLSBFirst,
	}
}

func TestExtractR0RowLSB(t *testing.T) {
	src := extractSample()
	req := baseReq()
	req.MaskR = 1 << 0 // R0 only
	// Row: R0 of each pixel: 1, 0, 1, 0 → 0b10100000 = 0xA0
	// (1,1) R=0b10000000 → R0=0
	got, err := ExtractBytes(src, req)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != 0xA0 {
		t.Fatalf("got %v want [0xA0]", got)
	}
}

func TestExtractR0ColLSB(t *testing.T) {
	src := extractSample()
	req := baseReq()
	req.MaskR = 1 << 0
	req.Traverse = TraverseCol
	// Col: (0,0),(0,1),(1,0),(1,1) → R0: 1, 1, 0, 0 → 0b11000000 = 0xC0
	got, err := ExtractBytes(src, req)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != 0xC0 {
		t.Fatalf("got %02X want C0", got[0])
	}
}

func TestExtractRGB0Order(t *testing.T) {
	// Single pixel R=1,G=0,B=1 → RGB0: 1,0,1 → 0b10100000
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{1, 0, 1, 255})
	req := baseReq()
	req.MaskR = 1
	req.MaskG = 1
	req.MaskB = 1
	req.ChannelOrder = []Channel{ChannelR, ChannelG, ChannelB}
	got, err := ExtractBytes(img, req)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != 0b10100000 {
		t.Fatalf("got %08b want 10100000", got[0])
	}

	// BGR order: B,G,R → 1,0,1 same; try BRG: B=1,R=1,G=0 → 0b11000000
	req.ChannelOrder = []Channel{ChannelB, ChannelR, ChannelG}
	got, err = ExtractBytes(img, req)
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != 0b11000000 {
		t.Fatalf("BRG got %08b want 11000000", got[0])
	}
}

func TestExtractMSBFirst(t *testing.T) {
	// R = 0b10000001 → bits 7 and 0 set. Select both, MSB first → bit7 then bit0: 1,1
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{0b10000001, 0, 0, 255})
	req := baseReq()
	req.MaskR = (1 << 7) | (1 << 0)
	req.BitOrder = BitOrderMSBFirst
	got, err := ExtractBytes(img, req)
	if err != nil {
		t.Fatal(err)
	}
	// bits: 1 (bit7), 1 (bit0) → 0b11000000
	if got[0] != 0b11000000 {
		t.Fatalf("MSB first got %08b", got[0])
	}

	req.BitOrder = BitOrderLSBFirst
	got, err = ExtractBytes(img, req)
	if err != nil {
		t.Fatal(err)
	}
	// bits: 1 (bit0), 1 (bit7) → same 0b11000000 for this value
	if got[0] != 0b11000000 {
		t.Fatalf("LSB first got %08b", got[0])
	}
}

func TestExtractMultiBitDistinctOrder(t *testing.T) {
	// R = 0b00000110 → bit1=1, bit2=1. Select bit1|bit2.
	// LSB first: bit1 then bit2 → 1,1
	// But use value where order matters: R = 0b00000010 → only bit1
	// Select bits 0,1,2: values 0,1,0
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{0b00000010, 0, 0, 255})
	req := baseReq()
	req.MaskR = 0b00000111 // bits 0,1,2
	req.BitOrder = BitOrderLSBFirst
	got, _ := ExtractBytes(img, req)
	// 0,1,0 → 0b01000000
	if got[0] != 0b01000000 {
		t.Fatalf("LSB multi got %08b want 01000000", got[0])
	}
	req.BitOrder = BitOrderMSBFirst
	// among 0,1,2 only those in mask, MSB first among selected: bit2, bit1, bit0 → 0,1,0
	got, _ = ExtractBytes(img, req)
	if got[0] != 0b01000000 {
		t.Fatalf("MSB multi got %08b want 01000000", got[0])
	}

	// R=0b00000100 bit2 only selected with bits 0,1,2
	img.SetRGBA(0, 0, color.RGBA{0b00000100, 0, 0, 255})
	req.BitOrder = BitOrderLSBFirst
	got, _ = ExtractBytes(img, req)
	// bit0=0,bit1=0,bit2=1 → 0b00100000
	if got[0] != 0b00100000 {
		t.Fatalf("LSB bit2 got %08b", got[0])
	}
	req.BitOrder = BitOrderMSBFirst
	got, _ = ExtractBytes(img, req)
	// bit2=1,bit1=0,bit0=0 → 0b10000000
	if got[0] != 0b10000000 {
		t.Fatalf("MSB bit2 got %08b want 10000000", got[0])
	}
}

func TestExtractNonByteAlignedPadding(t *testing.T) {
	// 3 bits only → one byte with low bits zero-padded
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{1, 1, 1, 255})
	req := baseReq()
	req.MaskR = 1
	req.MaskG = 1
	req.MaskB = 1
	got, err := ExtractBytes(img, req)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len %d", len(got))
	}
	// 1,1,1 + pad → 0b11100000
	if got[0] != 0b11100000 {
		t.Fatalf("got %08b", got[0])
	}
}

func TestExtractEmptySelection(t *testing.T) {
	src := extractSample()
	req := baseReq()
	err := ValidateExtract(req, true)
	if err == nil {
		t.Fatal("expected empty selection error")
	}
	_, err = ExtractBytes(src, req)
	if err == nil {
		t.Fatal("expected extract error")
	}
}

func TestExtractAlphaRequiresHasAlpha(t *testing.T) {
	req := baseReq()
	req.MaskA = 1
	req.ChannelOrder = []Channel{ChannelA, ChannelR, ChannelG, ChannelB}
	if err := ValidateExtract(req, false); err == nil {
		t.Fatal("expected alpha rejection")
	}
}

func TestPreviewMatchesFullExtract(t *testing.T) {
	src := sampleRGBA()
	req := baseReq()
	req.MaskR = 0xFF
	req.MaskG = 0xFF
	req.MaskB = 0xFF
	req.MaskA = 0xFF
	req.ChannelOrder = []Channel{ChannelR, ChannelG, ChannelB, ChannelA}

	full, err := ExtractBytes(src, req)
	if err != nil {
		t.Fatal(err)
	}
	prev, err := PreviewExtract(src, req)
	if err != nil {
		t.Fatal(err)
	}
	if prev.TotalBytes != int64(len(full)) {
		t.Fatalf("total %d vs %d", prev.TotalBytes, len(full))
	}
	// rebuild capture from rows is harder; compare via ExtractTo capture
	total, capture, err := ExtractTo(src, req, bytes.NewBuffer(nil), PreviewByteLimit)
	if err != nil {
		t.Fatal(err)
	}
	if total != int64(len(full)) {
		t.Fatal("total mismatch")
	}
	if !bytes.Equal(capture, full[:min(len(full), PreviewByteLimit)]) {
		t.Fatal("capture prefix mismatch")
	}
	if prev.PreviewBytes != len(capture) {
		t.Fatalf("preview bytes %d vs %d", prev.PreviewBytes, len(capture))
	}
}

func TestSaveExtractAtomic(t *testing.T) {
	src := extractSample()
	req := baseReq()
	req.MaskR = 1
	dir := t.TempDir()
	path := filepath.Join(dir, "out.bin")
	n, err := SaveExtract(src, req, path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if int64(len(data)) != n {
		t.Fatalf("size %d vs %d", len(data), n)
	}
	want, _ := ExtractBytes(src, req)
	if !bytes.Equal(data, want) {
		t.Fatalf("file %v want %v", data, want)
	}
}

func TestSelectedBitCount(t *testing.T) {
	req := baseReq()
	req.MaskR = 0b101
	req.MaskG = 0b1
	if SelectedBitCount(req) != 3 {
		t.Fatal(SelectedBitCount(req))
	}
}

func TestHexFormat(t *testing.T) {
	prev := BuildPreview([]byte{0x00, 0x41, 0xFF}, 3)
	if len(prev.Rows) != 1 {
		t.Fatal(len(prev.Rows))
	}
	if prev.Rows[0].Hex != "00 41 FF" {
		t.Fatalf("hex %q", prev.Rows[0].Hex)
	}
	if prev.Rows[0].ASCII != ".A." {
		t.Fatalf("ascii %q", prev.Rows[0].ASCII)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
