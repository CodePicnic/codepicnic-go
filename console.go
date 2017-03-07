package codepicnic

import (
	"github.com/Jeffail/gabs"
	"time"
)

type Console struct {
	id            int    `json:"id"`
	content       string `json:"content"`
	title         string `json:"title"`
	name          string `json:"name"`
	containerName string `json:"container_name"`
	containerType string `json:"container_type"`
	customImage   string `json:"custom_image"`
	createdAt     string `json:"created_at"`
	permalink     string `json:"permalink"`
	url           string `json:"url"`
	embedUrl      string `json:"embed_url"`
	terminalUrl   string `json:"terminal_url"`
	isHeadless    string `json:"is_headless"`
}

type ConsoleCollection struct {
	Consoles []Console `json:"consoles"`
}

type ConsoleRequest struct {
	Title    string
	Size     string
	Type     string
	Hostname string
	Mode     string
}

func (console *Console) Title() string {
	return console.title
}
func (console *Console) Name() string {
	return console.name
}
func (console *Console) ContainerName() string {
	return console.containerName
}
func (console *Console) ContainerType() string {
	return console.containerType
}
func (console *Console) Permalink() string {
	return console.permalink
}
func (console *Console) Url() string {
	return console.url
}
func (console *Console) EmbedUrl() string {
	return console.embedUrl
}
func (console *Console) TerminalUrl() string {
	return console.terminalUrl
}

func (console *Console) Status() (string, error) {
	cp_api_path := "/consoles/" + console.containerName + "/status"
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "GET",
		Timeout:  time.Second * 10,
	}
	body, err := api.Send()
	if err != nil {
		return "", err
	}
	jsonBody, err := gabs.ParseJSON(body)
	if err != nil {
		return "", err
	}
	status, ok := jsonBody.Path("state.status").Data().(string)
	if ok == false {
		return "", err
	}
	return status, nil
}

func (console *Console) Start() error {
	cp_api_path := "/consoles/" + console.containerName + "/start"
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
func (console *Console) Stop() error {
	cp_api_path := "/consoles/" + console.containerName + "/stop"
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
func (console *Console) Restart() error {
	cp_api_path := "/consoles/" + console.containerName + "/restart"
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

func (console *Console) Remove() error {
	cp_api_path := "/consoles" + "/" + console.containerName
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

func (console *Console) Exec(command string) ([]CommandJson, error) {
	var CmdCollection []CommandJson
	cp_api_path := "/consoles" + "/" + console.containerName + "/exec"
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

func (console *Console) ReadFile(file string) ([]byte, error) {
	cp_api_path := "/consoles" + "/" + console.containerName + "/read_file?path=" + file
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

func (console *Console) Search(term string) ([]FileJson, error) {
	var file_collection []FileJson
	cp_api_path := "/consoles" + "/" + console.containerName + "/search?term=" + term
	api := ApiRequest{
		Endpoint: cp_api_path,
		Method:   "GET",
	}
	body, err := api.Send()
	if err != nil {
		return file_collection, err
	}
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

	return file_collection, nil
}

func (console *Console) UploadFile(src_file string, dst_file string) ([]byte, error) {
	cp_api_path := "/consoles" + "/" + console.containerName + "/upload_file"
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
