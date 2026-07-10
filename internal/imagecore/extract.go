package imagecore

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
)

// ValidateExtract validates an extraction request against image state.
func ValidateExtract(req ExtractRequest, hasAlpha bool) error {
	if err := validateMask(req.MaskA, ChannelA, hasAlpha); err != nil {
		return err
	}
	if err := validateMask(req.MaskR, ChannelR, true); err != nil {
		return err
	}
	if err := validateMask(req.MaskG, ChannelG, true); err != nil {
		return err
	}
	if err := validateMask(req.MaskB, ChannelB, true); err != nil {
		return err
	}

	if SelectedBitCount(req) == 0 {
		return fmt.Errorf("未选择任何提取位")
	}

	switch req.Traverse {
	case TraverseRow, TraverseCol:
	default:
		return fmt.Errorf("无效遍历顺序: %s", req.Traverse)
	}

	switch req.BitOrder {
	case BitOrderLSBFirst, BitOrderMSBFirst:
	default:
		return fmt.Errorf("无效位序: %s", req.BitOrder)
	}

	if len(req.ChannelOrder) == 0 {
		return fmt.Errorf("通道顺序不能为空")
	}

	seen := map[Channel]bool{}
	for _, ch := range req.ChannelOrder {
		switch ch {
		case ChannelR, ChannelG, ChannelB:
		case ChannelA:
			if !hasAlpha {
				return fmt.Errorf("当前图像无有效 Alpha 通道，不能包含 A")
			}
		default:
			return fmt.Errorf("无效通道: %s", ch)
		}
		if seen[ch] {
			return fmt.Errorf("通道顺序中存在重复: %s", ch)
		}
		seen[ch] = true
	}

	// Every channel with a non-zero mask must appear in channel order.
	for _, ch := range []struct {
		c Channel
		m uint8
	}{
		{ChannelA, req.MaskA},
		{ChannelR, req.MaskR},
		{ChannelG, req.MaskG},
		{ChannelB, req.MaskB},
	} {
		if ch.m != 0 && !seen[ch.c] {
			return fmt.Errorf("已选择通道 %s 的位，但未出现在通道顺序中", ch.c)
		}
	}

	return nil
}

func validateMask(mask uint8, ch Channel, allowed bool) error {
	if mask != 0 && !allowed {
		return fmt.Errorf("当前图像无有效 Alpha 通道，不能选择 %s", ch)
	}
	return nil
}

// SelectedBitCount returns how many bits are selected across all masks.
func SelectedBitCount(req ExtractRequest) int {
	return popcount(req.MaskA) + popcount(req.MaskR) + popcount(req.MaskG) + popcount(req.MaskB)
}

func popcount(v uint8) int {
	n := 0
	for v != 0 {
		n += int(v & 1)
		v >>= 1
	}
	return n
}

// maskFor returns the bit mask for a channel.
func maskFor(req ExtractRequest, ch Channel) uint8 {
	switch ch {
	case ChannelA:
		return req.MaskA
	case ChannelR:
		return req.MaskR
	case ChannelG:
		return req.MaskG
	case ChannelB:
		return req.MaskB
	default:
		return 0
	}
}

// bitsForChannel returns selected bit indices in scan order for one channel.
// LSB first: 0,1,2,...  MSB first: 7,6,5,...
func bitsForChannel(mask uint8, order BitOrder) []int {
	var bits []int
	if order == BitOrderMSBFirst {
		for b := 7; b >= 0; b-- {
			if mask&(1<<uint(b)) != 0 {
				bits = append(bits, b)
			}
		}
	} else {
		for b := 0; b <= 7; b++ {
			if mask&(1<<uint(b)) != 0 {
				bits = append(bits, b)
			}
		}
	}
	return bits
}

// bitStreamWriter packs bits MSB-first into bytes (first extracted bit → bit7).
type bitStreamWriter struct {
	w      io.Writer
	cur    uint8
	filled int // bits filled in cur (0..7), filled from bit7 downward
	total  int64
	// optional capture for preview
	capture    []byte
	captureMax int
}

func newBitStreamWriter(w io.Writer, captureMax int) *bitStreamWriter {
	return &bitStreamWriter{w: w, captureMax: captureMax}
}

func (b *bitStreamWriter) WriteBit(bit uint8) error {
	// place into next free bit from MSB
	shift := 7 - b.filled
	if bit != 0 {
		b.cur |= 1 << uint(shift)
	}
	b.filled++
	if b.filled == 8 {
		return b.flushByte()
	}
	return nil
}

func (b *bitStreamWriter) flushByte() error {
	if b.w != nil {
		if _, err := b.w.Write([]byte{b.cur}); err != nil {
			return err
		}
	}
	if b.captureMax > 0 && len(b.capture) < b.captureMax {
		b.capture = append(b.capture, b.cur)
	}
	b.total++
	b.cur = 0
	b.filled = 0
	return nil
}

// Close pads remaining bits with zeros on the low side.
func (b *bitStreamWriter) Close() error {
	if b.filled > 0 {
		return b.flushByte()
	}
	return nil
}

// ExtractTo writes extracted bytes to w and returns total bytes written.
// If captureMax > 0, also keeps the first captureMax bytes for preview.
func ExtractTo(src *image.RGBA, req ExtractRequest, w io.Writer, captureMax int) (total int64, capture []byte, err error) {
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width == 0 || height == 0 {
		return 0, nil, fmt.Errorf("空图像")
	}

	// Precompute per-channel bit lists in channel order (skip empty masks).
	type chBits struct {
		offset int
		bits   []int
	}
	var plan []chBits
	for _, ch := range req.ChannelOrder {
		m := maskFor(req, ch)
		if m == 0 {
			continue
		}
		bits := bitsForChannel(m, req.BitOrder)
		if len(bits) == 0 {
			continue
		}
		plan = append(plan, chBits{offset: channelOffset(ch), bits: bits})
	}
	if len(plan) == 0 {
		return 0, nil, fmt.Errorf("未选择任何提取位")
	}

	bw := newBitStreamWriter(w, captureMax)
	srcPix := src.Pix
	stride := src.Stride

	emitPixel := func(x, y int) error {
		base := y*stride + x*4
		for _, p := range plan {
			v := srcPix[base+p.offset]
			for _, bit := range p.bits {
				var b uint8
				if v&(1<<uint(bit)) != 0 {
					b = 1
				}
				if err := bw.WriteBit(b); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if req.Traverse == TraverseCol {
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				if err := emitPixel(x, y); err != nil {
					return bw.total, bw.capture, err
				}
			}
		}
	} else {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if err := emitPixel(x, y); err != nil {
					return bw.total, bw.capture, err
				}
			}
		}
	}

	if err := bw.Close(); err != nil {
		return bw.total, bw.capture, err
	}
	return bw.total, bw.capture, nil
}

// ExtractBytes runs extraction fully into memory (for tests / small images).
func ExtractBytes(src *image.RGBA, req ExtractRequest) ([]byte, error) {
	var buf bytes.Buffer
	total, _, err := ExtractTo(src, req, &buf, 0)
	if err != nil {
		return nil, err
	}
	if int64(buf.Len()) != total {
		return nil, fmt.Errorf("内部错误: 写入长度不一致")
	}
	return buf.Bytes(), nil
}

// BuildPreview constructs a hex dump preview from captured bytes.
func BuildPreview(capture []byte, total int64) *ExtractPreview {
	limit := PreviewByteLimit
	if len(capture) > limit {
		capture = capture[:limit]
	}
	rows := make([]HexRow, 0, (len(capture)+15)/16)
	for i := 0; i < len(capture); i += 16 {
		end := i + 16
		if end > len(capture) {
			end = len(capture)
		}
		chunk := capture[i:end]
		rows = append(rows, HexRow{
			Offset: i,
			Hex:    formatHex(chunk),
			ASCII:  formatASCII(chunk),
		})
	}
	return &ExtractPreview{
		Rows:         rows,
		TotalBytes:   total,
		PreviewBytes: len(capture),
		Truncated:    total > int64(len(capture)),
	}
}

func formatHex(b []byte) string {
	const hexdigits = "0123456789ABCDEF"
	// "XX XX ..." with spaces
	if len(b) == 0 {
		return ""
	}
	out := make([]byte, 0, len(b)*3-1)
	for i, v := range b {
		if i > 0 {
			out = append(out, ' ')
		}
		out = append(out, hexdigits[v>>4], hexdigits[v&0x0f])
	}
	return string(out)
}

func formatASCII(b []byte) string {
	out := make([]byte, len(b))
	for i, v := range b {
		if v >= 32 && v <= 126 {
			out[i] = v
		} else {
			out[i] = '.'
		}
	}
	return string(out)
}

// SaveExtract streams extraction to path via temp file then atomic rename.
func SaveExtract(src *image.RGBA, req ExtractRequest, path string) (int64, error) {
	dir := ""
	if i := lastSlash(path); i >= 0 {
		dir = path[:i]
	}
	tmp, err := os.CreateTemp(dir, ".stegsolve-lsb-*.tmp")
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

	total, _, err := ExtractTo(src, req, tmp, 0)
	if err != nil {
		tmp.Close()
		return 0, err
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
	return total, nil
}

// PreviewExtract runs extraction capturing only the preview prefix.
// Uses io.Discard for the full stream so memory stays bounded.
func PreviewExtract(src *image.RGBA, req ExtractRequest) (*ExtractPreview, error) {
	total, capture, err := ExtractTo(src, req, io.Discard, PreviewByteLimit)
	if err != nil {
		return nil, err
	}
	return BuildPreview(capture, total), nil
}
