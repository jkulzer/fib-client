//go:build !android
// +build !android

package location

import (
	"fmt"
	"strconv"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/paulmach/orb"
)

import "C"

func GetLocation(parentWindow fyne.Window) (orb.Point, error) {

	latEntry := widget.NewEntry()
	lonEntry := widget.NewEntry()

	formDone := make(chan bool)

	content := []*widget.FormItem{
		{Text: "Latitude", Widget: latEntry},
		{Text: "Longitude", Widget: lonEntry},
	}
	callback := func(boolean bool) {
		fmt.Println(boolean)
		formDone <- true
	}
	dialog.ShowForm("select location", "confirm", "dismiss", content, callback, parentWindow)

	<-formDone

	var point orb.Point

	lat, err := strconv.ParseFloat(latEntry.Text, 64)
	if err != nil {
		return point, err
	}
	lon, err := strconv.ParseFloat(lonEntry.Text, 64)
	if err != nil {
		return point, err
	}

	point[0] = lon
	point[1] = lat

	return point, nil
}
