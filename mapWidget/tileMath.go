package mapWidget

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/project"

	"math"
)

// source: https://www.netzwolf.info/geo/math/tilebrowser.html?tx=550&ty=335&tz=10#tile

func XYToTile(x, y, zoom int) (float64, float64) {
	tileX := float64(x) + ((math.Pow(float64(2), float64(zoom))) / 2)
	tileY := float64(y) + ((math.Pow(float64(2), float64(zoom))) / 2)
	return tileX, tileY
}

func TileToCoords(tileX, tileY, zoom int) (lon float64, lat float64) {
	x := (float64(tileX)) / (math.Pow(2.0, float64(zoom)))
	y := (float64(tileY)) / (math.Pow(2.0, float64(zoom)))
	mercatorL := +(x*2 - 1) * math.Pi
	mercatorW := -(y*2 - 1) * math.Pi
	length := mercatorL
	width := 2*math.Atan(math.Exp(mercatorW)) - (math.Pi / 2)
	lon = length / math.Pi * 180
	lat = width / math.Pi * 180
	return lon, lat
}

func MercatorSize(tileX, tileY, zoom int) (float64, float64) {
	var leftTopPoint orb.Point
	var leftBottomPoint orb.Point
	var rightTopPoint orb.Point

	leftTopPoint[0], leftTopPoint[1] = TileToCoords(tileX, tileY, zoom)
	leftBottomPoint[0], leftBottomPoint[1] = TileToCoords(tileX+1, tileY, zoom)
	rightTopPoint[0], rightTopPoint[1] = TileToCoords(tileX, tileY+1, zoom)

	leftTopProj := project.Geometry(leftTopPoint, project.WGS84.ToMercator)
	leftBottomProj := project.Geometry(leftBottomPoint, project.WGS84.ToMercator)
	rightTopProj := project.Geometry(rightTopPoint, project.WGS84.ToMercator)

	leftTopProjPoint := leftTopProj.(orb.Point)
	leftBottomProjPoint := leftBottomProj.(orb.Point)
	rightTopProjPoint := rightTopProj.(orb.Point)

	lonDiff := leftTopProjPoint[1] - rightTopProjPoint[1]
	latDiff := leftBottomProjPoint[0] - leftTopProjPoint[0]

	return lonDiff, latDiff
}
