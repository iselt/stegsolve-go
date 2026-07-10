package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"stegsolve-go/internal/imagecore"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails application backend.
type App struct {
	ctx   context.Context
	store *imagecore.Store
}

// NewApp creates a new App.
func NewApp() *App {
	return &App{
		store: imagecore.NewStore(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// OpenImage opens a native file dialog and loads the selected image.
// Cancel returns (nil, nil) — cancellation is a normal outcome.
func (a *App) OpenImage() (*imagecore.ImageInfo, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "打开图像",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "图像文件",
				Pattern:     "*.png;*.jpg;*.jpeg;*.bmp;*.gif;*.webp",
			},
			{
				DisplayName: "所有文件",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("打开对话框失败: %w", err)
	}
	if path == "" {
		return nil, nil
	}
	return a.loadPath(path)
}

// LoadDroppedImage loads an image from a dropped file path.
func (a *App) LoadDroppedImage(path string) (*imagecore.ImageInfo, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("路径为空")
	}
	path = strings.TrimPrefix(path, "file://")
	return a.loadPath(path)
}

func (a *App) loadPath(path string) (*imagecore.ImageInfo, error) {
	img, err := imagecore.LoadFromPath(path)
	if err != nil {
		// Do not replace current image on failure
		return nil, err
	}
	info, err := img.ToInfo()
	if err != nil {
		return nil, fmt.Errorf("生成预览失败: %w", err)
	}
	a.store.Set(img)
	return info, nil
}

// RenderBitPlane returns a base64 PNG of the selected bit plane, or original
// when channel is empty / "ORIG".
func (a *App) RenderBitPlane(req imagecore.BitPlaneRequest) (string, error) {
	img, err := a.store.Get(req.ImageID)
	if err != nil {
		return "", err
	}

	ch := strings.ToUpper(string(req.Channel))
	if ch == "" || ch == "ORIG" || ch == "ORIGINAL" {
		pngBytes, err := imagecore.OriginalViewPNG(img.RGBA)
		if err != nil {
			return "", err
		}
		return imagecoreB64(pngBytes), nil
	}

	bpReq := imagecore.BitPlaneRequest{
		ImageID: req.ImageID,
		Channel: imagecore.Channel(ch),
		Bit:     req.Bit,
	}
	if err := imagecore.ValidateBitPlane(bpReq, img.HasAlpha); err != nil {
		return "", err
	}
	return imagecore.RenderBitPlaneB64(img.RGBA, bpReq.Channel, bpReq.Bit)
}

// SaveBitPlane opens a save dialog and writes the bit plane (or original) PNG.
func (a *App) SaveBitPlane(req imagecore.BitPlaneRequest) (*imagecore.SaveResult, error) {
	img, err := a.store.Get(req.ImageID)
	if err != nil {
		return nil, err
	}

	ch := strings.ToUpper(string(req.Channel))
	defaultName := defaultBitPlaneName(img.Name, ch, req.Bit)

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "导出位平面 PNG",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "PNG 图像", Pattern: "*.png"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("保存对话框失败: %w", err)
	}
	if path == "" {
		return &imagecore.SaveResult{Cancelled: true}, nil
	}
	if !strings.HasSuffix(strings.ToLower(path), ".png") {
		path += ".png"
	}

	var n int64
	if ch == "" || ch == "ORIG" || ch == "ORIGINAL" {
		pngBytes, err := imagecore.OriginalViewPNG(img.RGBA)
		if err != nil {
			return nil, err
		}
		n, err = writeAtomic(path, pngBytes)
		if err != nil {
			return nil, err
		}
	} else {
		bpReq := imagecore.BitPlaneRequest{
			ImageID: req.ImageID,
			Channel: imagecore.Channel(ch),
			Bit:     req.Bit,
		}
		if err := imagecore.ValidateBitPlane(bpReq, img.HasAlpha); err != nil {
			return nil, err
		}
		n, err = imagecore.SaveBitPlanePNG(img.RGBA, bpReq.Channel, bpReq.Bit, path)
		if err != nil {
			return nil, err
		}
	}

	return &imagecore.SaveResult{Path: path, Bytes: n}, nil
}

// PreviewLSB returns a hex/ASCII preview of the LSB extraction.
func (a *App) PreviewLSB(req imagecore.ExtractRequest) (*imagecore.ExtractPreview, error) {
	img, err := a.store.Get(req.ImageID)
	if err != nil {
		return nil, err
	}
	if err := imagecore.ValidateExtract(req, img.HasAlpha); err != nil {
		return nil, err
	}
	return imagecore.PreviewExtract(img.RGBA, req)
}

// SaveLSB streams the full extraction to a user-chosen .bin file.
func (a *App) SaveLSB(req imagecore.ExtractRequest) (*imagecore.SaveResult, error) {
	img, err := a.store.Get(req.ImageID)
	if err != nil {
		return nil, err
	}
	if err := imagecore.ValidateExtract(req, img.HasAlpha); err != nil {
		return nil, err
	}

	defaultName := defaultLSBName(img.Name)
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "导出 LSB 原始数据",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "二进制文件", Pattern: "*.bin"},
			{DisplayName: "所有文件", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("保存对话框失败: %w", err)
	}
	if path == "" {
		return &imagecore.SaveResult{Cancelled: true}, nil
	}

	n, err := imagecore.SaveExtract(img.RGBA, req, path)
	if err != nil {
		return nil, err
	}
	return &imagecore.SaveResult{Path: path, Bytes: n}, nil
}

func defaultBitPlaneName(srcName, ch string, bit int) string {
	base := strings.TrimSuffix(srcName, filepath.Ext(srcName))
	if base == "" {
		base = "image"
	}
	if ch == "" || ch == "ORIG" || ch == "ORIGINAL" {
		return base + "_original.png"
	}
	return fmt.Sprintf("%s_%s_bit%d.png", base, ch, bit)
}

func defaultLSBName(srcName string) string {
	base := strings.TrimSuffix(srcName, filepath.Ext(srcName))
	if base == "" {
		base = "image"
	}
	return base + "_lsb.bin"
}
