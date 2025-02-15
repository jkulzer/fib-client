package widgets

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-server/sharedModels"
)

type HistoryWidget struct {
	widget.BaseWidget
	content *fyne.Container
	history sharedModels.History
}

func NewHistoryWidget(env env.Env, parentWindow fyne.Window) *HistoryWidget {
	w := &HistoryWidget{}
	w.ExtendBaseWidget(w)
	w.content = container.NewVBox()

	history, err := client.GetHistory(env, parentWindow)
	if err != nil {
		log.Err(err).Msg("failed getting history")
		dialog.ShowError(err, parentWindow)
		return w
	}
	w.history = history
	return w
}

func (w *HistoryWidget) CreateRenderer() fyne.WidgetRenderer {
	w.content.RemoveAll()
	for _, item := range w.history {
		itemContainer := container.NewVBox(
			widget.NewLabel(item.Title),
			widget.NewLabel(item.Description),
		)
		w.content.Add(itemContainer)
	}
	return widget.NewSimpleRenderer(w.content)
}
