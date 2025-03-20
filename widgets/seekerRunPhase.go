package widgets

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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

	go func() {
		for {
			timer := time.NewTimer(2 * time.Second)
			<-timer.C

			gamePhase := client.GetGamePhase(env, parentWindow)
			if gamePhase == sharedModels.PhaseLocationNarrowing {
				log.Info().Msg("now in location narrowing phase")
				narrowingPhaseWidget := NewSeekerNarrowingPhaseWidget(env, parentWindow)
				gameFrame := NewGameFrameWidget(env, parentWindow, narrowingPhaseWidget)
				parentWindow.SetContent(gameFrame)
				return
			}
		}
	}()

	return w
}

func (w *SeekerRunPhaseWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
