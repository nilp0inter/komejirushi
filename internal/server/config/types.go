package config

import (
	"database/sql"
)

type Config struct {
	Docsets map[string]*sql.DB
}
