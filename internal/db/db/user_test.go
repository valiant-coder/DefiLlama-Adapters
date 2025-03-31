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

	data, total, err := r.GetStatisAddUserCount(context.Background(), "month", 5)
	if err != nil {
		log.Println(err)
		return
	}

	for _, v := range data {
		log.Println(v.Period, v.Count)
	}

	log.Println(total)
}

func TestGetStatisAddPasskeyCount(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	r := New()

	data, total, err := r.GetStatisAddPasskeyCount(context.Background(), "month", 5)
	if err != nil {
		log.Println(err)
		return
	}

	for _, v := range data {
		log.Println(v.Period, v.Count)
	}
	log.Println(total)
}

func TestGetStatisAddEvmCount(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	r := New()

	data, total, err := r.GetStatisAddEvmCount(context.Background(), "month", 5)
	if err != nil {
		log.Println(err)
		return
	}

	for _, v := range data {
		log.Println(v.Period, v.Count)
	}
	log.Println(total)
}

func TestGetStatisAddDepositCount(t *testing.T) {

	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	r := New()

	data, total, err := r.GetStatisAddDepositCount(context.Background(), "month", 5)
	if err != nil {
		log.Println(err)
		return
	}

	for _, v := range data {
		log.Println(v.Period, v.Count)
	}
	log.Println(total)
}

func TestGetUserTotalBalanceByIsEvmUser(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	r := New()

	data, err := r.GetUserTotalBalanceByIsEvmUser(context.Background(), false)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(data)

	log.Println("success1")
}

func TestGetUserCoinTotalBalanceByIsEvmUser(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config.yaml")
	r := New()

	data, err := r.GetUserCoinTotalBalanceByIsEvmUser(context.Background(), true)
	if err != nil {
		log.Println(err)
		return
	}

	for _, v := range data {
		log.Println(v.Coin, v.Amount)
	}

	log.Println("success")
}
