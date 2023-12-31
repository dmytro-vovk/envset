# EnvSet

Tiny package to populate your structs from environment variables.

## Usage

Install it:
```shell
go get github.com/dmytro-vovk/envset
```
Import it:
```go
import "github.com/dmytro-vovk/envset"
```
Use it:

1. Define your structure, e.g:
```go
type MyConfig struct {
	Webserver struct {
		Listen string `env:"LISTEN" default:":8080"`
	}
	Logfile string `env:"LOG_FILE"`
}
```
2. Create a variable of your type:
```go
var config &MyConfig 
```
3. Call the package to populate the variable:
```go
if err := envset.Set(&config); err != nil {
	log.Fatalf("Error setting config from environment: %s", err)
}
```
4. Set environment variables and run your app:
```shell
LISTEN="localhost:8088" LOG_FILE="/tmp/my_app.log" go run main.go
```
