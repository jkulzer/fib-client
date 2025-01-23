package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/helpers"
	"github.com/jkulzer/fib-server/sharedModels"
)

func GetGamePhase(env env.Env, parentWindow fyne.Window) sharedModels.GamePhase {

	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		log.Err(err).Msg("failed to get app config for state chekcing")
	} else {
		req, err := http.NewRequest("GET", env.Url+"/lobby/"+loginInfo.LobbyToken+"/phase", nil)
		if err != nil {
			dialog.ShowError(err, parentWindow)
		} else {
			req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Warn().Msg("couldn't make request" + fmt.Sprint(err))
				dialog.ShowError(err, parentWindow)
			} else {
				switch res.StatusCode {
				case http.StatusOK:
					httpResponse, err := helpers.ReadHttpResponse(res.Body)
					if err != nil {
						log.Warn().Msg("failed to read get game phase response: " + fmt.Sprint(err))
					} else {
						var phaseResponse sharedModels.PhaseResponse
						err = json.Unmarshal(httpResponse, &phaseResponse)
						if err != nil {
							log.Warn().Msg("failed to unmarshal get game phase struct: " + fmt.Sprint(err))
						} else {
							return phaseResponse.Phase
						}
					}
				case http.StatusUnauthorized:
					error := errors.New("unauthenticated")
					log.Warn().Msg(fmt.Sprint(error))
					dialog.ShowError(error, parentWindow)
				default:
					error := errors.New("http error " + fmt.Sprint(res.StatusCode) + " when getting game state")
					log.Warn().Msg(fmt.Sprint(error))
					dialog.ShowError(error, parentWindow)
				}
			}
		}
	}
	return sharedModels.PhaseInvalid
}
