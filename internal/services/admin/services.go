package admin

import "exapp-go/internal/db/db"

type AdminServices struct {
	repo *db.Repo
}

func New() *AdminServices {
	return &AdminServices{
		repo: db.New(),
	}
}
