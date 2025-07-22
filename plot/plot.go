package plot

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"image/color"
	"io"
)

type FunctionStyle func(*plotter.Function)

var (
	DashesStyle FunctionStyle = func(f *plotter.Function) {
		f.Dashes = []vg.Length{vg.Points(2), vg.Points(2)}
	}
	SetWidthStyle = func(w float64) FunctionStyle {
		return func(f *plotter.Function) {
			f.Width = vg.Points(w)
		}
	}
)

type Plot struct {
	*plot.Plot
}

func New() *Plot {
	p := plot.New()
	p.Title.Text = "Plot"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"
	return &Plot{p}
}

func (p *Plot) Title(s string) *Plot {
	p.Plot.Title.Text = s
	return p
}

func (p *Plot) LabelX(s string) *Plot {
	p.Plot.X.Label.Text = s
	return p
}

func (p *Plot) LabelY(s string) *Plot {
	p.Plot.Y.Label.Text = s
	return p
}

func (p *Plot) RangeX(start, end float64) *Plot {
	p.Plot.X.Min = start
	p.Plot.X.Max = end
	return p
}

func (p *Plot) RangeY(start, end float64) *Plot {
	p.Plot.Y.Min = start
	p.Plot.Y.Max = end
	return p
}

func (p *Plot) DrawFunction(f func(float64) float64, color color.RGBA, legend string, styles ...FunctionStyle) *Plot {
	fn := plotter.NewFunction(f)
	fn.Color = color
	for _, style := range styles {
		style(fn)
	}
	p.Add(fn)
	p.Legend.Add(legend)
	return p
}

func (p *Plot) Save(w, h vg.Length, file string) error {
	return p.Plot.Save(w, h, file)
}

func (p *Plot) WriterToPNG(w, h vg.Length) (io.WriterTo, error) {
	return p.Plot.WriterTo(w, h, "png")
}

func (p *Plot) WriterToTex(w, h vg.Length) (io.WriterTo, error) {
	return p.Plot.WriterTo(w, h, "tex")
}

func (p *Plot) WriterToJpg(w, h vg.Length) (io.WriterTo, error) {
	return p.Plot.WriterTo(w, h, "jpg")
}
