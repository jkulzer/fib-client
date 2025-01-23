package widgets

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-server/sharedModels"
)

type SeekerRunPhaseWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewSeekerRunPhaseWidget(env env.Env, parentWindow fyne.Window) *SeekerRunPhaseWidget {
	w := &SeekerRunPhaseWidget{}
	w.ExtendBaseWidget(w)
	w.content = container.NewVBox(widget.NewLabel("Time until hiding phase ends:"))

	runStartTime, err := client.RunStartTime(env, parentWindow)
	if err != nil {
		log.Err(err).Msg(fmt.Sprint(err))
		dialog.ShowError(err, parentWindow)
	}

	str := binding.NewString()
	str.Set("Countdown initializing")

	text := widget.NewLabelWithData(str)

	w.content.Add(text)

	go func() {

		for {
			timer := time.NewTimer(16 * time.Millisecond)
			<-timer.C

			countdown := time.Until(runStartTime.Add(sharedModels.RunDuration))

			countdownString := countdown.Truncate(10 * time.Millisecond).String()

			str.Set(countdownString)
			w.Refresh()
		}
	}()

	return w
}

func (w *SeekerRunPhaseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
