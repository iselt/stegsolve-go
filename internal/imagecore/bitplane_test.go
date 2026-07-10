package imagecore

import (
	"image"
	"image/color"
	"testing"
)

// fixed 2×2 RGBA matrix for exhaustive bit-plane checks.
//
//	(0,0) R=0b10101010 G=0b01010101 B=0b11110000 A=0b00001111
//	(1,0) R=0b00000001 G=0b00000010 B=0b00000100 A=0b00001000
//	(0,1) R=0b11111111 G=0b00000000 B=0b10101010 A=0b01010101
//	(1,1) R=0b10000000 G=0b01000000 B=0b00100000 A=0b00010000  (semi-transparent)
func sampleRGBA() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{0b10101010, 0b01010101, 0b11110000, 0b00001111})
	img.SetRGBA(1, 0, color.RGBA{0b00000001, 0b00000010, 0b00000100, 0b00001000})
	img.SetRGBA(0, 1, color.RGBA{0b11111111, 0b00000000, 0b10101010, 0b01010101})
	img.SetRGBA(1, 1, color.RGBA{0b10000000, 0b01000000, 0b00100000, 0b00010000})
	return img
}

func channelValue(img *image.RGBA, x, y int, ch Channel) uint8 {
	i := img.PixOffset(x, y)
	switch ch {
	case ChannelR:
		return img.Pix[i+0]
	case ChannelG:
		return img.Pix[i+1]
	case ChannelB:
		return img.Pix[i+2]
	case ChannelA:
		return img.Pix[i+3]
	}
	return 0
}

func TestRenderBitPlaneAll32(t *testing.T) {
	src := sampleRGBA()
	channels := []Channel{ChannelR, ChannelG, ChannelB, ChannelA}

	for _, ch := range channels {
		for bit := 0; bit <= 7; bit++ {
			gray, err := RenderBitPlane(src, ch, bit)
			if err != nil {
				t.Fatalf("%s bit%d: %v", ch, bit, err)
			}
			mask := uint8(1 << uint(bit))
			for y := 0; y < 2; y++ {
				for x := 0; x < 2; x++ {
					v := channelValue(src, x, y, ch)
					want := uint8(0)
					if v&mask != 0 {
						want = 255
					}
					got := gray.GrayAt(x, y).Y
					if got != want {
						t.Errorf("%s bit%d pixel(%d,%d): got %d want %d (chval=%08b)",
							ch, bit, x, y, got, want, v)
					}
				}
			}
		}
	}
}

func TestBitPlaneLSBIsBit0(t *testing.T) {
	// Single pixel R=1 → only bit0 should be white
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{1, 0, 0, 255})

	g0, _ := RenderBitPlane(img, ChannelR, 0)
	g1, _ := RenderBitPlane(img, ChannelR, 1)
	if g0.GrayAt(0, 0).Y != 255 {
		t.Fatal("bit0 (LSB) should be white when R=1")
	}
	if g1.GrayAt(0, 0).Y != 0 {
		t.Fatal("bit1 should be black when R=1")
	}
}

func TestBitPlaneMSBIsBit7(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{0x80, 0, 0, 255})

	g7, _ := RenderBitPlane(img, ChannelR, 7)
	g0, _ := RenderBitPlane(img, ChannelR, 0)
	if g7.GrayAt(0, 0).Y != 255 {
		t.Fatal("bit7 (MSB) should be white when R=0x80")
	}
	if g0.GrayAt(0, 0).Y != 0 {
		t.Fatal("bit0 should be black when R=0x80")
	}
}

func TestTransparentRGBNotPremultiplied(t *testing.T) {
	// Build NRGBA with semi-transparent pure red, convert via toNRGBA
	src := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	src.SetNRGBA(0, 0, color.NRGBA{R: 200, G: 100, B: 50, A: 128})
	rgba := toNRGBA(src)

	r := rgba.Pix[0]
	g := rgba.Pix[1]
	b := rgba.Pix[2]
	a := rgba.Pix[3]
	if a != 128 {
		t.Fatalf("alpha: got %d want 128", a)
	}
	// Allow ±1 rounding from un-premultiply path; NRGBA path should be exact.
	if r != 200 || g != 100 || b != 50 {
		t.Fatalf("RGB premultiplied or wrong: got R=%d G=%d B=%d want 200,100,50", r, g, b)
	}
}

func TestValidateBitPlaneNoAlpha(t *testing.T) {
	err := ValidateBitPlane(BitPlaneRequest{Channel: ChannelA, Bit: 0}, false)
	if err == nil {
		t.Fatal("expected error for alpha on opaque image")
	}
	err = ValidateBitPlane(BitPlaneRequest{Channel: ChannelR, Bit: 0}, false)
	if err != nil {
		t.Fatal(err)
	}
	err = ValidateBitPlane(BitPlaneRequest{Channel: ChannelR, Bit: 8}, true)
	if err == nil {
		t.Fatal("expected bit range error")
	}
}

func TestDetectAlpha(t *testing.T) {
	opaque := image.NewRGBA(image.Rect(0, 0, 1, 1))
	opaque.SetRGBA(0, 0, color.RGBA{1, 2, 3, 255})
	if detectAlpha(opaque) {
		t.Fatal("opaque should not have alpha")
	}
	trans := image.NewRGBA(image.Rect(0, 0, 1, 1))
	trans.SetRGBA(0, 0, color.RGBA{1, 2, 3, 254})
	if !detectAlpha(trans) {
		t.Fatal("A=254 should count as alpha")
	}
}
