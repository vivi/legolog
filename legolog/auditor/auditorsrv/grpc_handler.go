// Package auditorsrv contains server side implementations for auditor API.
package auditorsrv

import (
	"context"

	legolog_grpcint "github.com/huyuncong/MerkleSquare/legolog/legolog-grpcint"
)

// GetEpochUpdate implements server-side logic for client requesting epoch
// update from the auditor. The auditor periodically queries the server and
// maintains the latest checkpoint, so when client requests the latest
// checkpoint the auditor can simply return the cached checkpoint.
func (a *Auditor) GetEpochUpdate(ctx context.Context,
	req *legolog_grpcint.GetEpochUpdateRequest) (*legolog_grpcint.GetEpochUpdateResponse, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return &legolog_grpcint.GetEpochUpdateResponse{
		CkPoints: a.UpdateCheckpoints,
	}, nil
}

// GetEpochUpdateForPartition implements server-side logic for client requesting epoch
// update from the auditor for a partition. The auditor periodically queries the server and
// maintains the latest checkpoint, so when client requests the latest
// checkpoint the auditor can simply return the cached checkpoint.
func (a *Auditor) GetEpochUpdateForPartition(ctx context.Context,
	req *legolog_grpcint.GetEpochUpdateForPartitionRequest) (*legolog_grpcint.GetEpochUpdateForPartitionResponse, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return &legolog_grpcint.GetEpochUpdateForPartitionResponse{
		CkPoint: a.UpdateCheckpoints[req.GetPartition()],
	}, nil
}
