package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/huyuncong/MerkleSquare/constants"

	"github.com/immesys/bw2/crypto"
)

const ServerAddr = "localhost" + constants.ServerPort

const AuditorAddr = "localhost" + constants.AuditorPort
const VerifierAddr = "localhost" + constants.VerifierPort

func TestClient(t *testing.T) {
	ctx := context.Background()

	c, err := NewClient(ServerAddr, AuditorAddr, VerifierAddr)
	if err != nil {
		t.Error(errors.New("Failed to start client: " + err.Error()))
	}

	aliceUsername := []byte("alice")
	masterSK, masterVK := crypto.GenerateKeypair()
	_, aliceVK1 := crypto.GenerateKeypair()

	_, err = c.Register(ctx, aliceUsername, masterSK, masterVK)
	if err != nil {
		t.Error(errors.New("Failed to register first user: " + err.Error()))
	}
	pos, _, err := c.Append(ctx, aliceUsername, aliceVK1)
	if err != nil {
		t.Error(errors.New("Failed to append alice's PK: " + err.Error()))
	}
	if pos != 1 {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}

	time.Sleep(time.Second * 2)

	key, pos, err := c.LookUpMK(ctx, aliceUsername)
	if err != nil {
		t.Error(errors.New("Failed to look up alice's MK: " + err.Error()))
	}
	if !reflect.DeepEqual(key, masterVK) {
		t.Error(errors.New("Master key mismatch"))
	}
	if pos != 0 {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}

	key, pos, err = c.LookUpPK(ctx, aliceUsername)
	if err != nil {
		t.Error(errors.New("Failed to look up alice's PK: " + err.Error()))
	}
	if !reflect.DeepEqual(key, aliceVK1) {
		t.Error(errors.New("Public key mismatch"))
	}
	if pos != 1 {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}

	bobSK, bobVK := crypto.GenerateKeypair()
	_, err = c.Register(ctx, []byte("bob"), bobSK, bobVK)
	if err != nil {
		t.Error(errors.New("Failed to register second user: " + err.Error()))
	}
}
