# CodePicnic library for Go

## Instalation

```javascript
go get "github.com/CodePicnic/codepicnic-go"
```

## Usage

### Initialization

```golang

package main

import (
    "fmt"
    "github.com/CodePicnic/codepicnic-go"
)

func main() {
    client_id := "XXXXXXXXXXXXXX"
    secret_id := "YYYYYYYYYYYYYY"

    err := codepicnic.Init(client_id, secret_id)
    if err != nil {
        fmt.Println(err.Error())
    } else {
        token, _ := codepicnic.GetToken()
        fmt.Println(token)
    }
}

```

### Create Console

```golang
console_request := codepicnic.ConsoleRequest{
  Title: "My Awesome Console",
  Type:  "golang",
}
console, err := codepicnic.CreateConsole(console_request)
if err != nil {
  fmt.Println(err.Error())
  return
}

```

### Get Consoles List 

```golang
var consoles []codepicnic.Console
consoles, err = codepicnic.ListConsoles()`
```

### Get Console object
```golang
var console codepicnic.Console
console, err = codepicnic.GetConsole("3b0e40daaad6cd0ac3ec16efa5a25762")

```

### Start, Stop, Restart a  Console 

```golang
console.Start()
console.Stop()
console.Restart()
```

### Remove a  Console 

```golang
console.Remove()
```
### Get status of a Console: running, stopped, exited

```golang
    status, err := console.Status()
    if status == "exited" {
        console.Start()
    }

```
### Get Info about a Console

```golang
fmt.Println(console.Title())
fmt.Println(console.Name())
fmt.Println(console.ContainerName())
fmt.Println(console.ContainerType())
```

