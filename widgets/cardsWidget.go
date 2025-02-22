package widgets

import (
	"errors"
	"fmt"
	"slices"
	"time"

	fyne "fyne.io/fyne/v2"
	// "fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-server/sharedModels"
)

type CardsWidget struct {
	widget.BaseWidget
	content      *fyne.Container
	env          env.Env
	parentWindow fyne.Window
}

func NewCardsWidget(env env.Env, parentWindow fyne.Window) *CardsWidget {
	w := &CardsWidget{}
	w.ExtendBaseWidget(w)
	w.env = env
	w.parentWindow = parentWindow
	w.content = container.NewVBox()

	go func() {
		for {
			timer := time.NewTimer(2 * time.Second)
			<-timer.C

			w.Refresh()
		}
	}()
	return w
}

func (w *CardsWidget) CreateRenderer() fyne.WidgetRenderer {
	cardActions, err := client.GetCardActions(w.env, w.parentWindow)
	if err != nil {
		dialog.ShowError(err, w.parentWindow)
	}
	w.content.Add(
		widget.NewButton("Check for in progress draw", func() {
			draw, err := client.GetDraw(w.env, w.parentWindow)
			if err != nil {
				log.Err(err).Msg("failed getting draw")
				dialog.ShowError(err, w.parentWindow)
				return
			}
			if len(draw.Cards) < 1 {
				dialog.ShowInformation("Card draw", "There is no draw in progress", w.parentWindow)
			} else {
				cardSelectDialog(w.env, w.parentWindow)
			}
		}),
	)
	if len(cardActions.Draws) == 0 {
		w.content.Add(widget.NewLabel("no remaining draws"))
	} else {
		drawContainer := container.NewVBox()
		for _, cardAction := range cardActions.Draws {
			drawContainer.Add(
				container.NewVBox(
					widget.NewLabel("Draw "+fmt.Sprint(cardAction.CardsToDraw)+" cards and pick "+fmt.Sprint(cardAction.CardsToPick)),
					widget.NewButton("Draw!", func() {
						log.Debug().Msg("use card draw with ID " + fmt.Sprint(cardAction.DrawID))
						err := client.DrawCards(w.env, w.parentWindow, cardAction.DrawID)
						if errors.Is(err, client.ErrAlreadyDrewCards) {
							dialog.ShowInformation("Card drawing", "You need to pick cards from your previous draw before you can draw new cards", w.parentWindow)
							return
						}
						if err != nil {
							log.Err(err).Msg("failed drawing cards")
							dialog.ShowError(err, w.parentWindow)
							return
						}
						cardSelectDialog(w.env, w.parentWindow)
					}),
				),
			)
		}
		w.content.Add(drawContainer)
	}
	hiderHand, err := client.GetHiderHand(w.env, w.parentWindow)
	if err != nil {
		log.Err(err).Msg("failed getting hider hand")
		dialog.ShowError(err, w.parentWindow)
		return widget.NewSimpleRenderer(container.NewScroll(w.content))
	}
	cardGrid := container.NewGridWithRows(2)
	for _, handCard := range hiderHand.List {
		cardGrid.Add(NewCardWidget(handCard, DisplayCardWidget, nil, 0))
	}
	w.content.Add(widget.NewLabel("Your hand:"))
	w.content.Add(cardGrid)
	return widget.NewSimpleRenderer(container.NewScroll(w.content))
}

func cardSelectDialog(env env.Env, parentWindow fyne.Window) {
	drawnCardsContainer := NewCardSelectWidget(env, parentWindow)
	cardDrawDialog := dialog.NewCustom("Drawn cards", "dismiss", drawnCardsContainer, parentWindow)
	cardDrawDialog.Resize(fyne.NewSize(300, 600))
	cardDrawDialog.Show()
}

type CardWidget struct {
	widget.BaseWidget
	content          *fyne.Container
	cardIndex        uint
	selected         bool
	selectedText     *widget.Label
	widgetType       CardWidgetType
	cardSelectWidget *CardSelectWidget
	card             sharedModels.Card
}

type CardWidgetType int

const (
	DrawCardWidget CardWidgetType = iota
	PlayCardWidget
	DisplayCardWidget
)

func NewCardWidget(card sharedModels.Card, widgetType CardWidgetType, cardSelectWidget *CardSelectWidget, cardIndex uint) *CardWidget {
	w := &CardWidget{
		cardIndex:        cardIndex,
		content:          container.NewVBox(),
		selectedText:     widget.NewLabel("selected"),
		card:             card,
		cardSelectWidget: cardSelectWidget,
		widgetType:       widgetType,
	}
	w.ExtendBaseWidget(w)

	title := widget.NewLabel(card.Title)
	title.TextStyle = fyne.TextStyle{
		Bold:      true,
		Underline: true,
	}
	// title.Wrapping = fyne.TextWrapWord
	var description *widget.Label
	if w.card.Description != "" {
		description = widget.NewLabel(w.card.Description)
	} else {
		description = widget.NewLabel("No Description")
	}
	w.content.Add(title)
	w.content.Add(description)

	description.Wrapping = fyne.TextWrapWord

	return w
}

func (w *CardWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}

func (w *CardWidget) Refresh() {
	w.BaseWidget.Refresh()
}

func (w *CardWidget) Tapped(*fyne.PointEvent) {
	switch w.widgetType {
	case DrawCardWidget:
		if w.selected {
			w.selected = false
			w.content.Remove(w.selectedText)
			w.cardSelectWidget.RemoveCardFromPool(w.card.IDInDB)
		} else {
			w.selected = true
			w.content.Add(w.selectedText)
			w.cardSelectWidget.AddCardToPool(w.card.IDInDB)
		}
		w.cardSelectWidget.Refresh()
		w.cardSelectWidget.BaseWidget.Refresh()
		w.Refresh()
		w.BaseWidget.Refresh()
	}
}

type CardSelectWidget struct {
	widget.BaseWidget
	content       *fyne.Container
	env           env.Env
	draw          sharedModels.CurrentDraw
	parentWindow  fyne.Window
	selectedCards []uint
	pickButton    *widget.Button
}

func NewCardSelectWidget(env env.Env, parentWindow fyne.Window) *CardSelectWidget {
	w := &CardSelectWidget{
		env:          env,
		parentWindow: parentWindow,
	}
	w.ExtendBaseWidget(w)
	draw, err := client.GetDraw(w.env, w.parentWindow)
	if err != nil {
		log.Err(err).Msg("failed getting draw")
		dialog.ShowError(err, w.parentWindow)
		return w
	}
	w.draw = draw
	w.pickButton = widget.NewButton("Pick cards", func() {
		fmt.Println("picking cards:")
		fmt.Println(w.selectedCards)
		err := client.PickCards(env, parentWindow, w.selectedCards)
		if err != nil {
			log.Err(err).Msg("failed picking cards")
			dialog.ShowError(err, parentWindow)
		}
	})
	cardGrid := container.NewGridWithRows(2)
	for cardIndex, card := range w.draw.Cards {
		cardGrid.Add(NewCardWidget(card, DrawCardWidget, w, uint(cardIndex)))
	}
	w.content = container.NewBorder(
		nil,
		w.pickButton,
		nil,
		nil,
		cardGrid,
	)
	return w
}

func (w *CardSelectWidget) Refresh() {
	log.Debug().Msg("amount of cards: " + fmt.Sprint(len(w.selectedCards)) + " with a limit of " + fmt.Sprint(w.draw.ToPick))
	if len(w.selectedCards) <= int(w.draw.ToPick) {
		log.Debug().Msg("valid amount of cards to pick selected")
		w.pickButton.Enable()
	} else {
		w.pickButton.Disable()
	}
	w.BaseWidget.Refresh()
}

func (w *CardSelectWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}

func (w *CardSelectWidget) AddCardToPool(cardDBID uint) {
	w.selectedCards = append(w.selectedCards, cardDBID)
}
func (w *CardSelectWidget) RemoveCardFromPool(cardDBID uint) {
	lengthBeforeDelete := len(w.selectedCards)
	for index, cardIDInList := range w.selectedCards {
		if cardIDInList == cardDBID {
			w.selectedCards = slices.Delete(w.selectedCards, index, index+1)
			log.Debug().Msg("deleting card at index " + fmt.Sprint(index))
		}
	}
	if len(w.selectedCards) != lengthBeforeDelete-1 {
		log.Warn().Msg("deletion of card with id " + fmt.Sprint(cardDBID) + " in list " + fmt.Sprint(w.selectedCards) + " failed")
	}
}
