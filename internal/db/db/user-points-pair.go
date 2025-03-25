package db

import "gorm.io/gorm"

func init() {
	
	addMigrateFunc(func(r *Repo) error {
		
		return r.AutoMigrate(&UserPointsPair{})
	})
}

type UserPointsPair struct {
	gorm.Model
	Pair        string `gorm:"column:pair;type:varchar(255);not null;uniqueIndex:idx_pair"`
	Coefficient uint64 `gorm:"column:coefficient;type:bigint(20);not null;default:1"`
}
