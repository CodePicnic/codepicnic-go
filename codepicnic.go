package codepicnic

import (
	"bytes"
	"encoding/json"
	"errors"
	//"fmt"
	//"github.com/Jeffail/gabs"
	//"io"
	"io/ioutil"
	//"mime/multipart"
	"net/http"
	//"os"
	"strings"
)

const ERROR_NOT_AUTHORIZED = "Not Authorized"
const ERROR_NOT_CONNECTED = "Disconnected"
const ERROR_EMPTY_CREDENTIALS = "No Credentials"
const ERROR_EMPTY_TOKEN = "No Token"
const ERROR_INVALID_TOKEN = "Invalid Token"
const ERROR_USAGE_EXCEEDED = "Usage Exceeded"

var user_agent = "CodePicnic GO"

var codepicnic_api = "https://codepicnic.com"
var codepicnic_oauth = "https://codepicnic.com/oauth/token"

type Token struct {
	Access  string `json:"access_token"`
	Type    string `json:"token_type"`
	Expires string `json:"expires_in"`
	Created string `json:"created_at"`
}

type ConsoleJson struct {
	Id            int    `json:"id"`
	Content       string `json:"content"`
	Title         string `json:"title"`
	Name          string `json:"name"`
	ContainerName string `json:"container_name"`
	ContainerType string `json:"container_type"`
	CustomImage   string `json:"custom_image"`
	CreatedAt     string `json:"created_at"`
	Permalink     string `json:"permalink"`
	//Url           string `json:"url"`
	//TerminalUrl   string `json:"terminal_url"`
}

type ConsoleCollection struct {
	Consoles []ConsoleJson `json:"consoles"`
}

func GetTokenAccess(client_id string, client_secret string) (string, error) {
	//TODO: return expiration
	cp_payload := `{ "grant_type": "client_credentials","client_id": "` + client_id + `", "client_secret": "` + client_secret + `"}`
	var jsonStr = []byte(cp_payload)
	req, err := http.NewRequest("POST", codepicnic_oauth, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", user_agent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "read: connection refused") {
			return "", errors.New(ERROR_NOT_CONNECTED)
		}
	}
	if resp.StatusCode == 401 {
		return "", errors.New(ERROR_NOT_AUTHORIZED)
	}
	defer resp.Body.Close()
	var token Token
	_ = json.NewDecoder(resp.Body).Decode(&token)
	return token.Access, nil
}

func ListConsoles(access_token string) ([]ConsoleJson, error) {

	cp_consoles_url := codepicnic_api + "/api/consoles/all.json"
	req, err := http.NewRequest("GET", cp_consoles_url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+access_token)
	req.Header.Set("User-Agent", user_agent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return nil, errors.New(ERROR_INVALID_TOKEN)
	} else if resp.StatusCode == 429 {
		return nil, errors.New(ERROR_USAGE_EXCEEDED)
	}
	var console_collection ConsoleCollection
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &console_collection)
	if err != nil {
		panic(err)
	}
	return console_collection.Consoles, nil
}
