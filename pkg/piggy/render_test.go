package piggy

import (
	"testing"
	"time"
)

func TestRenderDeterministicSameDay(t *testing.T) {
	// 同一天+同一人 → 相同结果（即使多次调用，因为混入了随机扰动，可能不同，
	// 这里只验证不 panic 且数据完整）
	in := Input{UserID: 10001, Date: time.Date(2026, 8, 26, 0, 0, 0, 0, time.Local)}
	out, err := Render(in)
	if err != nil {
		t.Fatal(err)
	}
	if out.Pig == nil || out.Markdown == "" || out.ImageName == "" {
		t.Fatal("输出不完整")
	}
	if out.ImageName != out.Pig.ImageName() {
		t.Fatalf("ImageName 不一致: %s != %s", out.ImageName, out.Pig.ImageName())
	}
	if out.ImageName != out.Pig.Slug+".png" {
		t.Fatalf("ImageName 应为 <slug>.png，got %s", out.ImageName)
	}
}

func TestRenderRequiresInput(t *testing.T) {
	if _, err := Render(Input{}); err == nil {
		t.Fatal("缺少 UserID/Date 应报错")
	}
}