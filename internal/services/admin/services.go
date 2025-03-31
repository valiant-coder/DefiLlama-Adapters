package admin

import (
	ckhdb "exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
)

type AdminServices struct {
	repo      *db.Repo
	ckhdbRepo *ckhdb.ClickHouseRepo
}

func New() *AdminServices {
	return &AdminServices{
		repo:      db.New(),
		ckhdbRepo: ckhdb.New(),
	}
}
