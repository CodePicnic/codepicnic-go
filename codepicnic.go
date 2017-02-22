package codepicnic

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Jeffail/gabs"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const ERROR_NOT_AUTHORIZED = "Not Authorized"
const ERROR_NOT_CONNECTED = "Disconnected"
const ERROR_NOT_FOUND = "Not Found"
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
}

func Init(client_id string, client_secret string) error {
	//TODO: return expiration
	cp = codepicnic{
		ClientId:     client_id,
		ClientSecret: client_secret,
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

func GetToken() (string, error) {
	return cp.Token, nil
}

func ListConsoles() ([]ConsoleJson, error) {
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
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "POST",
	}
	_, err := api.Send()
	if err != nil {
		return err
	}
	return nil
}
func (console *ConsoleJson) Stop() error {
	cp_api_path := "/consoles/" + console.ContainerName + "/stop"
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "POST",
	}
	_, err := api.Send()
	if err != nil {
		return err
	}
	return nil
}
func (console *ConsoleJson) Restart() error {
	cp_api_path := "/consoles/" + console.ContainerName + "/restart"
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "POST",
	}
	_, err := api.Send()
	if err != nil {
		return err
	}
	return nil
}

func (console *ConsoleJson) Show() ([]byte, error) {
	cp_api_path := "/consoles/" + console.ContainerName + "/show"
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "POST",
	}
	body, err := api.Send()
	if err != nil {
		return "", err
	}
	return body, nil
}

func (console *ConsoleJson) Remove() error {
	cp_api_path := "/consoles" + "/" + console.ContainerName
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "DELETE",
	}
	_, err := api.Send()
	if err != nil {
		return err
	}
	return nil
}

func (console *ConsoleJson) Exec(command string) ([]CommandJson, error) {
	var CmdCollection []CommandJson
	cp_api_path := "/consoles" + "/" + console.ContainerName + "/exec"
	cp_payload := ` { "commands": "` + command + `" }`
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "POST",
		Payload:  cp_payload,
	}
	body, err := api.Send()
	if err != nil {
		return CmdCollection, err
	}
	jsonBody, err := gabs.ParseJSON(body)
	if err != nil {
		return CmdCollection, err
	}
	jsonPaths, _ := jsonBody.ChildrenMap()
	for key, child := range jsonPaths {
		var cmd CommandJson
		cmd.command = string(key)
		cmd.result = child.Data().(string)
		CmdCollection = append(CmdCollection, cmd)
	}
	return CmdCollection, nil
}

func (console *ConsoleJson) ReadFile(file string) ([]byte, error) {
	cp_api_path := "/consoles" + "/" + console.ContainerName + "/read_file?path=" + file
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "GET",
	}
	body, err := api.Send()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (console *ConsoleJson) Search(term string) ([]FileJson, error) {
	var file_collection []FileJson
	cp_api_path := "/consoles" + "/" + console.ContainerName + "/search?term=" + term
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "GET",
	}
	body, err := api.Send()
	if err != nil {
		return file_collection, err
	}
	/*
		err = json.Unmarshal(body, &file_collection)
		if err != nil {
			return file_collection, err
		}
		return file_collection, nil*/
	jsonBody, err := gabs.ParseJSON(body)
	if err != nil {
		return file_collection, err
	}
	jsonPaths, _ := jsonBody.Children()
	for _, child := range jsonPaths {
		file := child.Data().(map[string]interface{})
		f := FileJson{
			Name: file["name"].(string),
			Size: file["size"].(float64),
			Type: file["type"].(string),
			Path: file["path"].(string),
		}
		file_collection = append(file_collection, f)
	}

	/*
		for key, child := range jsonPaths {
			var file FileJson
			file.Name = string(key)
			file.Data = child.Data().(map[string]interface)
			fmt.Printf("%+v", file.Data)
			//file.Path = child.Data().(map[string]string)
			file_collection = append(file_collection, file)
		}*/
	return file_collection, nil
}

func (console *ConsoleJson) UploadFile(src_file string, dst_file string) ([]byte, error) {
	cp_api_path := "/consoles" + "/" + console.ContainerName + "/upload_file"
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "POST",
	}
	body, err := api.Upload(src_file, dst_file)
	if err != nil {
		return nil, err
	}
	return body, nil
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
	req.Header.Set("User-Agent", user_agent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return nil, errors.New(ERROR_INVALID_TOKEN)
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
