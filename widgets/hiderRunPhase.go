package widgets

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	// "fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/jkulzer/fib-client/client"
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/location"

	"github.com/jkulzer/fib-server/sharedModels"
)

type HiderRunPhaseWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

func NewHiderRunPhaseWidget(env env.Env, parentWindow fyne.Window) *HiderRunPhaseWidget {
	w := &HiderRunPhaseWidget{}
	w.ExtendBaseWidget(w)

	saveLocationButton := widget.NewButton("Save Hiding Zone", func() {
		go func() {
			point, err := location.GetLocation(parentWindow)
			if err != nil {
				log.Err(err).Msg(fmt.Sprint(err) + " failed getting location in run phase widget")
				dialog.ShowError(err, parentWindow)
			}
			err = client.ValidateAndSetHidingZone(env, parentWindow, point)
			if err != nil {
				log.Err(err).Msg(fmt.Sprint(err))
				dialog.ShowError(err, parentWindow)
				return
			}
			dialog.ShowInformation("Location", "Saved hiding zone location", parentWindow)

		}()
	})

	w.content = container.NewVBox(
		widget.NewLabel("You need to save a hiding location before the hiding time ends.\nCurrently, the app can't request a new location, you need to use another app that uses GPS to get your current location"),
		saveLocationButton,
	)

	runStartTime, err := client.RunStartTime(env, parentWindow)
	if err != nil {
		log.Err(err).Msg(fmt.Sprint(err))
		dialog.ShowError(err, parentWindow)
	}

	// Create text object with large font
	countdownText := canvas.NewText("Countdown initializing", theme.ForegroundColor())
	countdownText.Alignment = fyne.TextAlignCenter
	countdownText.TextSize = 48 // Big font size
	countdownText.TextStyle = fyne.TextStyle{Bold: true}

	// Create centered container with padding
	centered := container.New(
		layout.NewPaddedLayout(),
		container.NewCenter(
			countdownText,
		),
	)

	w.content.Add(centered)

	go func() {
		endTime := runStartTime.Add(sharedModels.RunDuration)
		ticker := time.NewTicker(50 * time.Millisecond) // Smooth animation
		defer ticker.Stop()

		updateText := func(s string) {
			countdownText.Text = s
			canvas.Refresh(countdownText) // Force redraw
		}

		for {
			select {
			case <-ticker.C:
				remaining := time.Until(endTime)
				if remaining <= 0 {
					updateText("RUN PHASE DOWN")
					return
				}

				// Format with milliseconds
				remaining = remaining.Truncate(10 * time.Millisecond)
				hours := int(remaining.Hours())
				minutes := int(remaining.Minutes()) % 60
				seconds := int(remaining.Seconds()) % 60
				millis := remaining.Milliseconds() % 1000 / 10

				var timeStr string
				if hours > 0 {
					timeStr = fmt.Sprintf("%02d:%02d:%02d.%02d", hours, minutes, seconds, millis)
				} else {
					timeStr = fmt.Sprintf("%02d:%02d.%02d", minutes, seconds, millis)
				}

				updateText(timeStr)
			}
		}
	}()

	// str := binding.NewString()
	// str.Set("Countdown initializing")
	//
	// text := widget.NewLabelWithData(str)
	//
	// w.content.Add(text)
	//
	// go func() {
	//
	// 	for {
	// 		timer := time.NewTimer(16 * time.Millisecond)
	// 		<-timer.C
	//
	// 		countdown := time.Until(runStartTime.Add(sharedModels.RunDuration))
	//
	// 		countdownString := countdown.Truncate(10 * time.Millisecond).String()
	//
	// 		str.Set(countdownString)
	// 	}
	// }()
	//
	// go func() {
	// 	for {
	// 		timer := time.NewTimer(2 * time.Second)
	// 		<-timer.C
	//
	// 		gamePhase := client.GetGamePhase(env, parentWindow)
	// 		if gamePhase == sharedModels.PhaseLocationNarrowing {
	// 			log.Info().Msg("now in location narrowing phase")
	// 			narrowingPhaseWidget := NewHiderNarrowingPhaseWidget(env, parentWindow)
	// 			gameFrame := NewGameFrameWidget(env, parentWindow, narrowingPhaseWidget)
	// 			parentWindow.SetContent(gameFrame)
	// 			return
	// 		}
	// 	}
	// }()

	return w
}

func (w *HiderRunPhaseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
