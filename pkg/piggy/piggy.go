// Package piggy 根据用户 QQ 和日期，随机抽取一只猪猪，生成文本+图片的 Markdown，
// 并可调用 md2img-lite 渲染为 PNG 图片。
//
// 用法：
//
//	out, err := piggy.Render(ctx, piggy.Input{
//	    UserID: 123456,
//	    Date:   time.Date(2026, 8, 26, 0, 0, 0, 0, time.Local),
//	})
//	if err != nil { ... }
//	os.WriteFile("pig.png", out.PNG, 0644)
//	fmt.Println(out.Markdown)
package piggy

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Input 是渲染猪猪卡片所需的输入。
type Input struct {
	UserID   int64     // QQ 号
	Nickname string    // 用户昵称（可选，为空时回退显示 QQ 号）
	Date     time.Time // 今天日期（调用方传入，不内置 time.Now）
}

// Output 是渲染结果。
type Output struct {
	Pig       *Pig   // 抽中的猪猪
	Markdown  string // 生成的 Markdown 文本（含图片引用）
	PNG       []byte // 渲染出的 PNG 图片字节
	ImageName string // 图片文件名（img_NNN.jpg）
}

// Render 根据 userID + date + 随机数抽取一只猪猪，生成 Markdown 并渲染为 PNG。
//
// 选猪算法：用 (userID, date) 生成一个确定性的种子，再加一个随机扰动，
// 最终从 149 只猪猪中选出一只，保证「同一天同一人大致稳定、不同天/不同人有变化」。
func Render(in Input) (*Output, error) {
	if in.Date.IsZero() {
		return nil, errors.New("piggy: date is required")
	}
	if in.UserID <= 0 {
		return nil, errors.New("piggy: userID must be positive")
	}
	if len(pigs) == 0 {
		return nil, errors.New("piggy: no pig data loaded")
	}

	// 1. 确定性种子：FNV-1a(userID, yyyy-mm-dd)
	seed := seedFromUserAndDate(in.UserID, in.Date)

	// 2. 加入随机扰动：再叠加一个 rand 随机量
	rng := rand.New(rand.NewSource(seed))
	// 额外混入一个随机数，让「同一人同一天多次调用」也可能不同
	jitter := rng.Int63()
	finalSeed := seed ^ mix(jitter)
	rng2 := rand.New(rand.NewSource(finalSeed))

	// 3. 从 149 只中选一只
	idx := rng2.Intn(len(pigs))
	pig := pigs[idx]

	// 4. 生成 Markdown
	imgName := pig.ImageName()
	md := buildMarkdown(in, pig, imgName)

	return &Output{
		Pig:       &pig,
		Markdown:  md,
		ImageName: imgName,
	}, nil
}

// RenderPNG 在 Render 基础上进一步将 Markdown 渲染为 PNG 图片字节。
// renderFn 是 md2img.Render 的调用（避免在本包硬依赖 md2img，便于测试/解耦）。
// imagesDir 用于查找图片文件；为空则不加载图片（md2img 会回退显示 [alt]）。
func RenderPNG(in Input, renderFn func(md []byte) ([]byte, error)) (*Output, error) {
	out, err := Render(in)
	if err != nil {
		return nil, err
	}
	if renderFn == nil {
		return nil, errors.New("piggy: renderFn is nil")
	}
	png, err := renderFn([]byte(out.Markdown))
	if err != nil {
		return nil, fmt.Errorf("piggy: render markdown: %w", err)
	}
	out.PNG = png
	return out, nil
}

// seedFromUserAndDate 用 userID + 日期字符串生成一个稳定的 int64 种子。
func seedFromUserAndDate(userID int64, date time.Time) int64 {
	h := fnv.New64a()
	h.Write([]byte(strconv.FormatInt(userID, 10)))
	h.Write([]byte{'|'})
	h.Write([]byte(date.Format("2006-01-02")))
	return int64(h.Sum64())
}

// mix 简单位运算混淆，避免高低位差异过大。
func mix(x int64) int64 {
	x ^= x >> 31
	return x
}

// buildMarkdown 构造猪猪卡片的 Markdown 文本。
func buildMarkdown(in Input, pig Pig, imgName string) string {
	dateStr := in.Date.Format("2006-01-02")
	var b []byte
	b = append(b, `# 今日猪猪`...)
	b = append(b, '\n', '\n')
	b = append(b, `![`...)
	b = append(b, pig.Name...)
	b = append(b, `](`...)
	b = append(b, imgName...)
	b = append(b, ')', '\n', '\n')
	b = append(b, `> **用户**：`...)
	if in.Nickname != "" {
		// 有昵称显示 @昵称
		b = append(b, '@')
		b = append(b, in.Nickname...)
	} else {
		// 无昵称回退显示 QQ 号
		b = append(b, strconv.FormatInt(in.UserID, 10)...)
	}
	b = append(b, `  `...)
	b = append(b, '\n')
	b = append(b, `> **日期**：`...)
	b = append(b, dateStr...)
	b = append(b, ' ', '\n', '\n')
	b = append(b, `## `...)
	b = append(b, pig.Name...)
	b = append(b, ` *(`...)
	b = append(b, pig.Slug...)
	b = append(b, `)*`...)
	b = append(b, '\n', '\n')
	appendQuoteBlock(&b, `**描述**：`+pig.Desc)
	b = append(b, '\n', '\n')
	appendQuoteBlock(&b, `**性格分析**：`+pig.Analysis)
	b = append(b, '\n')
	return string(b)
}

// appendQuoteBlock 将文本以引用块形式追加（多行自动加 "> " 前缀）。
func appendQuoteBlock(b *[]byte, text string) {
	for i, line := range strings.Split(text, "\n") {
		if line == "" {
			continue
		}
		if i > 0 {
			*b = append(*b, '\n')
		}
		*b = append(*b, '>', ' ')
		*b = append(*b, line...)
	}
}
