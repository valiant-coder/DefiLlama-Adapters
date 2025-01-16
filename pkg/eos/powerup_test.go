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
		10000,
		40800000000,
		200000,
	)
	fmt.Println(err)
	if err != nil {
		t.Fatal(err)
	}
}
