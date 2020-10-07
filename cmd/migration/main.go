package main

import (
	"fmt"
	"log"
	"os"

	config "github.com/valensto/api_apbp/configs"
	"github.com/valensto/api_apbp/infra/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func run() error {
	conf, err := config.Load()
	if err != nil {
		return err
	}

	mongoStore := store.New(conf.DB)

	err = mongoStore.Open()
	if err != nil {
		return err
	}
	defer mongoStore.Close()

	err = mongoStore.BindBD("apbp")
	if err != nil {
		return err
	}

	if err = mongoStore.User().Migrate(); err != nil {
		return err
	}

	if err = mongoStore.Product().Migrate(); err != nil {
		return err
	}

	if err = mongoStore.Order().Migrate(); err != nil {
		return err
	}

	fmt.Println(conf.App.JWTSecret)

	return nil
}
