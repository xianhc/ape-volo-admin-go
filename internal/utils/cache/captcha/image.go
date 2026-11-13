package captcha

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/big"
	"sync"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/mojocn/base64Captcha"
	"golang.org/x/image/math/fixed"
)

// 字体只解析一次：每张图都重新解析 TTF 属于浪费，且解析失败属于启动期就能发现的问题。
var (
	fontOnce sync.Once
	digitFnt *truetype.Font
)

// digitFont 返回绘制数字用的字体。
//
// 用验证码库内置的字体，避免再往仓库里塞一个 TTF 文件。
func digitFont() *truetype.Font {
	fontOnce.Do(func() {
		digitFnt = base64Captcha.DefaultEmbeddedFonts.LoadFontByName("fonts/RitaSmith.ttf")
	})
	return digitFnt
}

// RandomCode 生成指定长度的纯数字验证码。
func RandomCode(length int) string {
	if length <= 0 {
		length = 6
	}
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			buf[i] = '0'
			continue
		}
		buf[i] = byte('0' + n.Int64())
	}
	return string(buf)
}

// NewID 生成验证码 ID，16 字节随机数的十六进制表示。
//
// 用加密随机数而不是自增 ID：ID 会回传给客户端，可预测的 ID 等于把验证码空间缩小。
func NewID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", 0)
	}
	return hex.EncodeToString(buf)
}

// RenderDigits 把数字串渲染成 PNG 图片。
//
// 字体大小由 fontSize 决定，不再像位图字体那样随图片高度等比缩放——
// 同一套配置下，图片放大不会把数字一起放大，留白由 width/height 控制。
// 字形叠加穿越字符的线条与噪点，再做一次高斯模糊。
func RenderDigits(code string, width, height, fontSize int) ([]byte, error) {
	if len(code) == 0 {
		return nil, fmt.Errorf("验证码内容为空")
	}
	if width <= 0 || height <= 0 || fontSize <= 0 {
		return nil, fmt.Errorf("验证码尺寸或字号非法: %dx%d 字号 %d", width, height, fontSize)
	}

	// 先在足够高的画布上绘制，量出真实墨迹范围后再居中裁进目标尺寸。
	// 直接用字体度量算基线不靠谱：不同 TTF 的上下伸部差别很大，装饰性字体甚至会顶出画布。
	probeHeight := fontSize * 3
	probe := image.NewRGBA(image.Rect(0, 0, width, probeHeight))
	fill(probe, background)

	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFont(digitFont())
	ctx.SetFontSize(float64(fontSize))
	ctx.SetClip(probe.Bounds())
	ctx.SetDst(probe)

	scale := fixed.Int26_6(fontSize * 64) // DPI 固定 72，1pt = 1px
	baseline := float64(probeHeight - fontSize)
	cellWidth := float64(width) / float64(len(code))
	for i := 0; i < len(code); i++ {
		ctx.SetSrc(image.NewUniform(randomInk()))
		advance := float64(digitFont().HMetric(scale, digitFont().Index(rune(code[i]))).AdvanceWidth) / 64
		x := cellWidth*float64(i) + (cellWidth-advance)/2
		y := baseline + float64(randomInt(7)-3)
		if _, err := ctx.DrawString(string(code[i]), freetype.Pt(int(x), int(y))); err != nil {
			return nil, err
		}
	}

	ink, ok := inkBounds(probe)
	if !ok {
		return nil, fmt.Errorf("验证码渲染失败: 没有产生可见字形")
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	fill(img, background)
	dx := (width - ink.Dx()) / 2
	dy := (height - ink.Dy()) / 2
	draw.Draw(img, image.Rect(dx, dy, dx+ink.Dx(), dy+ink.Dy()), probe, ink.Min, draw.Src)

	drawNoise(img, width, height)
	drawInterference(img, width, height)
	img = blurDigits(img)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// blurDigits 使用 3×3 高斯核轻度模糊图像，边缘像素延用最近的画布内像素。
func blurDigits(src *image.RGBA) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	weights := [3]int{1, 2, 1}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b int
			for ky, wy := range weights {
				sy := max(bounds.Min.Y, min(bounds.Max.Y-1, y+ky-1))
				for kx, wx := range weights {
					sx := max(bounds.Min.X, min(bounds.Max.X-1, x+kx-1))
					pixel := src.RGBAAt(sx, sy)
					weight := wx * wy
					r += int(pixel.R) * weight
					g += int(pixel.G) * weight
					b += int(pixel.B) * weight
				}
			}
			dst.SetRGBA(x, y, color.RGBA{R: uint8((r + 8) / 16), G: uint8((g + 8) / 16), B: uint8((b + 8) / 16), A: 255})
		}
	}
	return dst
}

// drawNoise 在画布上铺设稀疏的单像素噪点。
func drawNoise(img *image.RGBA, width, height int) {
	count := width * height / 64
	for i := 0; i < count; i++ {
		x := randomInt(width)
		y := randomInt(height)
		drawDot(img, x, y, 0, randomInk())
	}
}

// drawInterference 叠加横向波纹和斜线，线条穿过字符所在的中部区域。
func drawInterference(img *image.RGBA, width, height int) {
	for i, count := 0, 1+randomInt(2); i < count; i++ {
		ink := randomInk()
		baseY := float64(height/4 + randomInt(max(1, height/2)))
		amplitude := float64(2 + randomInt(max(1, height/10)))
		phase := float64(randomInt(628)) / 100
		waves := 1 + float64(randomInt(15))/10
		slope := float64(randomInt(max(1, height/2)) - height/4)
		for x := 0; x < width; x++ {
			position := float64(x) / float64(max(1, width-1))
			y := baseY + slope*(position-0.5) + amplitude*math.Sin(position*waves*2*math.Pi+phase)
			drawDot(img, x, int(math.Round(y)), 0, ink)
		}
	}
	ink := randomInk()
	x0, x1 := randomInt(width), randomInt(width)
	y0, y1 := randomInt(height), randomInt(height)
	steps := max(1, int(math.Abs(float64(x1-x0))), int(math.Abs(float64(y1-y0))))
	for step := 0; step <= steps; step++ {
		x := x0 + (x1-x0)*step/steps
		y := y0 + (y1-y0)*step/steps
		drawDot(img, x, y, 0, ink)
	}
}

// drawDot 绘制实心圆点，越出画布的像素由图像边界裁掉。
func drawDot(img *image.RGBA, x, y, radius int, ink color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				img.SetRGBA(x+dx, y+dy, ink)
			}
		}
	}
}

// randomInk 返回一个偏深的随机颜色。
func randomInk() color.RGBA {
	const base = 40
	return color.RGBA{
		R: uint8(base + randomInt(120)),
		G: uint8(base + randomInt(120)),
		B: uint8(base + randomInt(120)),
		A: 255,
	}
}

// randomInt 返回 [0, n) 的随机数，n <= 0 时返回 0。
func randomInt(n int) int {
	if n <= 0 {
		return 0
	}
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(v.Int64())
}

// background 是验证码底色，浅灰比纯白更不容易被前端样式吃掉。
var background = color.RGBA{R: 245, G: 245, B: 245, A: 255}

// fill 用底色铺满整张图。
func fill(img *image.RGBA, c color.Color) {
	draw.Draw(img, img.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
}

// inkBounds 返回与底色不同的像素范围，用于把字形居中。
func inkBounds(img *image.RGBA) (image.Rectangle, bool) {
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X-1, bounds.Min.Y-1
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.RGBAAt(x, y) != background {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if maxX < minX || maxY < minY {
		return image.Rectangle{}, false
	}
	return image.Rect(minX, minY, maxX+1, maxY+1), true
}
