package imagecore

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
)

// ValidateBitPlane checks channel/bit constraints.
func ValidateBitPlane(req BitPlaneRequest, hasAlpha bool) error {
	if req.Bit < 0 || req.Bit > 7 {
		return fmt.Errorf("位索引必须在 0–7（0=LSB，7=MSB），收到: %d", req.Bit)
	}
	switch req.Channel {
	case ChannelR, ChannelG, ChannelB:
		return nil
	case ChannelA:
		if !hasAlpha {
			return fmt.Errorf("当前图像无有效 Alpha 通道")
		}
		return nil
	default:
		return fmt.Errorf("无效通道: %s", req.Channel)
	}
}

// RenderBitPlane produces a black/white image: bit==1 → white, bit==0 → black.
// Bit 0 is LSB, bit 7 is MSB.
func RenderBitPlane(src *image.RGBA, ch Channel, bit int) (*image.Gray, error) {
	if bit < 0 || bit > 7 {
		return nil, fmt.Errorf("位索引必须在 0–7，收到: %d", bit)
	}
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	out := image.NewGray(image.Rect(0, 0, w, h))
	mask := uint8(1 << uint(bit))

	chOffset := channelOffset(ch)
	if chOffset < 0 {
		return nil, fmt.Errorf("无效通道: %s", ch)
	}

	srcPix := src.Pix
	dstPix := out.Pix
	srcStride := src.Stride
	dstStride := out.Stride

	for y := 0; y < h; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride
		for x := 0; x < w; x++ {
			v := srcPix[srcRow+x*4+chOffset]
			if v&mask != 0 {
				dstPix[dstRow+x] = 255
			} else {
				dstPix[dstRow+x] = 0
			}
		}
	}
	return out, nil
}

func channelOffset(ch Channel) int {
	switch ch {
	case ChannelR:
		return 0
	case ChannelG:
		return 1
	case ChannelB:
		return 2
	case ChannelA:
		return 3
	default:
		return -1
	}
}

// RenderBitPlanePNG encodes a bit plane as PNG bytes.
func RenderBitPlanePNG(src *image.RGBA, ch Channel, bit int) ([]byte, error) {
	gray, err := RenderBitPlane(src, ch, bit)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := encodePNG(&buf, gray); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderBitPlaneB64 returns base64-encoded PNG of the bit plane.
func RenderBitPlaneB64(src *image.RGBA, ch Channel, bit int) (string, error) {
	pngBytes, err := RenderBitPlanePNG(src, ch, bit)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(pngBytes), nil
}

// SaveBitPlanePNG writes a bit plane PNG to path (atomic via temp file).
func SaveBitPlanePNG(src *image.RGBA, ch Channel, bit int, path string) (int64, error) {
	pngBytes, err := RenderBitPlanePNG(src, ch, bit)
	if err != nil {
		return 0, err
	}
	return writeFileAtomic(path, pngBytes)
}

func encodePNG(w io.Writer, img image.Image) error {
	enc := png.Encoder{CompressionLevel: png.DefaultCompression}
	return enc.Encode(w, img)
}

// OriginalViewPNG returns a PNG of the original RGBA image (for "原图" view).
func OriginalViewPNG(src *image.RGBA) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodePNG(&buf, src); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeFileAtomic(path string, data []byte) (int64, error) {
	dir := ""
	if i := lastSlash(path); i >= 0 {
		dir = path[:i]
	}
	tmp, err := os.CreateTemp(dir, ".stegsolve-*.tmp")
	if err != nil {
		return 0, fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	n, err := tmp.Write(data)
	if err != nil {
		tmp.Close()
		return 0, fmt.Errorf("写入临时文件失败: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return 0, fmt.Errorf("同步临时文件失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return 0, fmt.Errorf("关闭临时文件失败: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return 0, fmt.Errorf("替换目标文件失败: %w", err)
	}
	cleanup = false
	return int64(n), nil
}

func lastSlash(path string) int {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return i
		}
	}
	return -1
}
