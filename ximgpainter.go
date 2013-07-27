// Very liberally copied the code from freetype-go (file: paint.go)
// Adapted to the *xgraphics.Image type.
package vgximg

import (
	"code.google.com/p/freetype-go/freetype/raster"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"image/color"
	"image/draw"
)

type XimgPainter struct {
	// The image to compose onto.
	Image *xgraphics.Image
	// The Porter-Duff composition operator.
	Op draw.Op
	// The 16-bit color to paint the spans.
	cr, cg, cb, ca uint32
}

// Paint satisfies the Painter interface by painting ss onto an xgraphics.Image.
func (r *XimgPainter) Paint(ss []raster.Span, done bool) {
	b := r.Image.Bounds()
	for _, s := range ss {
		if s.Y < b.Min.Y {
			continue
		}
		if s.Y >= b.Max.Y {
			return
		}
		if s.X0 < b.Min.X {
			s.X0 = b.Min.X
		}
		if s.X1 > b.Max.X {
			s.X1 = b.Max.X
		}
		if s.X0 >= s.X1 {
			continue
		}
		// This code is similar to drawGlyphOver in $GOROOT/src/pkg/image/draw/draw.go.
		ma := s.A >> 16
		const m = 1<<16 - 1
		i0 := (s.Y-r.Image.Rect.Min.Y)*r.Image.Stride + (s.X0-r.Image.Rect.Min.X)*4
		i1 := i0 + (s.X1-s.X0)*4
		if r.Op == draw.Over {
			for i := i0; i < i1; i += 4 {
				dr := uint32(r.Image.Pix[i+0])
				dg := uint32(r.Image.Pix[i+1])
				db := uint32(r.Image.Pix[i+2])
				da := uint32(r.Image.Pix[i+3])
				a := (m - (r.ca * ma / m)) * 0x101
				r.Image.Pix[i+0] = uint8((dr*a + r.cr*ma) / m >> 8)
				r.Image.Pix[i+1] = uint8((dg*a + r.cg*ma) / m >> 8)
				r.Image.Pix[i+2] = uint8((db*a + r.cb*ma) / m >> 8)
				r.Image.Pix[i+3] = uint8((da*a + r.ca*ma) / m >> 8)
			}
		} else {
			for i := i0; i < i1; i += 4 {
				r.Image.Pix[i+0] = uint8(r.cr * ma / m >> 8)
				r.Image.Pix[i+1] = uint8(r.cg * ma / m >> 8)
				r.Image.Pix[i+2] = uint8(r.cb * ma / m >> 8)
				r.Image.Pix[i+3] = uint8(r.ca * ma / m >> 8)
			}
		}
	}
}

// SetColor sets the color to paint the spans.
func (r *XimgPainter) SetColor(c color.Color) {
	r.cr, r.cg, r.cb, r.ca = c.RGBA()
}

// NewXimgPainter creates a new XimgPainter for the given image.
func NewXimgPainter(m *xgraphics.Image) *XimgPainter {
	return &XimgPainter{Image: m}
}
