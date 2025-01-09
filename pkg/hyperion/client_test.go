package hyperion

import (
	"context"
	"testing"
)

func TestClient_GetTokens(t *testing.T) {
	c := NewClient("http://44.223.68.11:7000")
	tokens, err := c.GetTokens(context.Background(), "abcabcabcabc")
	if err != nil {
		t.Errorf("GetTokens() error = %v", err)
	}
	t.Logf("tokens: %v", tokens)
}
