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

func GetCardActions(env env.Env, parentWindow fyne.Window) (sharedModels.CardDraws, error) {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return sharedModels.CardDraws{}, err
	}

	req, err := http.NewRequest("GET", env.Url+"/lobby/"+loginInfo.LobbyToken+"/cardActions", nil)
	if err != nil {
		return sharedModels.CardDraws{}, err
	}
	req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return sharedModels.CardDraws{}, err
	}

	byteBody, err := helpers.ReadHttpResponse(res.Body)
	if err != nil {
		return sharedModels.CardDraws{}, err
	}

	var actionsResponse sharedModels.CardDraws
	err = json.Unmarshal(byteBody, &actionsResponse)
	if err != nil {
		return sharedModels.CardDraws{}, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		return actionsResponse, nil
	case http.StatusBadRequest:
		return sharedModels.CardDraws{}, errors.New("Lobby doesn't exist")
	default:
		return sharedModels.CardDraws{}, errors.New("getting remaining card actions failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}

var ErrAlreadyDrewCards = errors.New("Already drew cards")

func DrawCards(env env.Env, parentWindow fyne.Window, drawID uint) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", env.Url+"/lobby/"+loginInfo.LobbyToken+"/drawCards/"+fmt.Sprint(drawID), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return errors.New("Lobby doesn't exist or invalid draw ID")
	case http.StatusConflict:
		return ErrAlreadyDrewCards
	default:
		return errors.New("drawing cards failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}

func GetDraw(env env.Env, parentWindow fyne.Window) (sharedModels.CurrentDraw, error) {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return sharedModels.CurrentDraw{}, err
	}

	req, err := http.NewRequest("GET", env.Url+"/lobby/"+loginInfo.LobbyToken+"/draw", nil)
	if err != nil {
		return sharedModels.CurrentDraw{}, err
	}
	req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return sharedModels.CurrentDraw{}, err
	}

	byteBody, err := helpers.ReadHttpResponse(res.Body)
	if err != nil {
		return sharedModels.CurrentDraw{}, err
	}

	var drawResponse sharedModels.CurrentDraw
	err = json.Unmarshal(byteBody, &drawResponse)
	if err != nil {
		return sharedModels.CurrentDraw{}, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		return drawResponse, nil
	case http.StatusBadRequest:
		return sharedModels.CurrentDraw{}, errors.New("Lobby doesn't exist")
	default:
		return sharedModels.CurrentDraw{}, errors.New("getting drawn cards failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}
