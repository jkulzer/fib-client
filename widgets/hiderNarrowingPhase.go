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

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/location"
	// "github.com/jkulzer/fib-server/sharedModels"
)

type HiderNarrowingPhaseWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewHiderNarrowingPhaseWidget(env env.Env, parentWindow fyne.Window) *HiderNarrowingPhaseWidget {
	w := &HiderNarrowingPhaseWidget{}
	w.ExtendBaseWidget(w)

	w.content = container.NewVBox()
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
	w.content.Add(setLocationButton)

	return w
}

func (w *HiderNarrowingPhaseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
