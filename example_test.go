package args_test

import (
	"fmt"
	"os"

	"github.com/kosadoge/args"
)

func Example() {
	// $ app -p 8888 --debug true
	os.Args = []string{"app", "-p", "8888", "--debug", "true"}

	fs := args.New()
	var (
		port  = fs.String("port,p", "9999", "listen port")
		debug = fs.Bool("debug", false, "enable debug mode")
	)
	fs.Parse(os.Args[1:])

	fmt.Println("port:", *port)
	fmt.Println("debug:", *debug)
	// Output:
	// port: 8888
	// debug: true
}

func Example_withEnvironments() {
	// $ PORT=8888 DEBUG=true app
	os.Args = []string{"app"}
	os.Setenv("PORT", "8888")
	os.Setenv("DEBUG", "true")

	fs := args.New()
	var (
		port  = fs.String("port,p", "9999", "listen port")
		debug = fs.Bool("debug", false, "enable debug mode")
	)
	fs.Parse(os.Args[1:], args.Env())

	fmt.Println("port:", *port)
	fmt.Println("debug:", *debug)
	// Output:
	// port: 8888
	// debug: true
}

func Example_withConfigFile() {
	// $ cat ./testdata/json_config.json
	// {
	// 	"port": "8888",
	// 	"debug": true,
	// 	"log_level": "info"
	// }
	// $ app --config ./testdata/json_config.json
	os.Args = []string{"app", "--config", "./testdata/json_config.json"}

	fs := args.New()
	var (
		port   = fs.String("port,p", "9999", "listen port")
		debug  = fs.Bool("debug", false, "enable debug mode")
		config = fs.String("config", "", "config file")
	)
	fs.Parse(os.Args[1:], args.Json(config))

	fmt.Println("port:", *port)
	fmt.Println("debug:", *debug)
	// Output:
	// port: 8888
	// debug: true
}
