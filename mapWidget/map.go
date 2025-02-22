package mapWidget

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/nfnt/resize"

	"github.com/rs/zerolog/log"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/project"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"golang.org/x/image/draw"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

const tileSize = 256

// Map widget renders an interactive map using OpenStreetMap tile data.
type Map struct {
	widget.BaseWidget

	pixels       *image.NRGBA
	w, h         int
	zoom, x, y   int
	dragX, dragY float32

	cl *http.Client

	tileSource       string // url to download xyz tiles (example: "https://tile.openstreetmap.org/%d/%d/%d.png")
	hideAttribution  bool   // enable copyright attribution
	attributionLabel string // label for attribution (example: "OpenStreetMap")
	attributionURL   string // url for attribution (example: "https://openstreetmap.org")
	hideZoomButtons  bool   // enable zoom buttons
	hideMoveButtons  bool   // enable move map buttons

	lineColor color.Color

	featureCollection *geojson.FeatureCollection // overlay to render
}

type linePos struct {
	startX float32
	startY float32
	endX   float32
	endY   float32
}

// MapOption configures the provided map with different features.
type MapOption func(*Map)

// WithOsmTiles configures the map to use osm tile source.
func WithOsmTiles() MapOption {
	return func(m *Map) {
		m.tileSource = "https://tile.openstreetmap.org/%d/%d/%d.png"
		// m.tileSource = "https://b.tile.openstreetmap.fr/hot/%d/%d/%d.png"
		m.attributionLabel = "OpenStreetMap"
		m.attributionURL = "https://openstreetmap.org"
		m.hideAttribution = false
	}
}

// WithTileSource configures the map to use a custom tile source.
func WithTileSource(tileSource string) MapOption {
	return func(m *Map) {
		m.tileSource = tileSource
	}
}

// WithAttribution configures the map widget to display an attribution.
func WithAttribution(enable bool, label, url string) MapOption {
	return func(m *Map) {
		m.hideAttribution = !enable
		m.attributionLabel = label
		m.attributionURL = url
	}
}

// WithZoomButtons enables or disables zoom controls.
func WithZoomButtons(enable bool) MapOption {
	return func(m *Map) {
		m.hideZoomButtons = !enable
	}
}

// WithScrollButtons enables or disables map scroll controls.
func WithScrollButtons(enable bool) MapOption {
	return func(m *Map) {
		m.hideMoveButtons = !enable
	}
}

// WithHTTPClient configures the map to use a custom http client.
func WithHTTPClient(client *http.Client) MapOption {
	return func(m *Map) {
		m.cl = client
	}
}

// NewMap creates a new instance of the map widget.
func NewMap(fc *geojson.FeatureCollection) *Map {
	m := &Map{cl: &http.Client{}}
	WithOsmTiles()(m)

	// m.lineColor = color.RGBA{
	// 	R: 245,
	// 	G: 239,
	// 	B: 31,
	// 	A: 255,
	// }
	m.lineColor = color.RGBA{
		R: 255,
		G: 0,
		B: 0,
		A: 255,
	}
	m.zoom = 10
	m.x = 38
	m.y = -176
	m.featureCollection = fc
	m.ExtendBaseWidget(m)
	return m
}

func (m *Map) SetFeatureCollection(fc *geojson.FeatureCollection) {
	m.featureCollection = fc
}

// NewMapWithOptions creates a new instance of the map widget with provided map options.
func NewMapWithOptions(fc *geojson.FeatureCollection, opts ...MapOption) *Map {
	m := NewMap(fc)
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// MinSize returns the smallest possible size for a widget.
// For our map this is a constant size representing a single tile on a device with
// the highest known DPI (4x).
func (m *Map) MinSize() fyne.Size {
	return fyne.NewSize(64, 64)
}

// PanEast will move the map to the East by 1 tile.
func (m *Map) PanEast() {
	m.x++
	m.Refresh()
}

// PanNorth will move the map to the North by 1 tile.
func (m *Map) PanNorth() {
	m.y--
	m.Refresh()
}

// PanSouth will move the map to the South by 1 tile.
func (m *Map) PanSouth() {
	m.y++
	m.Refresh()
}

// PanWest will move the map to the west by 1 tile.
func (m *Map) PanWest() {
	m.x--
	m.Refresh()
}

// Zoom sets the zoom level to a specific value, between 0 and 19.
func (m *Map) Zoom(zoom int) {
	if zoom < 10 || zoom > 19 {
		return
	}
	delta := zoom - m.zoom
	if delta > 0 {
		for i := 0; i < delta; i++ {
			m.zoomInStep()
		}
	} else if delta < 0 {
		for i := 0; i > delta; i-- {
			m.zoomOutStep()
		}
	}
	m.Refresh()
}

// ZoomIn steps the scale of this map to be one step zoomed in.
func (m *Map) ZoomIn() {
	if m.zoom >= 19 {
		return
	}
	m.zoomInStep()
	m.Refresh()
}

// ZoomOut steps the scale of this map to be one step zoomed out.
func (m *Map) ZoomOut() {
	if m.zoom <= 0 {
		return
	}
	m.zoomOutStep()
	m.Refresh()
}

// CreateRenderer returns the renderer for this widget.
// A map renderer is simply the map Raster with user interface elements overlaid.
func (m *Map) CreateRenderer() fyne.WidgetRenderer {
	var zoom fyne.CanvasObject
	if !m.hideZoomButtons {
		zoom = container.NewVBox(
			newMapButton(theme.ZoomInIcon(), m.ZoomIn),
			newMapButton(theme.ZoomOutIcon(), m.ZoomOut))
	}

	var move fyne.CanvasObject
	if !m.hideMoveButtons {
		buttonLayout := container.NewGridWithColumns(3, layout.NewSpacer(),
			newMapButton(theme.MoveUpIcon(), m.PanNorth), layout.NewSpacer(),
			newMapButton(theme.NavigateBackIcon(), m.PanWest), layout.NewSpacer(),
			newMapButton(theme.NavigateNextIcon(), m.PanEast), layout.NewSpacer(),
			newMapButton(theme.MoveDownIcon(), m.PanSouth), layout.NewSpacer())
		move = container.NewVBox(buttonLayout)
	}

	var copyright fyne.CanvasObject
	if !m.hideAttribution {
		license, _ := url.Parse(m.attributionURL)
		link := widget.NewHyperlink(m.attributionLabel, license)
		copyright = container.NewHBox(layout.NewSpacer(), link)
	}

	overlay := container.NewBorder(nil, copyright, move, zoom)

	// customOverlay := m.geoJSONOverlay()

	c := container.NewStack(canvas.NewRaster(m.draw), canvas.NewRaster(m.overlay), container.NewPadded(overlay))
	// c := customOverlay
	return widget.NewSimpleRenderer(c)
}

func (m *Map) draw(w, h int) image.Image {
	log.Debug().Msg("drawing map")
	scale := 1
	tileSize := tileSize
	// TODO use retina tiles once OSM supports it in their server (text scaling issues)...
	if c := fyne.CurrentApp().Driver().CanvasForObject(m); c != nil {
		scale = int(c.Scale())
		if scale < 1 {
			scale = 1
		}
		tileSize = tileSize * scale
	}

	if m.w != w || m.h != h {
		m.pixels = image.NewNRGBA(image.Rect(0, 0, w, h))
	}

	midTileX := (w - tileSize*2) / 2
	midTileY := (h - tileSize*2) / 2
	if m.zoom == 0 {
		midTileX += tileSize / 2
		midTileY += tileSize / 2
	}

	count := 1 << m.zoom
	mx := m.x + int(float32(count)/2-0.5)
	my := m.y + int(float32(count)/2-0.5)
	firstTileX := mx - int(math.Ceil(float64(midTileX)/float64(tileSize)))
	firstTileY := my - int(math.Ceil(float64(midTileY)/float64(tileSize)))

	for x := firstTileX; (x-firstTileX)*tileSize <= w+tileSize; x++ {
		for y := firstTileY; (y-firstTileY)*tileSize <= h+tileSize; y++ {
			if x < 0 || y < 0 || x >= int(count) || y >= int(count) {
				continue
			}

			src, err := getTile(m.tileSource, x, y, m.zoom, m.cl)
			if err != nil {
				fyne.LogError("tile fetch error", err)
				continue
			}

			pos := image.Pt(midTileX+(x-mx)*tileSize,
				midTileY+(y-my)*tileSize)
			scaled := src
			if scale > 1 {
				scaled = resize.Resize(uint(tileSize), uint(tileSize), src, resize.Lanczos2)
			}
			draw.Copy(m.pixels, pos, scaled, image.Rect(0, 0, tileSize, tileSize), draw.Over, nil)
		}
	}

	startTime := time.Now()
	log.Debug().Msg(fmt.Sprint("overlay drawing took ", time.Since(startTime)))

	return m.pixels
}

func (m *Map) overlay(w, h int) image.Image {

	scale := 1
	tileSize := tileSize
	// TODO use retina tiles once OSM supports it in their server (text scaling issues)...
	if c := fyne.CurrentApp().Driver().CanvasForObject(m); c != nil {
		scale = int(c.Scale())
		if scale < 1 {
			scale = 1
		}
		tileSize = tileSize * scale
	}

	size := m.Size()

	// Convert tile bounds to geographic coordinates
	x, y := XYToTile(m.x, m.y, m.zoom)
	middleLon, middleLat := TileToCoords(int(x), int(y), m.zoom)

	var middlePoint orb.Point
	middlePoint[0] = middleLon
	middlePoint[1] = middleLat

	middlePointProj := project.Point(middlePoint, project.WGS84.ToMercator)

	tileCoordWidth, tileCoordHeight := MercatorSize(int(x), int(y), m.zoom)

	projCoordPerPixelWidth := tileCoordWidth / 256
	projCoordPerPixelHeight := tileCoordHeight / 256

	offsetToLeft := float32(tileCoordWidth/256) * (size.Width / 2)
	offsetToTop := float32(tileCoordHeight/256) * (size.Height / 2)

	var offsetVectorProj orb.Point
	offsetVectorProj[0] = float64(offsetToLeft)
	offsetVectorProj[1] = float64(offsetToTop)

	var leftTopProj orb.Point
	leftTopProj[0] = middlePointProj[0] - float64(offsetToLeft)
	leftTopProj[1] = middlePointProj[1] + float64(offsetToTop)

	var rightBottomProj orb.Point
	rightBottomProj[0] = middlePointProj[0] + float64(offsetToLeft)
	rightBottomProj[1] = middlePointProj[1] - float64(offsetToTop)

	img := image.NewRGBA(image.Rect(0, 0, m.pixels.Bounds().Max.X, m.pixels.Bounds().Max.Y))
	gc := draw2dimg.NewGraphicContext(img)

	gc.Clear()
	gc.DrawImage(m.pixels)

	for _, feature := range m.featureCollection.Features {
		switch feature.Geometry.GeoJSONType() {
		case "LineString":
			lineString := feature.Geometry.(orb.LineString)
			m.drawLineString(lineString, middlePointProj, size, projCoordPerPixelWidth, projCoordPerPixelHeight, gc)
		case "Polygon":
			renderPolygon(feature.Geometry, gc, middlePointProj, size, projCoordPerPixelWidth, projCoordPerPixelHeight)
		case "MultiPolygon":
			multiPolygon, _ := feature.Geometry.(orb.MultiPolygon)
			for _, polygon := range multiPolygon {
				renderPolygon(polygon, gc, middlePointProj, size, projCoordPerPixelWidth, projCoordPerPixelHeight)
			}
		}
	}

	// gc.Clear()

	transparentImg := adjustTransparency(img, 0.5)
	// // gc.SetFillColor(color.Transparent)
	// // gc.SetFillColor(color.RGBA{255, 255, 255, 255})
	// // gc.MoveTo(10, 10)
	// // gc.LineTo(100, 100)
	// // gc.MoveTo(100, 100)
	// // gc.LineTo(200, 100)
	// // gc.MoveTo(200, 100)
	// // gc.LineTo(10, 10)
	// // gc.Close()
	// // gc.FillStroke()
	// geometry.FillStyle(gc, 0, 0, 300, 300)

	b := img.Bounds()
	dest := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dest, dest.Bounds(), transparentImg, b.Min, draw.Over)
	m.pixels = dest
	return m.pixels
}

func (m *Map) zoomInStep() {
	m.zoom++
	m.x *= 2
	m.y *= 2
}

func (m *Map) zoomOutStep() {
	m.zoom--
	m.x /= 2
	m.y /= 2
}

func (m *Map) SetPosition(x, y int) {
	m.x = x
	m.y = y
}
func (m *Map) GetPosition() (int, int) {
	return m.x, m.y
}

func (m *Map) GetZoom() int {
	return m.zoom
}

func (m *Map) drawLineString(lineString orb.LineString, middlePointProj orb.Point, size fyne.Size, projCoordPerPixelWidth, projCoordPerPixelHeight float64, gc *draw2dimg.GraphicContext) {
	linePositions := getLinePositions(lineString, middlePointProj, size, projCoordPerPixelWidth, projCoordPerPixelHeight)
	for _, position := range linePositions {
		m.drawLine(position, gc)
	}
}

func getLinePositions(lineString orb.LineString, middlePointProj orb.Point, size fyne.Size, projCoordPerPixelWidth, projCoordPerPixelHeight float64) []linePos {
	var linePositions []linePos
	lsLastIndex := len(lineString) - 1
	for lsIndex, point := range lineString {
		if lsIndex != lsLastIndex {

			endPoint := lineString[lsIndex+1]

			projPoint := project.Point(point, project.WGS84.ToMercator)
			projEndPoint := project.Point(endPoint, project.WGS84.ToMercator)

			startLonDiff := projPoint[0] - middlePointProj[0]
			startLatDiff := middlePointProj[1] - projPoint[1]

			endLonDiff := projEndPoint[0] - middlePointProj[0]
			endLatDiff := middlePointProj[1] - projEndPoint[1]

			scaleAddX := float32(size.Width / 1.54)
			scaleAddY := float32(size.Height / 1.54)

			startX := float32(startLonDiff/projCoordPerPixelWidth) + scaleAddX
			startY := float32(startLatDiff/projCoordPerPixelHeight) + scaleAddY
			endX := float32(endLonDiff/projCoordPerPixelWidth) + scaleAddX
			endY := float32(endLatDiff/projCoordPerPixelHeight) + scaleAddY
			linePositions = append(linePositions, linePos{startX, startY, endX, endY})
		}
	}
	return linePositions
}

func (m *Map) drawLine(linePosition linePos, gc *draw2dimg.GraphicContext) {
	gc.SetFillColor(m.lineColor)
	gc.SetStrokeColor(m.lineColor)
	gc.SetLineWidth(1)

	gc.MoveTo(float64(linePosition.startX), float64(linePosition.startY))
	gc.LineTo(float64(linePosition.endX), float64(linePosition.endY))
	gc.Close()
	gc.FillStroke()
}

func adjustTransparency(src image.Image, alphaFactor float64) *image.NRGBA {
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			origColor := src.At(x, y)
			r, g, b, a := origColor.RGBA() // Returns 16-bit values

			// Convert to 8-bit alpha and apply transparency factor
			newAlpha := uint8(float64(uint8(a>>8)) * alphaFactor)

			// Create new color with modified alpha
			newColor := color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: newAlpha,
			}
			dst.Set(x, y, newColor)
		}
	}
	return dst
}

func renderPolygon(featureGeometry orb.Geometry, gc *draw2dimg.GraphicContext, middlePointProj orb.Point, size fyne.Size, projCoordPerPixelWidth float64, projCoordPerPixelHeight float64) {
	rings := []orb.Ring(featureGeometry.(orb.Polygon))
	ringListLen := len(rings)
	for ringIndex, ring := range rings {
		if ringIndex < ringListLen-1 {
			nextRing := rings[ringIndex+1]
			if ring[len(ring)-1] == nextRing[len(nextRing)-1] {
				log.Debug().Msg("reversed")
				rings[ringIndex+1].Reverse()
			} else if ring[0] == nextRing[0] {
				log.Debug().Msg("reversed")
				ring.Reverse()
			} else if ring[0] == nextRing[len(nextRing)-1] {
				log.Debug().Msg("reversed")
				ring.Reverse()
				rings[ringIndex+1].Reverse()
			}
		}
		linePositions := getLinePositions(orb.LineString(ring), middlePointProj, size, projCoordPerPixelWidth, projCoordPerPixelHeight)

		gc.SetFillRule(draw2d.FillRuleEvenOdd)
		// gc.SetFillRule(draw2d.FillRuleWinding)

		gc.SetFillColor(color.RGBA{83, 118, 245, 255})
		gc.SetStrokeColor(color.Transparent)
		gc.SetLineWidth(0)
		for _, linePosition := range linePositions {
			gc.MoveTo(float64(linePosition.startX), float64(linePosition.startY))
			gc.LineTo(float64(linePosition.endX), float64(linePosition.endY))
		}
		gc.MoveTo(float64(linePositions[0].startX), float64(linePositions[0].startY))
		gc.LineTo(float64(linePositions[0].endX), float64(linePositions[0].endY))
		gc.Close()
	}
	gc.FillStroke()
}
