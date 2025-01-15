package env

import (
	"gorm.io/gorm"
)

type Env struct {
	DB  *gorm.DB
	Url string
}
