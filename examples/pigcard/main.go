// 示例：用 piggy 包生成今日猪猪卡片图片。
//
// 用法：
//
//	go run ./examples/pigcard -id 123456 -date 2026-08-26 -out pig.png
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/meteor-liu-xinyu/daily-piggy/pkg/piggy"
	md2img "github.com/meteor-liu-xinyu/md2img-lite"
)

func main() {
	userID := flag.Int64("id", 0, "用户 QQ 号（必填）")
	nickname := flag.String("nick", "", "用户昵称（可选，为空时回退显示 QQ 号）")
	dateStr := flag.String("date", "", "今天日期，格式 yyyy-mm-dd（必填）")
	out := flag.String("out", "pig.png", "输出 PNG 文件路径")
	width := flag.Int("width", 800, "画布宽度")
	padding := flag.Int("padding", 48, "左右内边距")
	fontRegular := flag.String("font-regular", "", "正文字体文件路径（可选）")
	fontMono := flag.String("font-mono", "", "等宽字体文件路径（可选）")
	flag.Parse()

	if *userID <= 0 || *dateStr == "" {
		fmt.Fprintln(os.Stderr, "用法：pigcard -id <QQ号> -date <yyyy-mm-dd> [-out pig.png]")
		os.Exit(1)
	}

	date, err := time.ParseInLocation("2006-01-02", *dateStr, time.Local)
	if err != nil {
		fmt.Fprintf(os.Stderr, "日期格式错误：%v\n", err)
		os.Exit(1)
	}

	// 1. 构造 md2img.Render 适配函数（注入图片加载器 + 选项）
	renderFn := func(md []byte) ([]byte, error) {
		opts := []md2img.Option{
			md2img.WithWidth(*width),
			md2img.WithPadding(*padding),
			md2img.WithLineHeight(1.8),
			md2img.WithImageLoader(piggy.LoadImage("")), // 嵌入图片
		}
		if *fontRegular != "" {
			opts = append(opts, md2img.WithFontPath("regular", *fontRegular))
		}
		if *fontMono != "" {
			opts = append(opts, md2img.WithFontPath("mono", *fontMono))
		}
		return md2img.Render(md, opts...)
	}

	// 2. 渲染
	o, err := piggy.RenderPNG(piggy.Input{UserID: *userID, Nickname: *nickname, Date: date}, renderFn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "渲染失败：%v\n", err)
		os.Exit(1)
	}

	// 3. 输出
	if err := os.WriteFile(*out, o.PNG, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "写入文件失败：%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ 已生成：%s（%d bytes）\n", *out, len(o.PNG))
	fmt.Printf("猪猪：%s（%s） #%d\n", o.Pig.Name, o.Pig.Slug, o.Pig.ID)
	fmt.Println("\n--- Markdown ---")
	fmt.Println(o.Markdown)
}
