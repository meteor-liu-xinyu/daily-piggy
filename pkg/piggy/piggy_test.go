package piggy

import (
	"testing"
)

func TestImageName(t *testing.T) {
	p := pigs[0]
	if got := p.ImageName(); got != "human.png" {
		t.Fatalf("ImageName() = %q, want human.png", got)
	}
	if len(pigs) != 149 {
		t.Fatalf("pigs 数量 = %d, want 149", len(pigs))
	}
}

func TestEmbeddedImagesLoadable(t *testing.T) {
	// 每只猪猪的嵌入图片都必须能加载并解码
	loader := LoadImage("")
	for i, p := range pigs {
		img, err := loader(p.ImageName())
		if err != nil {
			t.Fatalf("pig #%d (%s): 加载 %s 失败: %v", p.ID, p.Slug, p.ImageName(), err)
		}
		b := img.Bounds()
		if b.Dx() <= 0 || b.Dy() <= 0 {
			t.Fatalf("pig #%d: 图片尺寸无效 %v", p.ID, b)
		}
		_ = i
	}
}

func TestLoadImageFallback(t *testing.T) {
	// 不存在图片应返回错误
	loader := LoadImage("")
	if _, err := loader("nope.png"); err == nil {
		t.Fatal("加载不存在的图片应报错")
	}
}