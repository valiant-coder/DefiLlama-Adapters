package eos

import (
	"context"
	"fmt"
	"testing"
)

func TestPowerUp(t *testing.T) {
	err := PowerUp(context.Background(),
		"http://44.223.68.11:8888",
		"",
		"",
		1000000000,
		100000,
	)
	fmt.Println(err)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetPowerUpStatusWeight(t *testing.T) {
	net, cpu, err := GetPowerUpStatusWeight(context.Background(), "http://44.223.68.11:8888")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(net, cpu)
}
