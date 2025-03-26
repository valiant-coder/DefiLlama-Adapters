package db

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/queryparams"
	"exapp-go/pkg/utils"
	"log"
	"testing"
)

func TestQueryUsers(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	r := New()

	queryparams := &queryparams.QueryParams{
		Offset: 0,
		Limit:  10,
	}
	resp, count, err := r.QueryUserList(context.Background(), queryparams)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(len(resp), count)
	log.Println(count)
}

func TestGetStatisAddUserCount(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	r := New()

	data, total, err := r.GetStatisAddUserCount(context.Background(), "day", 30)
	if err != nil {
		log.Println(err)
		return
	}

	for _, v := range data {
		log.Println(v.Period, v.Count)
	}

	log.Println(total)
}
