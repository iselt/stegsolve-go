package imagecore

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

// LoadedImage holds a decoded non-premultiplied RGBA image and metadata.
type LoadedImage struct {
	ID       string
	Name     string
	Format   string
	FileSize int64
	HasAlpha bool
	RGBA     *image.RGBA
}

// LoadFromPath reads and decodes an image file into non-premultiplied RGBA8.
func LoadFromPath(path string) (*LoadedImage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("读取文件信息失败: %w", err)
	}

	return LoadFromReader(f, filepath.Base(path), info.Size())
}

// LoadFromReader decodes an image from r.
func LoadFromReader(r io.Reader, name string, fileSize int64) (*LoadedImage, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("读取图像数据失败: %w", err)
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("无法识别图像格式或文件已损坏: %w", err)
	}

	pixels := int64(cfg.Width) * int64(cfg.Height)
	if pixels <= 0 {
		return nil, fmt.Errorf("无效的图像尺寸: %dx%d", cfg.Width, cfg.Height)
	}
	if pixels > MaxPixels {
		return nil, fmt.Errorf("图像像素数 %d 超过限制 %d（约 5000 万）", pixels, MaxPixels)
	}

	img, format2, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("解码图像失败: %w", err)
	}
	if format == "" {
		format = format2
	}

	rgba := toNRGBA(img)
	hasAlpha := detectAlpha(rgba)

	return &LoadedImage{
		ID:       uuid.NewString(),
		Name:     name,
		Format:   strings.ToUpper(format),
		FileSize: fileSize,
		HasAlpha: hasAlpha,
		RGBA:     rgba,
	}, nil
}

// toNRGBA converts any image to non-premultiplied RGBA8 pixel storage.
// image.RGBA.Pix is used as a raw 8-bit buffer of non-premultiplied channels
// for steganalysis; do not use draw.Draw from NRGBA (it premultiplies).
func toNRGBA(src image.Image) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fast path: copy NRGBA bytes directly (already non-premultiplied).
	if s, ok := src.(*image.NRGBA); ok {
		for y := 0; y < h; y++ {
			srcOff := (y + b.Min.Y - s.Rect.Min.Y) * s.Stride
			dstOff := y * dst.Stride
			// Account for Min.X offset in source
			srcOff += (b.Min.X - s.Rect.Min.X) * 4
			copy(dst.Pix[dstOff:dstOff+w*4], s.Pix[srcOff:srcOff+w*4])
		}
		return dst
	}

	// Fast path: opaque or already-RGBA sources — still un-premultiply via color model
	// when needed. For *image.RGBA from JPEG/BMP, A=255 and values are fine as-is
	// for opaque images. For safety we always go through non-premultiplied conversion.
	if s, ok := src.(*image.RGBA); ok {
		// If fully opaque, raw copy is safe (JPEG/BMP path).
		opaque := true
		for i := 3; i < len(s.Pix); i += 4 {
			if s.Pix[i] != 255 {
				opaque = false
				break
			}
		}
		if opaque && s.Rect.Eq(b) && s.Stride == 4*w {
			copy(dst.Pix, s.Pix)
			return dst
		}
	}

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			// Prefer NRGBAModel to obtain non-premultiplied 8-bit channels.
			nrgba := color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA)
			i := dst.PixOffset(x-b.Min.X, y-b.Min.Y)
			dst.Pix[i+0] = nrgba.R
			dst.Pix[i+1] = nrgba.G
			dst.Pix[i+2] = nrgba.B
			dst.Pix[i+3] = nrgba.A
		}
	}
	return dst
}

// detectAlpha reports whether any pixel has A != 255.
func detectAlpha(img *image.RGBA) bool {
	pix := img.Pix
	for i := 3; i < len(pix); i += 4 {
		if pix[i] != 255 {
			return true
		}
	}
	return false
}

// PreviewPNGB64 encodes the full image as a lossless PNG data URL payload (raw base64).
func (li *LoadedImage) PreviewPNGB64() (string, error) {
	var buf bytes.Buffer
	if err := encodePNG(&buf, li.RGBA); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// ToInfo builds the public ImageInfo DTO.
func (li *LoadedImage) ToInfo() (*ImageInfo, error) {
	b64, err := li.PreviewPNGB64()
	if err != nil {
		return nil, err
	}
	bounds := li.RGBA.Bounds()
	return &ImageInfo{
		ImageID:       li.ID,
		Name:          li.Name,
		Format:        li.Format,
		Width:         bounds.Dx(),
		Height:        bounds.Dy(),
		FileSize:      li.FileSize,
		HasAlpha:      li.HasAlpha,
		PreviewPNGB64: b64,
	}, nil
}

// PixelChannel returns the 8-bit channel value at (x,y).
func (li *LoadedImage) PixelChannel(x, y int, ch Channel) (uint8, error) {
	bounds := li.RGBA.Bounds()
	if x < 0 || y < 0 || x >= bounds.Dx() || y >= bounds.Dy() {
		return 0, fmt.Errorf("坐标越界")
	}
	i := li.RGBA.PixOffset(x, y)
	pix := li.RGBA.Pix
	switch ch {
	case ChannelR:
		return pix[i+0], nil
	case ChannelG:
		return pix[i+1], nil
	case ChannelB:
		return pix[i+2], nil
	case ChannelA:
		return pix[i+3], nil
	default:
		return 0, fmt.Errorf("无效通道: %s", ch)
	}
}
