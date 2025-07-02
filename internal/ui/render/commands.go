package render

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/ducnd58233/gobrowser/internal/browser"
	"github.com/ducnd58233/gobrowser/internal/ui/types"
)

type DrawCommand interface {
	Execute(gtx layout.Context, theme *material.Theme, scrollY float64)
	GetBounds() types.Bounds
	GetNode() browser.Node
}

type drawText struct {
	x, y           float64
	text           string
	fontSize       float64
	color          color.NRGBA
	node           browser.Node
	cachedTextSize unit.Sp
}

func NewDrawText(
	x, y float64,
	text string,
	fontSize float64,
	color color.NRGBA,
	node browser.Node,
) DrawCommand {
	return &drawText{
		x:              x,
		y:              y,
		text:           text,
		fontSize:       fontSize,
		color:          color,
		node:           node,
		cachedTextSize: unit.Sp(float32(fontSize)),
	}
}

func (dt *drawText) Execute(
	gtx layout.Context,
	theme *material.Theme,
	scrollY float64,
) {
	adjustedY := dt.y - scrollY

	if !dt.isVisible(adjustedY, float64(gtx.Constraints.Max.Y)) {
		return
	}

	label := material.Body1(theme, dt.text)
	label.Color = dt.color
	label.TextSize = dt.cachedTextSize

	stack := op.Offset(image.Pt(int(dt.x), int(adjustedY))).Push(gtx.Ops)
	defer stack.Pop()
	label.Layout(gtx)
}

func (dt *drawText) GetBounds() types.Bounds {
	return types.Bounds{
		X:      dt.x,
		Y:      dt.y,
		Width:  float64(len(dt.text)) * dt.fontSize * browser.CharWidthRatio,
		Height: dt.fontSize,
	}
}

func (dt *drawText) GetNode() browser.Node {
	return dt.node
}

func (dt *drawText) isVisible(adjustedY, maxY float64) bool {
	textHeight := dt.fontSize
	if textHeight <= 0 {
		textHeight = browser.DefaultFontSize
	}

	bounds := types.Bounds{
		X:      dt.x,
		Y:      adjustedY,
		Width:  float64(len(dt.text)) * dt.fontSize * browser.CharWidthRatio,
		Height: textHeight,
	}

	viewportBounds := types.Bounds{
		X:      0,
		Y:      -textHeight * 2,   // Allow some overflow for smooth scrolling
		Width:  float64(maxY) * 2, // Use a reasonable width
		Height: maxY + textHeight*2,
	}

	return bounds.Intersects(&viewportBounds)
}

type drawRect struct {
	bounds types.Bounds
	color  color.NRGBA
	node   browser.Node
}

func NewDrawRect(
	bounds types.Bounds,
	color color.NRGBA,
	node browser.Node,
) DrawCommand {
	return &drawRect{
		bounds: bounds,
		color:  color,
		node:   node,
	}
}

func (dr *drawRect) Execute(
	gtx layout.Context, theme *material.Theme, scrollY float64,
) {
	adjustedY := dr.bounds.Y - scrollY

	if !dr.isVisible(adjustedY, float64(gtx.Constraints.Max.Y)) {
		return
	}

	rect := image.Rectangle{
		Min: image.Pt(int(dr.bounds.X), int(adjustedY)),
		Max: image.Pt(int(dr.bounds.X+dr.bounds.Width), int(adjustedY+dr.bounds.Height)),
	}

	defer clip.Rect(rect).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: dr.color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func (dr *drawRect) GetBounds() types.Bounds {
	return dr.bounds
}

func (dr *drawRect) GetNode() browser.Node {
	return dr.node
}

func (dr *drawRect) isVisible(adjustedY, maxY float64) bool {
	adjustedBounds := types.Bounds{
		X:      dr.bounds.X,
		Y:      adjustedY,
		Width:  dr.bounds.Width,
		Height: dr.bounds.Height,
	}

	viewportBounds := types.Bounds{
		X:      0,
		Y:      0,
		Width:  float64(maxY) * 2,
		Height: maxY,
	}

	return adjustedBounds.Intersects(&viewportBounds)
}

type drawLine struct {
	x1, y1, x2, y2 float64
	color          color.NRGBA
	width          float64
	node           browser.Node
}

func NewDrawLine(
	x1, y1, x2, y2 float64,
	color color.NRGBA,
	width float64,
	node browser.Node,
) DrawCommand {
	return &drawLine{
		x1: x1, y1: y1, x2: x2, y2: y2,
		color: color,
		width: width,
		node:  node,
	}
}

func (dl *drawLine) Execute(
	gtx layout.Context,
	theme *material.Theme,
	scrollY float64,
) {
	adjY1 := dl.y1 - scrollY
	adjY2 := dl.y2 - scrollY

	if !dl.isVisible(adjY1, adjY2, float64(gtx.Constraints.Max.Y)) {
		return
	}

	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Point{X: float32(dl.x1), Y: float32(adjY1)})
	path.LineTo(f32.Point{X: float32(dl.x2), Y: float32(adjY2)})
	path.Close()

	defer clip.Stroke{
		Path:  path.End(),
		Width: float32(dl.width),
	}.Op().Push(gtx.Ops).Pop()
	paint.ColorOp{Color: dl.color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func (dl *drawLine) GetBounds() types.Bounds {
	minX := math.Min(dl.x1, dl.x2)
	maxX := math.Max(dl.x1, dl.x2)
	minY := math.Min(dl.y1, dl.y2)
	maxY := math.Max(dl.y1, dl.y2)
	return types.Bounds{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

func (dl *drawLine) GetNode() browser.Node {
	return dl.node
}

func (dl *drawLine) isVisible(y1, y2, maxY float64) bool {
	lineBounds := types.Bounds{
		X:      math.Min(dl.x1, dl.x2),
		Y:      math.Min(y1, y2),
		Width:  math.Abs(dl.x2 - dl.x1),
		Height: math.Abs(y2 - y1),
	}

	viewportBounds := types.Bounds{
		X:      0,
		Y:      0,
		Width:  float64(maxY) * 2,
		Height: maxY,
	}

	return lineBounds.Intersects(&viewportBounds)
}
