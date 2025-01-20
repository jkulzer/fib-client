package helpers

import (
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/models"

	"github.com/jkulzer/fib-server/sharedModels"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/rs/zerolog/log"

	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func ReadHttpResponse(input io.ReadCloser) ([]byte, error) {
	if b, err := io.ReadAll(input); err == nil {
		return b, err
	} else {
		return nil, err
	}
}

func ReadHttpResponseToString(input io.ReadCloser) (string, error) {
	if b, err := io.ReadAll(input); err == nil {
		return string(b), err
	} else {
		return "", err
	}
}

func GetAppConfig(env env.Env, parentWindow fyne.Window) (models.LoginInfo, error) {
	var loginInfo models.LoginInfo
	result := env.DB.First(&loginInfo)
	if result.Error != nil {
		log.Err(result.Error)
		return models.LoginInfo{}, result.Error
	} else if loginInfo.Token.String() == models.NullUuidString {
		log.Warn().Msg("auth token uuid string in app config is null")
		log.Debug().Msg(fmt.Sprint(loginInfo))
		// dialog.ShowError(result.Error, parentWindow)
		return models.LoginInfo{}, result.Error
	} else {
		return loginInfo, nil
	}
}

func GetGamePhase(env env.Env, parentWindow fyne.Window) sharedModels.GamePhase {

	loginInfo, err := GetAppConfig(env, parentWindow)
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
					httpResponse, err := ReadHttpResponse(res.Body)
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
