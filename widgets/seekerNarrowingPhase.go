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
	"github.com/jkulzer/fib-client/mapWidget"
	// "github.com/jkulzer/fib-server/sharedModels"
)

type SeekerNarrowingPhaseWidget struct {
	widget.BaseWidget
	content *fyne.Container
	fc      *geojson.FeatureCollection
}

func NewSeekerNarrowingPhaseWidget(env env.Env, parentWindow fyne.Window) *SeekerNarrowingPhaseWidget {
	w := &SeekerNarrowingPhaseWidget{}
	w.ExtendBaseWidget(w)

	mapData, err := client.GetMapData(env, parentWindow)
	if err != nil {
		dialog.ShowError(err, parentWindow)
		return w
	}

	w.content = container.NewStack()

	w.fc, err = geojson.UnmarshalFeatureCollection(mapData)
	if err != nil {
		dialog.ShowError(err, parentWindow)
		return w
	}

	mapWidgetInstance := mapWidget.NewMap(w.fc)
	tabs := container.NewAppTabs(
		container.NewTabItem("Map", mapWidgetInstance),
		container.NewTabItem("Questions", NewQuestionWidget(env, parentWindow, mapWidgetInstance)),
		container.NewTabItem("Curses", widget.NewLabel("TODO")),
	)
	tabs.SetTabLocation(container.TabLocationBottom)
	w.content = container.NewStack(tabs)

	return w
}

func (w *SeekerNarrowingPhaseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
