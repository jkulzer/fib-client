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
	content      *widget.Accordion
	history      sharedModels.History
	env          env.Env
	parentWindow fyne.Window
}

func NewHistoryWidget(env env.Env, parentWindow fyne.Window) *HistoryWidget {
	w := &HistoryWidget{}
	w.ExtendBaseWidget(w)
	w.content = widget.NewAccordion()

	w.env = env
	w.parentWindow = parentWindow
	log.Debug().Msg("getting history")
	history, err := client.GetHistory(w.env, w.parentWindow)
	if err != nil {
		log.Err(err).Msg("failed getting history")
		dialog.ShowError(err, w.parentWindow)
	}
	w.history = history
	return w
}

func (w *HistoryWidget) Refresh() {
	log.Debug().Msg("getting history")
	history, err := client.GetHistory(w.env, w.parentWindow)
	if err != nil {
		log.Err(err).Msg("failed getting history")
		dialog.ShowError(err, w.parentWindow)
	}
	w.history = history

	w.content.Items = nil
	for _, item := range w.history {
		itemContainer := widget.NewAccordionItem(
			item.Title, container.NewVBox(widget.NewLabel(item.Description)),
		)
		w.content.Append(itemContainer)
	}
	w.BaseWidget.Refresh()
}

func (w *HistoryWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewScroll(w.content))
}
