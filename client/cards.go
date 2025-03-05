package client

import (
	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/helpers"

	"github.com/jkulzer/fib-server/sharedModels"

	"bytes"
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

func PickCards(env env.Env, parentWindow fyne.Window, cardDBID []uint) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	marshaledBody, err := json.Marshal(sharedModels.CardIDList{CardIDList: cardDBID})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", env.Url+"/lobby/"+loginInfo.LobbyToken+"/pickFromDraw", bytes.NewReader(marshaledBody))
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
		return errors.New("You can't draw cards, it would exceed your maximum hand size of " + fmt.Sprint(sharedModels.MaxHandSize))
	default:
		return errors.New("picking cards failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}

func GetHiderHand(env env.Env, parentWindow fyne.Window) (sharedModels.CardList, error) {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return sharedModels.CardList{}, err
	}

	req, err := http.NewRequest("GET", env.Url+"/lobby/"+loginInfo.LobbyToken+"/hiderHand", nil)
	if err != nil {
		return sharedModels.CardList{}, err
	}
	req.Header.Add("Authorization", "Bearer "+loginInfo.Token.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return sharedModels.CardList{}, err
	}

	byteBody, err := helpers.ReadHttpResponse(res.Body)
	if err != nil {
		return sharedModels.CardList{}, err
	}

	var cardList sharedModels.CardList
	err = json.Unmarshal(byteBody, &cardList)
	if err != nil {
		return sharedModels.CardList{}, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		return cardList, nil
	case http.StatusBadRequest:
		return sharedModels.CardList{}, errors.New("Lobby doesn't exist")
	default:
		return sharedModels.CardList{}, errors.New("getting hider deck failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}

func DiscardCard(env env.Env, parentWindow fyne.Window, cardToDiscard uint) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", env.Url+"/lobby/"+loginInfo.LobbyToken+"/discardCard/"+fmt.Sprint(cardToDiscard), nil)
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
		return errors.New("Lobby doesn't exist or invalid card ID")
	default:
		return errors.New("discarding card failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}

var ErrBadRequestCard error = errors.New("Lobby doesn't exist or invalid card ID")

func PlayCard(env env.Env, parentWindow fyne.Window, cardToPlay uint) error {
	loginInfo, err := helpers.GetAppConfig(env, parentWindow)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", env.Url+"/lobby/"+loginInfo.LobbyToken+"/playCard/"+fmt.Sprint(cardToPlay), nil)
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
		return ErrBadRequestCard
	default:
		return errors.New("playing card failed with http status code " + fmt.Sprint(res.StatusCode))
	}
}
