package main

import (
	"encoding/base64"
	"fmt"
	"os"
)

func imagecoreB64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func writeAtomic(path string, data []byte) (int64, error) {
	dir := ""
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			dir = path[:i]
			break
		}
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
		return 0, err
	}
	if err := tmp.Close(); err != nil {
		return 0, err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return 0, fmt.Errorf("替换目标文件失败: %w", err)
	}
	cleanup = false
	return int64(n), nil
}
