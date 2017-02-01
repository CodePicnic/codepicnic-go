package codepicnic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

type CodePicnic struct {
	ClientId string
	SecretId string
	Token    string
	Consoles []ConsoleJson
}

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

func (cp *CodePicnic) GetToken() error {
	//TODO: return expiration
	cp_payload := `{ "grant_type": "client_credentials","client_id": "` + cp.ClientId + `", "client_secret": "` + cp.SecretId + `"}`
	var jsonStr = []byte(cp_payload)
	req, err := http.NewRequest("POST", codepicnic_oauth, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", user_agent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "read: connection refused") {
			return errors.New(ERROR_NOT_CONNECTED)
		}
	}
	if resp.StatusCode == 401 {
		return errors.New(ERROR_NOT_AUTHORIZED)
	}
	defer resp.Body.Close()
	var token TokenJson
	_ = json.NewDecoder(resp.Body).Decode(&token)
	cp.Token = token.Access
	return nil

}

func (cp *CodePicnic) ListConsoles() error {

	cp_consoles_url := codepicnic_api + "/api/consoles/all.json"
	req, err := http.NewRequest("GET", cp_consoles_url, nil)
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
		return errors.New(ERROR_INVALID_TOKEN)
	} else if resp.StatusCode == 429 {
		return errors.New(ERROR_USAGE_EXCEEDED)
	}
	var console_collection ConsoleCollection
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &console_collection)
	if err != nil {
		panic(err)
	}
	cp.Consoles = console_collection.Consoles
	return nil
}

func (cp *CodePicnic) GetConsole(console_id string) (ConsoleJson, error) {
	var console_found ConsoleJson
	for _, console := range cp.Consoles {
		if console.ContainerName == console_id {
			console_found = console
			fmt.Println("found")
			break
		}
	}
	return console_found, nil
}

func (cp *CodePicnic) StartConsole(console ConsoleJson) error {

	cp_consoles_url := codepicnic_api + "/api/consoles/" + console.ContainerName + "/start"
	req, err := http.NewRequest("POST", cp_consoles_url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cp.Token)
	req.Header.Set("User-Agent", user_agent)
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	if resp.StatusCode == 401 {
		return errors.New(ERROR_INVALID_TOKEN)
	} else if resp.StatusCode == 429 {
		return errors.New(ERROR_USAGE_EXCEEDED)
	}
	return nil
}
func (cp *CodePicnic) StopConsole(console ConsoleJson) error {

	cp_consoles_url := codepicnic_api + "/api/consoles/" + console.ContainerName + "/stop"
	req, err := http.NewRequest("POST", cp_consoles_url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cp.Token)
	req.Header.Set("User-Agent", user_agent)
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	if resp.StatusCode == 401 {
		return errors.New(ERROR_INVALID_TOKEN)
	} else if resp.StatusCode == 429 {
		return errors.New(ERROR_USAGE_EXCEEDED)
	}
	return nil
}
func (cp *CodePicnic) RestartConsole(console ConsoleJson) error {

	cp_consoles_url := codepicnic_api + "/api/consoles/" + console.ContainerName + "/restart"
	req, err := http.NewRequest("POST", cp_consoles_url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cp.Token)
	req.Header.Set("User-Agent", user_agent)
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	if resp.StatusCode == 401 {
		return errors.New(ERROR_INVALID_TOKEN)
	} else if resp.StatusCode == 429 {
		return errors.New(ERROR_USAGE_EXCEEDED)
	}
	return nil
}

func (cp *CodePicnic) RemoveConsole(console ConsoleJson) error {
	cp_consoles_url := codepicnic_api + "/api/consoles" + "/" + console.ContainerName
	var jsonStr = []byte("")
	req, err := http.NewRequest("DELETE", cp_consoles_url, bytes.NewBuffer(jsonStr))
	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Authorization", "Bearer "+cp.Token)
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return errors.New(ERROR_INVALID_TOKEN)
	} else if resp.StatusCode == 429 {
		return errors.New(ERROR_USAGE_EXCEEDED)
	}
	if err != nil {
		panic(err)
	}
	return nil
}
