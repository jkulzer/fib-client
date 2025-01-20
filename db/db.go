package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"

	// "fyne.io/fyne/v2/app"
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"

	"github.com/jkulzer/fib-client/env"
	"github.com/jkulzer/fib-client/models"
)

func InitDB(app fyne.App) env.Env {

	dbPathUri := storage.NewFileURI(filepath.Join(app.Storage().RootURI().Path(), "sqlite.db"))

	fmt.Println("dev.jkulzer.findinberlin " + dbPathUri.Path())

	db, err := gorm.Open(sqlite.Open(dbPathUri.Path()), &gorm.Config{})
	if err != nil {
		log.Err(err).Msg("failed to create/open db")
	}

	err = db.AutoMigrate(&models.LoginInfo{})
	if err != nil {
		log.Err(err)
	}

	env := env.Env{
		DB: db,
	}

	return env
}
