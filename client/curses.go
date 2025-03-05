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

func GetCurses(env env.Env, parentWindow fyne.Window) ([]sharedModels.Card, error) {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return []sharedModels.Card{}, err
	}

	req, err := http.NewRequest("GET", env.Url+"/lobby/"+loginInfo.LobbyToken+"/curses", nil)
	if err != nil {
		return []sharedModels.Card{}, err
	}
	req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []sharedModels.Card{}, err
	}

	byteBody, err := helpers.ReadHttpResponse(res.Body)
	if err != nil {
		return []sharedModels.Card{}, err
	}

	var cursesResponse sharedModels.CardList
	err = json.Unmarshal(byteBody, &cursesResponse)
	if err != nil {
		return []sharedModels.Card{}, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		return cursesResponse.List, nil
	case http.StatusBadRequest:
		return []sharedModels.Card{}, errors.New("Lobby doesn't exist")
	default:
		return []sharedModels.Card{}, errors.New("getting curses failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}
