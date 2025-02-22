package client

import (
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/helpers"

	"github.com/jkulzer/fib-server/sharedModels"

	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"fyne.io/fyne/v2"
)

func GetHistory(env env.Env, parentWindow fyne.Window) (sharedModels.History, error) {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return sharedModels.History{}, err
	}

	req, err := http.NewRequest("GET", env.Url+"/lobby/"+loginInfo.LobbyToken+"/history", nil)
	if err != nil {
		return sharedModels.History{}, err
	}
	req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return sharedModels.History{}, err
	}

	byteBody, err := helpers.ReadHttpResponse(res.Body)
	if err != nil {
		return sharedModels.History{}, err
	}
	var historyResponse sharedModels.History
	err = json.Unmarshal(byteBody, &historyResponse)
	if err != nil {
		return sharedModels.History{}, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		return historyResponse, nil
	case http.StatusBadRequest:
		return sharedModels.History{}, errors.New("Lobby doesn't exist")
	case http.StatusForbidden:
		return sharedModels.History{}, errors.New("You are not the seeker and can't ask questions")
	default:
		return sharedModels.History{}, errors.New("getting history failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}
