// Copies the code from vgimg.go (plotinum package) to create
// a new concrete type that satisfies vg.Canvas.
package vgximg

import (
	"github.com/llgcode/draw2d"
	"code.google.com/p/plotinum/vg"
	"fmt"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
	"image"
	"image/color"
	"image/draw"
)

// dpi is the number of dots per inch.
const dpi = 96

// Canvas implements the vg.Canvas interface,
// drawing to an image.Image using draw2d.
type XImgCanvas struct {
	gc draw2d.GraphicContext
	//img   image.Image
	w, h  vg.Length
	color []color.Color

	// width is the current line width.
	width vg.Length

	// X window values
	x    *xgbutil.XUtil
	ximg *xgraphics.Image
	wid  *xwindow.Window
}

// New returns a new image canvas with
// the size specified  rounded up to the
// nearest pixel.
func New(width, height vg.Length, name string) (*XImgCanvas, error) {
	w := width.Inches() * dpi
	h := height.Inches() * dpi
	img := image.NewRGBA(image.Rect(0, 0, int(w+0.5), int(h+0.5)))

	return NewImage(img, name)
}

// NewImage returns a new image canvas
// that draws to the given image.  The
// minimum point of the given image
// should probably be 0,0.
func NewImage(img draw.Image, name string) (*XImgCanvas, error) {
	w := float64(img.Bounds().Max.X - img.Bounds().Min.X)
	h := float64(img.Bounds().Max.Y - img.Bounds().Min.Y)

	X, err := xgbutil.NewConn()
	if err != nil {
		return nil, err
	}
	keybind.Initialize(X)
	ximg := xgraphics.New(X, image.Rect(0, 0, int(w), int(h)))
	err = ximg.CreatePixmap()
	if err != nil {
		return nil, err
	}
	painter := NewXimgPainter(ximg)
	gc := draw2d.NewGraphicContextWithPainter(ximg, painter)
	gc.SetDPI(dpi)
	gc.Scale(1, -1)
	gc.Translate(0, -h)

	wid := ximg.XShowExtra(name, true)
	go func() {
		xevent.Main(X)
	}()

	c := &XImgCanvas{
		gc:    gc,
		w:     vg.Inches(w / dpi),
		h:     vg.Inches(h / dpi),
		color: []color.Color{color.Black},
		x:     X,
		ximg:  ximg,
		wid:   wid,
	}
	vg.Initialize(c)
	return c, nil
}

func (c *XImgCanvas) Paint() {
	c.ximg.XDraw()
	c.ximg.XPaint(c.wid.Id)
}

func (c *XImgCanvas) Size() (w, h vg.Length) {
	return c.w, c.h
}

func (c *XImgCanvas) SetLineWidth(w vg.Length) {
	c.width = w
	c.gc.SetLineWidth(w.Dots(c))
}

func (c *XImgCanvas) SetLineDash(ds []vg.Length, offs vg.Length) {
	dashes := make([]float64, len(ds))
	for i, d := range ds {
		dashes[i] = d.Dots(c)
	}
	c.gc.SetLineDash(dashes, offs.Dots(c))
}

func (c *XImgCanvas) SetColor(clr color.Color) {
	if clr == nil {
		clr = color.Black
	}
	c.gc.SetFillColor(clr)
	c.gc.SetStrokeColor(clr)
	c.color[len(c.color)-1] = clr
}

func (c *XImgCanvas) Rotate(t float64) {
	c.gc.Rotate(t)
}

func (c *XImgCanvas) Translate(x, y vg.Length) {
	c.gc.Translate(x.Dots(c), y.Dots(c))
}

func (c *XImgCanvas) Scale(x, y float64) {
	c.gc.Scale(x, y)
}

func (c *XImgCanvas) Push() {
	c.color = append(c.color, c.color[len(c.color)-1])
	c.gc.Save()
}

func (c *XImgCanvas) Pop() {
	c.color = c.color[:len(c.color)-1]
	c.gc.Restore()
}

func (c *XImgCanvas) Stroke(p vg.Path) {
	if c.width == 0 {
		return
	}
	c.outline(p)
	c.gc.Stroke()
}

func (c *XImgCanvas) Fill(p vg.Path) {
	c.outline(p)
	c.gc.Fill()
}

func (c *XImgCanvas) outline(p vg.Path) {
	c.gc.BeginPath()
	for _, comp := range p {
		switch comp.Type {
		case vg.MoveComp:
			c.gc.MoveTo(comp.X.Dots(c), comp.Y.Dots(c))

		case vg.LineComp:
			c.gc.LineTo(comp.X.Dots(c), comp.Y.Dots(c))

		case vg.ArcComp:
			c.gc.ArcTo(comp.X.Dots(c), comp.Y.Dots(c),
				comp.Radius.Dots(c), comp.Radius.Dots(c),
				comp.Start, comp.Angle)

		case vg.CloseComp:
			c.gc.Close()

		default:
			panic(fmt.Sprintf("Unknown path component: %d", comp.Type))
		}
	}
}

func (c *XImgCanvas) DPI() float64 {
	return float64(c.gc.GetDPI())
}

func (c *XImgCanvas) FillString(font vg.Font, x, y vg.Length, str string) {
	c.gc.Save()
	defer c.gc.Restore()

	data, ok := fontMap[font.Name()]
	if !ok {
		panic(fmt.Sprintf("Font name %s is unknown", font.Name()))
	}
	if !registeredFont[font.Name()] {
		draw2d.RegisterFont(data, font.Font())
		registeredFont[font.Name()] = true
	}
	c.gc.SetFontData(data)
	c.gc.Translate(x.Dots(c), y.Dots(c))
	c.gc.Scale(1, -1)
	c.gc.FillString(str)
}

var (
	// RegisteredFont contains the set of font names
	// that have already been registered with draw2d.
	registeredFont = map[string]bool{}

	// FontMap contains a mapping from vg's font
	// names to draw2d.FontData for the corresponding
	// font.  This is needed to register the  fonts with
	// draw2d.
	fontMap = map[string]draw2d.FontData{
		"Courier": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilyMono,
			Style:  draw2d.FontStyleNormal,
		},
		"Courier-Bold": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilyMono,
			Style:  draw2d.FontStyleBold,
		},
		"Courier-Oblique": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilyMono,
			Style:  draw2d.FontStyleItalic,
		},
		"Courier-BoldOblique": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilyMono,
			Style:  draw2d.FontStyleItalic | draw2d.FontStyleBold,
		},
		"Helvetica": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilySans,
			Style:  draw2d.FontStyleNormal,
		},
		"Helvetica-Bold": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilySans,
			Style:  draw2d.FontStyleBold,
		},
		"Helvetica-Oblique": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilySans,
			Style:  draw2d.FontStyleItalic,
		},
		"Helvetica-BoldOblique": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilySans,
			Style:  draw2d.FontStyleItalic | draw2d.FontStyleBold,
		},
		"Times-Roman": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilySerif,
			Style:  draw2d.FontStyleNormal,
		},
		"Times-Bold": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilySerif,
			Style:  draw2d.FontStyleBold,
		},
		"Times-Italic": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilySerif,
			Style:  draw2d.FontStyleItalic,
		},
		"Times-BoldItalic": draw2d.FontData{
			Name:   "Nimbus",
			Family: draw2d.FontFamilySerif,
			Style:  draw2d.FontStyleItalic | draw2d.FontStyleBold,
		},
	}
)
