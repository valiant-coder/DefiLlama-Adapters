package hyperion

import (
	"testing"
)

func TestClient_GetTokens(t *testing.T) {
	c := NewClient("http://44.223.68.11:7000")
	txID, err := c.GetEvmTxIDByEosTxID("d1f32a205ed5ca197af2771de3961a5d38165d296554609909e9da6dbbf711d4")
	if err != nil {
		t.Errorf("GetEvmTxIDByEosTxID() error = %v", err)
	}
	t.Logf("txID: %v", txID)

}
