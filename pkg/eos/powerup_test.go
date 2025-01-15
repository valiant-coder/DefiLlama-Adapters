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
	)
	fmt.Println(err)
	if err != nil {
		t.Fatal(err)
	}
}
