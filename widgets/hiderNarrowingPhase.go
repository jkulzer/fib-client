package widgets

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	// "fmt"
	// "time"
	//
	// "github.com/rs/zerolog/log"

	"github.com/paulmach/orb/geojson"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/location"
	"github.com/jkulzer/fib-client/mapWidget"
	// "github.com/jkulzer/fib-server/sharedModels"
)

type HiderNarrowingPhaseWidget struct {
	widget.BaseWidget
	content *fyne.Container
	fc      *geojson.FeatureCollection
}

func NewHiderNarrowingPhaseWidget(env env.Env, parentWindow fyne.Window) *HiderNarrowingPhaseWidget {
	w := &HiderNarrowingPhaseWidget{}
	w.ExtendBaseWidget(w)

	w.content = container.NewStack()

	mapData, err := client.GetMapData(env, parentWindow)
	if err != nil {
		dialog.ShowError(err, parentWindow)
		return w
	}

	w.fc, err = geojson.UnmarshalFeatureCollection(mapData)
	if err != nil {
		dialog.ShowError(err, parentWindow)
		return w
	}

	mapWidgetInstance := mapWidget.NewMap(w.fc, env, &parentWindow)
	historyWidgetInstance := NewHistoryWidget(env, parentWindow)
	cardsWidgetInstance := NewCardsWidget(env, parentWindow)

	setLocationButton := widget.NewButton("Set Location", func() {
		go func() {
			locationPoint, err := location.GetLocation(parentWindow)
			if err != nil {
				dialog.ShowError(err, parentWindow)
				return
			}
			err = client.SaveLocation(env, parentWindow, locationPoint)
			if err != nil {
				dialog.ShowError(err, parentWindow)
			}
		}()
	})

	tabs := container.NewAppTabs(
		container.NewTabItem("Map", mapWidgetInstance),
		container.NewTabItem("Cards", cardsWidgetInstance),
		container.NewTabItem("History", historyWidgetInstance),
		container.NewTabItem("Location", container.NewVBox(setLocationButton)),
	)
	tabs.SetTabLocation(container.TabLocationBottom)
	w.content.Add(tabs)

	return w
}

func (w *HiderNarrowingPhaseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
