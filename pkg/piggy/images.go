package piggy

import (
	"embed"
	"fmt"
	"image"
	_ "image/jpeg" // 注册 JPEG 解码器
	_ "image/png"  // 注册 PNG 解码器
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed images/*.png
var imgFS embed.FS

// LoadImage 根据 md 中引用的图片文件名（如 human.png）加载嵌入的猪猪图片。
// 本函数用作 md2img.WithImageLoader 的回调。
// 若图片不在嵌入资源中，则尝试从 imagesDir（若设置）读取。
func LoadImage(imagesDir string) func(src string) (image.Image, error) {
	return func(src string) (image.Image, error) {
		name := filepath.Base(src)

		// 1. 先查嵌入资源
		data, err := imgFS.ReadFile("images/" + name)
		if err == nil {
			im, _, derr := image.Decode(strings.NewReader(string(data)))
			if derr != nil {
				return nil, fmt.Errorf("piggy: decode embedded %s: %w", name, derr)
			}
			return im, nil
		}

		// 2. 回退到本地目录（imagesDir 非空时）
		if imagesDir != "" {
			path := filepath.Join(imagesDir, name)
			f, oerr := os.Open(path)
			if oerr == nil {
				defer f.Close()
				im, _, derr := image.Decode(f)
				if derr != nil {
					return nil, fmt.Errorf("piggy: decode %s: %w", path, derr)
				}
				return im, nil
			}
		}
		return nil, fs.ErrNotExist
	}
}
