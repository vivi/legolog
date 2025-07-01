// Package auditorclt contains client functions to make requests to
// auditor service.
package auditorclt

import (
	"context"
	"encoding/json"

	"github.com/huyuncong/MerkleSquare/core"
	legolog_grpcint "github.com/huyuncong/MerkleSquare/legolog/legolog-grpcint"
	"google.golang.org/grpc"
)

// Client is a client library for auditor operations.
type Client interface {
	// GetEpochUpdate fetches latest verified checkpoint from the auditor.
	//GetEpochUpdate(ctx context.Context) (*legolog_grpcint.GetEpochUpdateResponse, error)
	GetEpochUpdate(ctx context.Context) ([]*core.LegologDigest, []uint64, []uint64, error)
	GetEpochUpdateForPartition(ctx context.Context, partition uint64) (*core.LegologDigest, uint64, uint64, error)
}

// assert that auditorClient implements auditorclt.Client interfact
var _ Client = (*auditorClient)(nil)

// auditorClient is an implementation of auditorclt.Client
type auditorClient struct {
	client legolog_grpcint.AuditorClient
}

// NewAuditorClient creates and returns a connection to the auditor.
func NewAuditorClient(address string) (Client, error) {
	auditorConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	grpcClient := legolog_grpcint.NewAuditorClient(auditorConn)
	return &auditorClient{client: grpcClient}, nil
}

// GetEpochUpdate fetches latest verified checkpoint from the auditor.
func (a *auditorClient) GetEpochUpdate(ctx context.Context) (digests []*core.LegologDigest, numLeaves []uint64, epochs []uint64, err error) {
	resp, err := a.client.GetEpochUpdate(ctx, &legolog_grpcint.GetEpochUpdateRequest{})
	if err != nil {
		return
	}
	cps := resp.GetCkPoints()
	for _, cp := range cps {
		var digest core.LegologDigest
		json.Unmarshal(cp.GetMarshaledDigest(), &digest)
		digests = append(digests, &digest)
		numLeaves = append(numLeaves, cp.GetNumLeaves())
		epochs = append(epochs, cp.GetEpoch())
	}
	return
}

// GetEpochUpdateForPartition fetches latest verified checkpoint from the auditor for a given partition.
func (a *auditorClient) GetEpochUpdateForPartition(ctx context.Context, partition uint64) (*core.LegologDigest, uint64, uint64, error) {
	resp, err := a.client.GetEpochUpdateForPartition(ctx, &legolog_grpcint.GetEpochUpdateForPartitionRequest{Partition: partition})
	if err != nil {
		return nil, 0, 0, err
	}
	cp := resp.GetCkPoint()
	var digest core.LegologDigest
	json.Unmarshal(cp.GetMarshaledDigest(), &digest)
	numLeaves := cp.GetNumLeaves()
	epoch := cp.GetEpoch()
	return &digest, numLeaves, epoch, nil
}
