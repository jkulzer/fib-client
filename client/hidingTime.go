package client

import (
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/helpers"

	"github.com/jkulzer/fib-server/sharedModels"

	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func RunStartTime(env env.Env, parentWindow fyne.Window) (time.Time, error) {
	currentTime := time.Now()
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return currentTime, err
	}
	req, err := http.NewRequest("GET", env.Url+"/lobby/"+loginInfo.LobbyToken+"/runStartTime", nil)
	if err != nil {
		return currentTime, err
	}
	req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return currentTime, err
	}
	switch res.StatusCode {
	case http.StatusOK:
		var responseStruct sharedModels.TimeResponse
		responseBytes, err := helpers.ReadHttpResponse(res.Body)
		if err != nil {
			return currentTime, err
		}
		err = json.Unmarshal(responseBytes, &responseStruct)
		if err != nil {
			return currentTime, err
		} else {
			return responseStruct.Time, nil
		}
	default:
		err := errors.New("request for run start time failed with http status code " + fmt.Sprint(res.StatusCode))
		dialog.ShowError(err, parentWindow)
		return currentTime, err
	}
}
