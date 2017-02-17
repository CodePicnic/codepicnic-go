# CodePicnic library for Go

## Usage

### Initialization

```javascript

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

### Get Consoles List 

```
var consoles []codepicnic.ConsoleJson
consoles, err = codepicnic.ListConsoles()`
```

### Get Console object
```
var console codepicnic.ConsoleJson
console, err = codepicnic.GetConsole("3b0e40daaad6cd0ac3ec16efa5a25762")

```

### Start, Stop, Restart a  Console 

```
console.Start()
console.Stop()
console.Restart()
```

### Remove a  Console 

```
console.Remove()
```

