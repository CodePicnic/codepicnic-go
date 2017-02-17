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
	"strings"
	"time"
	//"os"
)

const ERROR_NOT_AUTHORIZED = "Not Authorized"
const ERROR_NOT_CONNECTED = "Disconnected"
const ERROR_EMPTY_CREDENTIALS = "No Credentials"
const ERROR_EMPTY_TOKEN = "No Token"
const ERROR_INVALID_TOKEN = "Invalid Token"
const ERROR_USAGE_EXCEEDED = "Usage Exceeded"

var user_agent = "CodePicnic GO"

type codepicnic struct {
	ClientId     string
	ClientSecret string
	Token        string
	Consoles     []ConsoleJson
}

var cp codepicnic

type TokenJson struct {
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

type request struct {
	Method  string
	Url     string
	Headers map[string]string
}

func Init(client_id string, client_secret string) error {
	//TODO: return expiration
	cp = codepicnic{
		ClientId:     client_id,
		ClientSecret: client_secret,
	}
	var token TokenJson
	body, err := OAuthRequest()
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &token)
	cp.Token = token.Access
	return nil
}

func GetToken() (string, error) {
	return cp.Token, nil
}

func ListConsoles() ([]ConsoleJson, error) {
	var console_collection ConsoleCollection

	body, err := ApiRequest("/consoles/all.json", "GET")
	if err != nil {
		return console_collection.Consoles, err
	}
	err = json.Unmarshal(body, &console_collection)
	if err != nil {
		return console_collection.Consoles, err
	}
	//cp.Consoles = console_collection.Consoles
	return console_collection.Consoles, nil
}

func ApiRequest(endpoint string, method string) ([]byte, error) {
	var codepicnic_api = "https://codepicnic.com/api"

	cp_consoles_url := codepicnic_api + endpoint
	req, err := http.NewRequest(method, cp_consoles_url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cp.Token)
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func OAuthRequest() ([]byte, error) {
	var codepicnic_oauth = "https://codepicnic.com/oauth/token"
	cp_payload := `{ "grant_type": "client_credentials","client_id": "` + cp.ClientId + `", "client_secret": "` + cp.ClientSecret + `"}`
	var jsonStr = []byte(cp_payload)
	req, err := http.NewRequest("POST", codepicnic_oauth, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", user_agent)
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if err != nil {
		if strings.Contains(err.Error(), "read: connection refused") {
			return nil, errors.New(ERROR_NOT_CONNECTED)
		}
	}
	if resp.StatusCode == 401 {
		return nil, errors.New(ERROR_NOT_AUTHORIZED)
	}
	return body, nil
}

func GetConsole(console_id string) (ConsoleJson, error) {
	var console_found ConsoleJson
	consoles, err := ListConsoles()
	if err != nil {
		return console_found, err
	}
	for _, console := range consoles {
		if console.ContainerName == console_id {
			console_found = console
			break
		}
	}
	return console_found, nil
}

func (console *ConsoleJson) Start() error {
	cp_api_path := "/consoles/" + console.ContainerName + "/start"
	_, err := ApiRequest(cp_api_path, "POST")
	if err != nil {
		return err
	}
	return nil
}
func (console *ConsoleJson) Stop() error {
	cp_api_path := "/consoles/" + console.ContainerName + "/stop"
	_, err := ApiRequest(cp_api_path, "POST")
	if err != nil {
		return err
	}
	return nil
}
func (console *ConsoleJson) Restart() error {
	cp_api_path := "/consoles/" + console.ContainerName + "/restart"
	_, err := ApiRequest(cp_api_path, "POST")
	if err != nil {
		return err
	}
	return nil
}

func (console *ConsoleJson) Remove() error {
	cp_api_path := "/consoles" + "/" + console.ContainerName
	_, err := ApiRequest(cp_api_path, "DELETE")
	if err != nil {
		return err
	}
	return nil
}
