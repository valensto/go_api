package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/labstack/gommon/color"
	config "github.com/valensto/api_apbp/configs"
	"github.com/valensto/api_apbp/internal/api"
	"github.com/valensto/api_apbp/internal/store"
	"github.com/valensto/api_apbp/pkg/mailer"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	conf, err := config.Load()
	if err != nil {
		return err
	}

	fmt.Println("loading...")

	srv, err := api.NewServer(conf.App)
	if err != nil {
		return fmt.Errorf("error during server initialisation. got=%w", err)
	}

	mongoStore := store.New(conf.DB)
	srv.Store = &mongoStore

	mailer := mailer.NewMailer(conf.Mailer)
	srv.Mailer = mailer

	err = srv.InitStructValidator()
	if err != nil {
		return fmt.Errorf("error during struct validator initialisation. got=%w", err)
	}

	err = srv.Store.Open()
	if err != nil {
		return fmt.Errorf("error during opening store. got=%w", err)
	}
	defer srv.Store.Close()

	err = srv.Store.BindBD("apbp")
	if err != nil {
		return fmt.Errorf("error during binding database. got=%w", err)
	}

	Banner()

	http.ListenAndServe(":8000", srv.Router)

	return nil
}

// Banner Print banner
func Banner() {
	b := `<blue>
   
██╗   ██╗ █████╗ ███████╗██╗   ██╗
╚██╗ ██╔╝██╔══██╗██╔════╝██║   ██║
 ╚████╔╝ ███████║███████╗██║   ██║
  ╚██╔╝  ██╔══██║╚════██║██║   ██║
   ██║   ██║  ██║███████║╚██████╔╝
   ╚═╝   ╚═╝  ╚═╝╚══════╝ ╚═════╝ 
</>
   <yellow>https://www.yasu.io - %v ©</>
			
`
	t := time.Now()
	y := t.Year()
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	c.Run()
	color.Printf(b, y)
}
