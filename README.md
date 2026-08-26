# daily-piggy

为聊天机器人（Chatbot）设计的「每日猪猪」卡片生成依赖包。输入用户 QQ 号（可选昵称）与日期，输出一只随机猪猪的 Markdown 文本，并可进一步渲染为带图的 PNG 卡片，供机器人发送到聊天群/私聊。

纯 Go 实现，零 CGO，图片资源通过 `go:embed` 内嵌，开箱即用。

## ✨ 特性

- **Chatbot 友好 API**：一个 `piggy.RenderPNG()` 调用即可拿到图片字节，直接发消息；也提供纯 `Render()` 拿 Markdown 文本
- **每日一变**：`用户 QQ + 日期` 决定随机种子，同人同日稳定、不同日期/不同用户结果不同
- **昵称回退**：支持昵称（显示 `@昵称`），未传入时回退显示 QQ 号
- **149 种猪猪**：完整收录猪猪图鉴全部品种（名称/描述/性格分析 + 图片）
- **资源内嵌**：全部图片 `go:embed` 进二进制，无外部文件依赖
- **体积优化**：图片量化压缩（7.31 MB → 1.54 MB，省 79%），保留透明通道

## 📦 安装

```bash
go get github.com/meteor-liu-xinyu/daily-piggy
```

依赖 Go ≥ 1.25（渲染依赖 [md2img-lite](https://github.com/meteor-liu-xinyu/md2img-lite)）。

## 🚀 快速开始

```go
package main

import (
	"os"
	"time"

	"github.com/meteor-liu-xinyu/daily-piggy/pkg/piggy"
	md2img "github.com/meteor-liu-xinyu/md2img-lite"
)

func main() {
	// 1. 用 md2img 渲染 Markdown → PNG（注入图片加载器）
	renderFn := func(md []byte) ([]byte, error) {
		return md2img.Render(md,
			md2img.WithWidth(800),
			md2img.WithImageLoader(piggy.LoadImage("")), // 加载内嵌图片
		)
	}

	// 2. 生成卡片
	out, err := piggy.RenderPNG(piggy.Input{
		UserID:   123456,              // QQ 号（必填）
		Nickname: "Meteor",            // 昵称（可选，空则回退显示 QQ 号）
		Date:     time.Now(),          // 今天日期（必须由调用方传入）
	}, renderFn)
	if err != nil {
		panic(err)
	}

	// 3. 发送/保存
	os.WriteFile("pig.png", out.PNG, 0644)
}
```

### 纯文本模式（不需要图片时）

```go
out, _ := piggy.Render(piggy.Input{UserID: 123456, Date: time.Now()})
fmt.Println(out.Markdown) // Markdown 文本
fmt.Println(out.Pig.Name) // 猪猪名称
```

## 📖 API

| 类型 | 说明 |
|------|------|
| `Input{UserID int64, Nickname string, Date time.Time}` | 输入参数 |
| `Output{Pig *Pig, Markdown string, PNG []byte, ImageName string}` | 渲染结果 |
| `piggy.Render(in Input) (*Output, error)` | 生成 Markdown 文本 |
| `piggy.RenderPNG(in Input, renderFn func([]byte)([]byte, error)) (*Output, error)` | 生成 Markdown + PNG |
| `piggy.LoadImage(imagesDir string) func(string)(image.Image, error)` | 图片加载器（配合 md2img `WithImageLoader`） |

> `RenderPNG` 的 `renderFn` 参数将 md2img 调用解耦出去——可以传入任意 Markdown→PNG 渲染实现；`LoadImage("")` 加载内嵌图片，传非空路径可回退读取本地目录。

## 🎲 选猪算法

```
种子 = FNV-1a(userID | yyyy-mm-dd)   ← 确定性
再混入一个随机扰动                     ← 同人同日多次调用也有变化
最终从 149 种中选 1 种
```

同样的 `(userID, date)` 组合结果高度稳定，跨天自动换新。

## 🐷 数据与素材

- 数据来源：[pig.felislab.cc](https://pig.felislab.cc) 猪猪图鉴（149 种），收录于 `pkg/piggy/data_gen.go`（由 `gen_data.py` 生成，勿手改）
- 图片素材：网站原始 PNG（240×240，含透明通道），量化压缩后内嵌

## 🖥️ 命令行示例

```bash
go run ./examples/pigcard -id 123456 -date 2026-08-26 -nick Meteor -out pig.png
# 指定渲染字体（默认走系统字体，Windows 微软雅黑）：
go run ./examples/pigcard -id 123456 -date 2026-08-26 -font-regular "C:\Windows\Fonts\msyh.ttc"
```

## 🧪 测试

```bash
go test ./...
```

覆盖：数据完整性（149 条）、全部嵌入图片可解码、图片加载回退、入参校验。

## 📁 目录结构

```
daily-piggy/
├── pkg/piggy/          # 核心依赖包
│   ├── piggy.go        # 主逻辑：Render / RenderPNG
│   ├── data_gen.go     # 149 种猪猪数据（生成，勿手改）
│   ├── images.go       # 内嵌图片加载（embed + LoadImage）
│   └── images/         # 猪猪图片（go:embed 内嵌）
├── examples/pigcard/   # CLI 示例
├── gen_data.py         # 数据生成脚本（从 PDF 提取的 txt 生成 data_gen.go）
└── pig_compendium.pdf  # 数据源：猪猪图鉴
```

## 🔒 License

[MIT](./LICENSE)