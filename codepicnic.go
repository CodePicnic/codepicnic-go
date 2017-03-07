package codepicnic

import (
	"bytes"
	"encoding/json"
	"errors"
	//"fmt"
	"github.com/Jeffail/gabs"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const ERROR_NOT_AUTHORIZED = "Not Authorized"
const ERROR_NOT_CONNECTED = "Disconnected"
const ERROR_CONNECTION_REFUSED = "Connection Refused"
const ERROR_TCP_TIMEOUT = "TCP Timeout"
const ERROR_CLIENT_TIMEOUT = "Client Timeout"
const ERROR_TLS_TIMEOUT = "TLS Handshake Timeout"
const ERROR_DNS_LOOKUP = "Host not found"
const ERROR_NOT_FOUND = "Not Found"
const ERROR_EMPTY_CREDENTIALS = "No Credentials"
const ERROR_EMPTY_TOKEN = "No Token"
const ERROR_INVALID_TOKEN = "Invalid Token"
const ERROR_RENEW_TOKEN = "Token not Renewed"
const ERROR_USAGE_EXCEEDED = "Usage Exceeded"
const ERROR_CONSOLE = "Console Error"

type codepicnic struct {
	ClientId     string
	ClientSecret string
	Token        string
	UserAgent    string
}

var cp codepicnic

type TokenJson struct {
	Access  string `json:"access_token"`
	Type    string `json:"token_type"`
	Expires string `json:"expires_in"`
	Created string `json:"created_at"`
}

type StackJson struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	ShortName  string `json:"short_name"`
	Version    string `json:"version"`
	ImageName  string `json:"image_name"`
	Group      string `json:"group"`
}

type StackCollection struct {
	Stacks []StackJson `json:"container_types"`
}

type CommandJson struct {
	command string
	result  string
}

type FileJson struct {
	Name string  `json:"name"`
	Path string  `json:"path"`
	Type string  `json:"type"`
	Size float64 `json:"size"`
}

type ApiRequest struct {
	Method   string
	Endpoint string
	Payload  string
	Timeout  time.Duration
}

func Init(client_id string, client_secret string) error {
	//TODO: return expiration
	cp = codepicnic{
		ClientId:     client_id,
		ClientSecret: client_secret,
		UserAgent:    "CodePicnic GO",
	}
	var token TokenJson
	body, err := oauthRequest()
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &token)
	cp.Token = token.Access
	return nil
}

func RefreshToken() error {
	var token TokenJson
	body, err := oauthRequest()
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

func SetUserAgent(user_agent string) error {
	cp.UserAgent = user_agent
	return nil
}

func ListConsoles() ([]Console, error) {
	var console_collection ConsoleCollection

	api := ApiRequest{
		Endpoint: "/consoles/all.json",
		Method:   "GET",
	}
	body, err := api.Send()
	if err != nil {
		return console_collection.Consoles, err
	}
	err = json.Unmarshal(body, &console_collection)
	if err != nil {
		return console_collection.Consoles, err
	}
	return console_collection.Consoles, nil
}

func GetConsole(console_id string) (Console, error) {
	var console Console
	cp_api_path := "/consoles/" + console_id
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "GET",
	}
	body, err := api.Send()
	if err != nil {
		return console, err
	}
	jsonBody, err := gabs.ParseJSON(body)
	if err != nil {
		return console, err
	}
	console_json, err := jsonBody.Search("console").ChildrenMap()
	if err != nil {
		return console, err
	}
	console = Console{
		id:            int(console_json["id"].Data().(float64)),
		content:       sanitize(console_json["content"].Data()),
		title:         sanitize(console_json["title"].Data().(string)),
		name:          sanitize(console_json["name"].Data().(string)),
		containerName: sanitize(console_json["container_name"].Data().(string)),
		containerType: sanitize(console_json["container_type"].Data().(string)),
		customImage:   sanitize(console_json["created_at"].Data().(string)),
		createdAt:     sanitize(console_json["created_at"].Data().(string)),
		permalink:     sanitize(console_json["permalink"].Data().(string)),
		url:           sanitize(console_json["url"].Data().(string)),
		embedUrl:      sanitize(console_json["embed_url"].Data().(string)),
		terminalUrl:   sanitize(console_json["terminal_url"].Data().(string)),
	}
	return console, nil
}

func CreateConsole(console_req ConsoleRequest) (Console, error) {
	var console Console
	if console_req.Size == "" {
		console_req.Size = "medium"
	}
	if console_req.Type == "" {
		console_req.Type = "bash"
	}
	if console_req.Mode == "" {
		console_req.Mode = "draft"
	}

	cp_api_path := "/consoles"
	cp_payload := ` { "console": { "container_size": "` + console_req.Size + `", "container_type": "` + console_req.Type + `", "title": "` + console_req.Title + `" , "hostname": "` + console_req.Hostname + `", "current_mode": "` + console_req.Mode + `" }  }`
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "POST",
		Payload:  cp_payload,
	}
	body, err := api.Send()
	if err != nil {
		return console, err
	}
	jsonBody, err := gabs.ParseJSON(body)
	if err != nil {
		return console, err
	}
	container_name, ok := jsonBody.Search("container_name").Data().(string)
	if ok != true {
		return console, errors.New(ERROR_CONSOLE)
	}
	console, err = GetConsole(container_name)
	if err != nil {
		return console, err
	}

	return console, nil

}

func ListStacks() ([]StackJson, error) {
	var stack_collection StackCollection

	api := ApiRequest{
		Endpoint: "/container_types.json",
		Method:   "GET",
	}
	body, err := api.Send()
	err = json.Unmarshal(body, &stack_collection)
	if err != nil {
		return stack_collection.Stacks, err
	}
	return stack_collection.Stacks, nil
}

func (api *ApiRequest) Send() ([]byte, error) {
	var codepicnic_api = "https://codepicnic.com/api"
	var req *http.Request
	cp_api_url := codepicnic_api + api.Endpoint
	if len(api.Payload) > 0 {
		var jsonStr = []byte(api.Payload)
		req, _ = http.NewRequest(api.Method, cp_api_url, bytes.NewBuffer(jsonStr))
	} else {
		req, _ = http.NewRequest(api.Method, cp_api_url, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cp.Token)
	req.Header.Set("User-Agent", cp.UserAgent)
	/*
		var api_timeout time.Duration
		if api.Timeout > time.Second * 0 {
			api_timeout = api.Timeout
		} else {
			//default timeout
			api_timeout = time.Second * 60
		}*/
	var api_transport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	client := &http.Client{
		Timeout:   time.Second * 20,
		Transport: api_transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "no such host") {
			return nil, errors.New(ERROR_DNS_LOOKUP)
		} else if strings.Contains(err.Error(), "connection refused") {
			return nil, errors.New(ERROR_CONNECTION_REFUSED)
		} else if strings.Contains(err.Error(), "dial tcp: i/o timeout") {
			return nil, errors.New(ERROR_TCP_TIMEOUT)
		} else if strings.Contains(err.Error(), "exceeded while awaiting headers") {
			return nil, errors.New(ERROR_CLIENT_TIMEOUT)
		} else if strings.Contains(err.Error(), "TLS handshake timeout") {
			return nil, errors.New(ERROR_TLS_TIMEOUT)
		} else {
			return nil, err
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		err = RefreshToken()
		if err != nil {
			return nil, errors.New(ERROR_RENEW_TOKEN)
		} else {
			req.Header.Set("Authorization", "Bearer "+cp.Token)
			resp, err = client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
		}
		//return nil, errors.New(ERROR_INVALID_TOKEN)
	} else if resp.StatusCode == 429 {
		return nil, errors.New(ERROR_USAGE_EXCEEDED)
	} else if resp.StatusCode == 404 {
		return nil, errors.New(ERROR_NOT_FOUND)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (api *ApiRequest) Upload(src_file string, dst_file string) ([]byte, error) {
	var codepicnic_api = "https://codepicnic.com/api"
	var req *http.Request
	cp_api_url := codepicnic_api + api.Endpoint
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	temp_file, err := os.Open(src_file)
	if err != nil {
		return nil, err
	}
	fw, err := w.CreateFormFile("file", temp_file.Name())
	if _, err = io.Copy(fw, temp_file); err != nil {
		return nil, err
	}
	if fw, err = w.CreateFormField("path"); err != nil {
		return nil, err
	}
	if _, err = fw.Write([]byte("/app/" + dst_file)); err != nil {
		return nil, err
	}
	w.Close()
	req, err = http.NewRequest("POST", cp_api_url, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+cp.Token)
	req.Header.Set("User-Agent", cp.UserAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		err = RefreshToken()
		if err != nil {
			return nil, errors.New(ERROR_RENEW_TOKEN)
		} else {
			req.Header.Set("Authorization", "Bearer "+cp.Token)
			resp, err = client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
		}
	} else if resp.StatusCode == 429 {
		return nil, errors.New(ERROR_USAGE_EXCEEDED)
	} else if resp.StatusCode == 404 {
		return nil, errors.New(ERROR_NOT_FOUND)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func oauthRequest() ([]byte, error) {
	var codepicnic_oauth = "https://codepicnic.com/oauth/token"
	cp_payload := `{ "grant_type": "client_credentials","client_id": "` + cp.ClientId + `", "client_secret": "` + cp.ClientSecret + `"}`
	var jsonStr = []byte(cp_payload)
	req, err := http.NewRequest("POST", codepicnic_oauth, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", cp.UserAgent)
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
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

func sanitize(i interface{}) string {
	if i == nil {
		return ""
	} else {
		return i.(string)
	}

}
