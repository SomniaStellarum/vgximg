package main

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	"github.com/Arrow/vgximg"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(int64(0))

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Plotutil example"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	err = plotutil.AddLinePoints(p,
		"First", randomPoints(15))
	if err != nil {
		panic(err)
	}
	cnvs, err := vgximg.New(4*96, 4*96, "Example")
	if err != nil {
		panic(err)
	}
	p.Draw(plot.MakeDrawArea(cnvs))
	cnvs.Paint()
	time.Sleep(5 * time.Second)

	err = plotutil.AddLinePoints(p,
		"Second", randomPoints(15),
		"Third", randomPoints(15))
	if err != nil {
		panic(err)
	}
	p.Draw(plot.MakeDrawArea(cnvs))
	cnvs.Paint()
	time.Sleep(10 * time.Second)

	// Save the plot to a PNG file.
	//        if err := p.Save(4, 4, "points.png"); err != nil {
	//                panic(err)
	//        }
}

// randomPoints returns some random x, y points.
func randomPoints(n int) plotter.XYs {
	pts := make(plotter.XYs, n)
	for i := range pts {
		if i == 0 {
			pts[i].X = rand.Float64()
		} else {
			pts[i].X = pts[i-1].X + rand.Float64()
		}
		pts[i].Y = pts[i].X + 10*rand.Float64()
	}
	return pts
}
