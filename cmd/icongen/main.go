// Command icongen draws the app's icons: a chip outlined in white on a
// dark oil background, labelled with silicon's symbol and atomic number.
//
//	go run ./cmd/icongen            # writes the files below into assets/
//
// icon.png                  the icon, artwork in the middle so launchers that
//
//	crop to a circle don't cut it
//
// icon-ios.png              iOS's: opaque and full-bleed with a larger chip,
//
//	since iOS only rounds the corners
//
// android-foreground.png    the adaptive icon's chip on transparency (Android
//
//	8+), inside the 66dp safe zone of the 108dp canvas;
//	also the themed icon's shape (Android 13+)
//
// android-background.png    the adaptive icon's oil background
// launch.png                the chip alone, centred on iOS's launch screen
package main

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/mehran9675/silicondb/assets"
)

const (
	size = 1024
	ss   = 4 // supersampling factor
	big  = size * ss
)

func main() {
	dir := flag.String("dir", "assets", "output folder")
	flag.Parse()

	face := loadFont()
	write(filepath.Join(*dir, "icon.png"), render(chip(face, 1), true))
	write(filepath.Join(*dir, "icon-ios.png"), render(chip(face, 1.38), true))
	write(filepath.Join(*dir, "android-foreground.png"), render(chip(face, 1.2), false))
	write(filepath.Join(*dir, "android-background.png"), render(image.NewAlpha(image.Rect(0, 0, big, big)), true))
	write(filepath.Join(*dir, "launch.png"), shrink(render(chip(face, 1.9), false), 600))
}

// chip draws the chip and its label into a supersampled mask, scaled by s
// around the centre (1 is the launcher-safe size).
func chip(f *opentype.Font, s float64) *image.Alpha {
	m := image.NewAlpha(image.Rect(0, 0, big, big))
	t := func(v float64) float64 { return size/2 + (v-size/2)*s }
	const (
		lo, hi = 340.0, 684.0 // package body
		radius = 46.0
		stroke = 22.0
		pinLen = 64.0
		pinW   = 22.0
	)
	box := func(x0, y0, x1, y1, r float64, a uint8) {
		fillRoundRect(m, t(x0), t(y0), t(x1), t(y1), r*s, a)
	}
	box(lo, lo, hi, hi, radius, 255)
	box(lo+stroke, lo+stroke, hi-stroke, hi-stroke, radius-stroke, 0)
	for i := 0; i < 4; i++ {
		c := lo + (hi-lo)*(float64(i)+1)/5
		box(c-pinW/2, lo-pinLen, c+pinW/2, lo+1, pinW/2, 255)
		box(c-pinW/2, hi-1, c+pinW/2, hi+pinLen, pinW/2, 255)
		box(lo-pinLen, c-pinW/2, lo+1, c+pinW/2, pinW/2, 255)
		box(hi-1, c-pinW/2, hi+pinLen, c+pinW/2, pinW/2, 255)
	}
	text(m, f, "Si", 168*s, t(512), t(590), true)
	text(m, f, "14", 54*s, t(392), t(438), false)
	return m
}

// render paints the mask white, over the oil background when opaque.
func render(mask *image.Alpha, opaque bool) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			a := coverage(mask, x, y)
			if !opaque {
				img.SetNRGBA(x, y, color.NRGBA{255, 255, 255, uint8(255 * a)})
				continue
			}
			r, g, b := background(float64(x)+0.5, float64(y)+0.5)
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(r + (255-r)*a),
				G: uint8(g + (255-g)*a),
				B: uint8(b + (255-b)*a),
				A: 255,
			})
		}
	}
	return img
}

// background is dark oil: near-black with a faint green-blue sheen
// toward the top left and a warmer depth toward the bottom right.
func background(x, y float64) (r, g, b float64) {
	t := (x + y) / (2 * size)
	r, g, b = lerp(22, 12, t), lerp(30, 14, t), lerp(30, 17, t)
	sheen := math.Max(0, 1-math.Hypot(x-size*0.3, y-size*0.25)/(size*0.75))
	sheen *= sheen
	r, g, b = r+10*sheen, g+24*sheen, b+26*sheen
	glow := math.Max(0, 1-math.Hypot(x-size*0.5, y-size*0.5)/(size*0.45))
	r, g, b = r+6*glow, g+8*glow, b+10*glow
	return r, g, b
}

func loadFont() *opentype.Font {
	ttf, err := assets.Fonts.ReadFile("fonts/Inter-Bold.ttf")
	if err != nil {
		log.Fatal(err)
	}
	f, err := opentype.Parse(ttf)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

// text draws s at size px with its baseline at y; centered centers it on
// x, otherwise x is its left edge. Coordinates are in output pixels.
func text(m *image.Alpha, f *opentype.Font, s string, px, x, y float64, centered bool) {
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: px * ss, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		log.Fatal(err)
	}
	defer face.Close()
	d := &font.Drawer{Dst: m, Src: image.NewUniform(color.Alpha{A: 255}), Face: face}
	w := d.MeasureString(s)
	startX := fixed.I(int(x * ss))
	if centered {
		startX -= w / 2
	}
	d.Dot = fixed.Point26_6{X: startX, Y: fixed.I(int(y * ss))}
	d.DrawString(s)
}

// fillRoundRect sets the pixels of a rounded rectangle, given in output
// pixels, to a.
func fillRoundRect(m *image.Alpha, x0, y0, x1, y1, r float64, a uint8) {
	X0, Y0, X1, Y1, R := x0*ss, y0*ss, x1*ss, y1*ss, r*ss
	for y := int(Y0); y < int(math.Ceil(Y1)); y++ {
		for x := int(X0); x < int(math.Ceil(X1)); x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			cx := math.Max(X0+R, math.Min(fx, X1-R))
			cy := math.Max(Y0+R, math.Min(fy, Y1-R))
			if math.Hypot(fx-cx, fy-cy) <= R {
				m.SetAlpha(x, y, color.Alpha{A: a})
			}
		}
	}
}

// coverage averages the supersampled mask over one output pixel.
func coverage(m *image.Alpha, x, y int) float64 {
	sum := 0
	for j := 0; j < ss; j++ {
		off := m.PixOffset(x*ss, y*ss+j)
		for i := 0; i < ss; i++ {
			sum += int(m.Pix[off+i])
		}
	}
	return float64(sum) / float64(255*ss*ss)
}

// shrink scales a square image down to n pixels by averaging, keeping
// alpha premultiplied while averaging so edges stay clean.
func shrink(src *image.NRGBA, n int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, n, n))
	k := float64(size) / float64(n)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			var r, g, b, a, cnt float64
			for sy := int(float64(y) * k); sy < int(float64(y+1)*k); sy++ {
				for sx := int(float64(x) * k); sx < int(float64(x+1)*k); sx++ {
					c := src.NRGBAAt(sx, sy)
					al := float64(c.A) / 255
					r, g, b, a = r+float64(c.R)*al, g+float64(c.G)*al, b+float64(c.B)*al, a+al
					cnt++
				}
			}
			if a > 0 {
				dst.SetNRGBA(x, y, color.NRGBA{uint8(r / a), uint8(g / a), uint8(b / a), uint8(255 * a / cnt)})
			}
		}
	}
	return dst
}

func write(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }
