// Package verifierd contains implementations for verifier API.
package verifierd

import (
	"context"

	"github.com/huyuncong/MerkleSquare/core"
	"github.com/huyuncong/MerkleSquare/grpcint"
	legolog_grpcint "github.com/huyuncong/MerkleSquare/legolog/legolog-grpcint"
)

func (v *Verifier) VerifyRegisterAsync(ctx context.Context,
	req *grpcint.VerifyRegisterRequest) (*legolog_grpcint.VerifyRegisterResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	proofRequest := legolog_grpcint.GetMasterKeyProofRequest{
		Usr: req.GetUsr(),
		Key: req.GetKey(),
		Pos: req.GetPos(),
	}
	registerRequest := KeyRegisterRequest{
		ProofRequest: &proofRequest,
		signature:    req.GetSignature(),
		vrf:          req.GetVrfKey(),
	}
	v.RegisterLock.Lock()
	v.RegisterRequests = append(v.RegisterRequests, &registerRequest)
	v.RegisterLock.Unlock()
	v.keysLock.Lock()
	v.keys[string(req.GetVrfKey())] = core.AddKeyHash(
		v.keys[string(req.GetVrfKey())], req.GetVrfKey(), req.GetKey().GetMk(),
		req.GetSignature(), uint32(req.GetPos().GetPos()))
	v.keysLock.Unlock()
	resp := new(legolog_grpcint.VerifyRegisterResponse)
	return resp, nil
}

func (v *Verifier) VerifyAppendAsync(ctx context.Context,
	req *legolog_grpcint.VerifyAppendRequest) (*legolog_grpcint.VerifyAppendResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	proofRequest := legolog_grpcint.GetPublicKeyProofRequest{
		Usr:    req.GetUsr(),
		Key:    req.GetKey(),
		Pos:    req.GetPos(),
		Height: 0,
	}
	appendRequest := KeyAppendRequest{
		ProofRequest: &proofRequest,
		signature:    req.GetSignature(),
		vrf:          req.GetVrfKey(),
		nodeHash: core.ComputeLeafNodeHash(req.GetVrfKey(),
			req.GetKey().GetEk(), req.GetSignature(), uint32(req.GetPos().GetPos())),
	}
	v.AppendLock.Lock()
	v.AppendRequests = append(v.AppendRequests, &appendRequest)
	v.AppendLock.Unlock()
	v.keysLock.Lock()
	v.keys[string(req.GetVrfKey())] = core.AddKeyHash(
		v.keys[string(req.GetVrfKey())], req.GetVrfKey(), req.GetKey().GetEk(),
		req.GetSignature(), uint32(req.GetPos().GetPos()))
	v.keysLock.Unlock()
	resp := new(legolog_grpcint.VerifyAppendResponse)
	return resp, nil
}

func (v *Verifier) VerifyLookUpAsync(ctx context.Context,
	req *legolog_grpcint.VerifyLookUpRequest) (*legolog_grpcint.VerifyLookUpResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	proofRequest := legolog_grpcint.GetLookUpProofRequest{
		IsMasterKey:   req.GetIsMasterKey(),
		Usr:           req.GetUsr(),
		MasterKey:     req.GetMasterKey(),
		EncryptionKey: req.GetEncryptionKey(),
		Pos:           req.GetPos(),
	}
	lookUpRequest := KeyLookUpRequest{
		ProofRequest: &proofRequest,
		signature:    req.GetSignature(),
		vrf:          req.GetVrf(),
	}
	v.LookupLock.Lock()
	v.LookupRequests = append(v.LookupRequests, &lookUpRequest)
	v.LookupLock.Unlock()
	resp := new(legolog_grpcint.VerifyLookUpResponse)
	return resp, nil
}
