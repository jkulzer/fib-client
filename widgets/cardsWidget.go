package widgets

import (
	"errors"
	"fmt"
	"image/color"
	"slices"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
	w := &CardsWidget{
		env:          env,
		parentWindow: parentWindow,
	}
	w.ExtendBaseWidget(w)
	w.content = container.NewVBox()

	return w
}

func (w *CardsWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewScroll(w.content))
}

func (w *CardsWidget) SetContent() error {
	w.content.RemoveAll()
	log.Info().Msg("refreshing card widget")
	cardActions, err := client.GetCardActions(w.env, w.parentWindow)
	if err != nil {
		dialog.ShowError(err, w.parentWindow)
		return err
	}
	log.Debug().Msg(fmt.Sprint("card draws: ", cardActions))
	draw, err := client.GetDraw(w.env, w.parentWindow)
	if err != nil {
		log.Err(err).Msg("failed getting draw")
		dialog.ShowError(err, w.parentWindow)
		return err
	}
	if len(draw.Cards) > 0 {
		w.content.Add(
			widget.NewButton("Resume in progress draw", func() {
				cardSelectDialog(w.env, w.parentWindow, w)
			}),
		)
	}
	if len(cardActions.Draws) == 0 {
		w.content.Add(widget.NewLabel("no remaining draws"))
	} else {
		drawContainer := container.NewVBox()
		for _, cardAction := range cardActions.Draws {
			log.Debug().Msg("found a card draw")
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
						cardSelectDialog(w.env, w.parentWindow, w)
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
	}
	cardGrid := container.NewGridWithRows(2)
	for _, handCard := range hiderHand.List {
		cardGrid.Add(NewCardWidget(handCard, PlayCardWidget, nil, 0, w.env, w.parentWindow, w))
	}
	w.content.Add(widget.NewLabel("Your hand:"))
	w.content.Add(container.NewHScroll(cardGrid))
	w.content.Refresh()
	return nil
}

func (w *CardsWidget) Refresh() {
	w.SetContent()
	w.BaseWidget.Refresh()
}

func cardSelectDialog(env env.Env, parentWindow fyne.Window, cardsWidget *CardsWidget) {
	var cardDrawDialog *dialog.CustomDialog
	drawnCardsContainer := NewCardSelectWidget(env, parentWindow, &cardDrawDialog, cardsWidget)
	cardDrawDialog = dialog.NewCustom("Drawn cards", "dismiss", drawnCardsContainer, parentWindow)
	cardDrawDialog.Resize(fyne.NewSize(300, 600))
	cardDrawDialog.Show()
}

type CardWidget struct {
	widget.BaseWidget
	content          *fyne.Container
	cardIndex        uint
	selected         bool
	selectedText     *widget.Label
	env              env.Env
	parentWindow     fyne.Window
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

func NewCardWidget(card sharedModels.Card, widgetType CardWidgetType, cardSelectWidget *CardSelectWidget, cardIndex uint, env env.Env, parentWindow fyne.Window, cardsWidget *CardsWidget) *CardWidget {
	w := &CardWidget{
		cardIndex:        cardIndex,
		selectedText:     widget.NewLabel("selected"),
		card:             card,
		cardSelectWidget: cardSelectWidget,
		widgetType:       widgetType,
		env:              env,
		parentWindow:     parentWindow,
	}
	w.ExtendBaseWidget(w)

	title := widget.NewLabel(card.Title)
	title.TextStyle = fyne.TextStyle{
		Bold:      true,
		Underline: true,
	}
	var description *widget.Label
	if w.card.Description != "" {
		description = widget.NewLabel(w.card.Description)
	} else {
		description = widget.NewLabel("No Description")
	}

	var castingCost *widget.Label
	if w.card.CastingCostDescription != "" {
		castingCost = widget.NewLabel("Casting cost: " + w.card.CastingCostDescription)
	} else {
		castingCost = widget.NewLabel("Casting cost: No casting cost")
	}
	details := container.NewVBox(description, castingCost)
	w.content = container.NewBorder(
		title,
		widget.NewButton("Discard card", func() {
			dialog.ShowConfirm("Discard card", "Are you sure you want to discard this card?", func(confirmed bool) {
				if confirmed {
					err := client.DiscardCard(env, parentWindow, w.card.IDInDB)
					if err != nil {
						log.Err(err).Msg("failed discarding card")
						dialog.ShowError(err, parentWindow)
					}
					cardsWidget.Refresh()
				}
			}, parentWindow)
		}),
		nil,
		nil,
		details,
	)

	// description.Wrapping = fyne.TextWrapBreak

	return w
}

func (w *CardWidget) CreateRenderer() fyne.WidgetRenderer {
	bgRectangle := canvas.NewRectangle(color.RGBA{150, 150, 150, 250})
	bgRectangle.Resize(fyne.NewSize(100, 300))
	return widget.NewSimpleRenderer(container.NewStack(bgRectangle, w.content))
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
	case PlayCardWidget:
		dialog.ShowConfirm("Card", "Are you sure you want to play this card?", func(confirmed bool) {
			if confirmed {
				client.PlayCard(w.env, w.parentWindow, w.card.IDInDB)
			}
		}, w.parentWindow)
	}
}

type CardSelectWidget struct {
	widget.BaseWidget
	content        *fyne.Container
	env            env.Env
	draw           sharedModels.CurrentDraw
	parentWindow   fyne.Window
	selectedCards  []uint
	pickButton     *widget.Button
	cardDrawDialog *dialog.CustomDialog
	cardsWidget    *CardsWidget
}

func NewCardSelectWidget(env env.Env, parentWindow fyne.Window, cardDrawDialog **dialog.CustomDialog, cardsWidget *CardsWidget) *CardSelectWidget {
	w := &CardSelectWidget{
		env:            env,
		parentWindow:   parentWindow,
		cardDrawDialog: *cardDrawDialog,
		cardsWidget:    cardsWidget,
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
		w.cardsWidget.Refresh()
	})
	cardGrid := container.NewGridWithRows(2)
	for cardIndex, card := range w.draw.Cards {
		cardWidget := NewCardWidget(card, DrawCardWidget, w, uint(cardIndex), env, parentWindow, w.cardsWidget)
		cardGrid.Add(cardWidget)
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
