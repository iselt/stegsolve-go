package imagecore

// Channel identifies an RGBA channel.
type Channel string

const (
	ChannelA Channel = "A"
	ChannelR Channel = "R"
	ChannelG Channel = "G"
	ChannelB Channel = "B"
)

// TraverseOrder controls pixel scan direction.
type TraverseOrder string

const (
	TraverseRow TraverseOrder = "row"
	TraverseCol TraverseOrder = "col"
)

// BitOrder controls bit packing direction within a channel.
type BitOrder string

const (
	BitOrderLSBFirst BitOrder = "lsb" // bit0 → bit7 within channel
	BitOrderMSBFirst BitOrder = "msb" // bit7 → bit0 within channel
)

// ImageInfo is metadata returned after loading an image.
type ImageInfo struct {
	ImageID      string `json:"imageId"`
	Name         string `json:"name"`
	Format       string `json:"format"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int64  `json:"fileSize"`
	HasAlpha     bool   `json:"hasAlpha"`
	PreviewPNGB64 string `json:"previewPngB64"`
}

// BitPlaneRequest selects a single channel bit plane.
// Bit 0 is LSB, bit 7 is MSB.
type BitPlaneRequest struct {
	ImageID string  `json:"imageId"`
	Channel Channel `json:"channel"`
	Bit     int     `json:"bit"`
}

// ExtractRequest configures LSB (multi-bit) extraction.
// Masks use bit N set when channel bit N is selected (bit0 = LSB).
type ExtractRequest struct {
	ImageID      string        `json:"imageId"`
	MaskA        uint8         `json:"maskA"`
	MaskR        uint8         `json:"maskR"`
	MaskG        uint8         `json:"maskG"`
	MaskB        uint8         `json:"maskB"`
	ChannelOrder []Channel     `json:"channelOrder"`
	Traverse     TraverseOrder `json:"traverse"`
	BitOrder     BitOrder      `json:"bitOrder"`
}

// HexRow is one line of hex dump preview.
type HexRow struct {
	Offset int    `json:"offset"`
	Hex    string `json:"hex"`
	ASCII  string `json:"ascii"`
}

// ExtractPreview is a bounded hex/ASCII preview of extracted bytes.
type ExtractPreview struct {
	Rows         []HexRow `json:"rows"`
	TotalBytes   int64    `json:"totalBytes"`
	PreviewBytes int      `json:"previewBytes"`
	Truncated    bool     `json:"truncated"`
}

// SaveResult reports the outcome of a save dialog / write.
type SaveResult struct {
	Cancelled bool   `json:"cancelled"`
	Path      string `json:"path"`
	Bytes     int64  `json:"bytes"`
}

// MaxPixels is the hard limit for a single loaded image.
const MaxPixels = 50_000_000

// PreviewByteLimit is the number of extracted bytes shown in the hex preview.
const PreviewByteLimit = 4096
