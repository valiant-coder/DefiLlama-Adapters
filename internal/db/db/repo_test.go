package db

import (
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func TestRepo_AutoMigrate(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.toml")
	r := New()
	err := r.Debug().AutoMigrate()
	if err != nil {
		t.Fatal(err)
	}
}
