package onedex

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"testing"
)

func TestSignupClient_GetAccApplies(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_testnet2.yaml")
	client := NewSignupClient(config.Conf().Eos.NodeURL, config.Conf().Eos.OneDex.SignUpContract, config.Conf().Eos.OneDex.Actor, config.Conf().Eos.OneDex.ActorPrivateKey, config.Conf().Eos.OneDex.ActorPermission)
	pubkey, err := client.GetPubkeyByUID(context.Background(), "67998410")
	if err != nil {
		t.Fatalf("GetPubkeyByUID failed: %v", err)
	}
	t.Logf("GetPubkeyByUID response: %v", pubkey)
}
