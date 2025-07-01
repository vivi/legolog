// Package merkleclt contains client functions to make requests to
// merkle service.
package legolog

import (
	"context"
	"strconv"

	legolog_grpcint "github.com/huyuncong/MerkleSquare/legolog/legolog-grpcint"
	"github.com/immesys/bw2/crypto"
	"google.golang.org/grpc"
)

// Client is a client library for verifier operations.
type BasicClient interface {
	Register(ctx context.Context, req *legolog_grpcint.RegisterRequest) (
		*legolog_grpcint.RegisterResponse, error)
	Append(ctx context.Context, req *legolog_grpcint.AppendRequest,
		masterSK, masterVK /*, key*/ []byte) (*legolog_grpcint.AppendResponse, []byte, error)
	LookUpMK(ctx context.Context, req *legolog_grpcint.LookUpMKRequest) (
		*legolog_grpcint.LookUpMKResponse, error)
	LookUpPK(ctx context.Context, req *legolog_grpcint.LookUpPKRequest) (
		*legolog_grpcint.LookUpPKResponse, error)

	// the following two functions currently have emtpy implementations
	LookUpMKVerify(ctx context.Context, req *legolog_grpcint.LookUpMKVerifyRequest) (
		*legolog_grpcint.LookUpMKVerifyResponse, error)
	LookUpPKVerify(ctx context.Context, req *legolog_grpcint.LookUpPKVerifyRequest) (
		*legolog_grpcint.LookUpPKVerifyResponse, error)

	GetNewCheckPoint(ctx context.Context, req *legolog_grpcint.GetNewCheckPointRequest) (
		*legolog_grpcint.GetNewCheckPointResponse, error)

	GetNewUpdateCheckPoint(ctx context.Context, req *legolog_grpcint.GetNewCheckPointRequest) (
		*legolog_grpcint.GetNewCheckPointResponse, error)
	GetNewVerifyCheckPoint(ctx context.Context, req *legolog_grpcint.GetNewCheckPointRequest) (
		*legolog_grpcint.GetNewCheckPointResponse, error)
	/*
		GetMasterKeyProof(ctx context.Context, req *legolog_grpcint.GetMasterKeyProofRequest) (
			*legolog_grpcint.GetMasterKeyProofResponse, error)
		GetPublicKeyProof(ctx context.Context,
			req *legolog_grpcint.GetPublicKeyProofRequest) (*legolog_grpcint.GetPublicKeyProofResponse, error)
		GetLookUpProof(ctx context.Context, req *legolog_grpcint.GetLookUpProofRequest) (
			*legolog_grpcint.GetLookUpProofResponse, error)
		GetMonitoringProofForTest(ctx context.Context, req *legolog_grpcint.GetMonitoringProofForTestRequest) (
			*legolog_grpcint.GetMonitoringProofForTestResponse, error)
	*/
}

// assert that legologClient implements BasicClient interface
var _ BasicClient = (*legologClient)(nil)

// legologClient is an implementation of merkleclt.Client
type legologClient struct {
	client legolog_grpcint.LegoLogClient
}

// NewlegologClient creates and returns a connection to the merkleSquare server.
func NewLegologClient(address string) (BasicClient, error) {
	legologConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	grpcClient := legolog_grpcint.NewLegoLogClient(legologConn)
	return &legologClient{client: grpcClient}, nil
}

// NewlegologClientWithMaxMsgSize creates and returns a connection to the merkleSquare server.
// func NewlegologClientWithMaxMsgSize(address string, maxMsgSize int) (Client, error) {
// 	merkleConn, err := grpc.Dial(address,
// 		grpc.WithDefaultCallOptions(
// 			grpc.MaxCallSendMsgSize(maxMsgSize),
// 			grpc.MaxCallRecvMsgSize(maxMsgSize)),
// 		grpc.WithInsecure(),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	grpcClient := legolog_grpcint.NewLegoLogClient(merkleConn)
// 	return &legologClient{client: grpcClient}, nil
// }

func (m *legologClient) Register(ctx context.Context, req *legolog_grpcint.RegisterRequest) (
	*legolog_grpcint.RegisterResponse, error) {
	return m.client.Register(ctx, req)
}

func (m *legologClient) Append(ctx context.Context, req *legolog_grpcint.AppendRequest,
	masterSK, masterVK /*, key*/ []byte) (*legolog_grpcint.AppendResponse, []byte, error) {
	stream, err := m.client.Append(ctx)
	if err != nil {
		return nil, nil, err
	}

	stream.Send(req)

	response, err := stream.Recv()
	if err != nil {
		return nil, nil, err
	}

	signature := make([]byte, 64)
	crypto.SignBlob(masterSK, masterVK, signature,
		append(req.Value.Value, []byte(strconv.Itoa(int(response.GetPos().GetPos())))...))
	req.Signature = signature
	// fmt.Printf("sig: %x\n", signature)
	stream.Send(req)

	response, err = stream.Recv()
	if err != nil {
		return nil, nil, err
	}

	stream.CloseSend()

	return response, signature, nil
}

func (m *legologClient) LookUpMK(ctx context.Context,
	req *legolog_grpcint.LookUpMKRequest) (
	*legolog_grpcint.LookUpMKResponse, error) {
	return m.client.LookUpMK(ctx, req)
}

func (m *legologClient) LookUpPK(ctx context.Context,
	req *legolog_grpcint.LookUpPKRequest) (
	*legolog_grpcint.LookUpPKResponse, error) {
	return m.client.LookUpPK(ctx, req)
}

func (m *legologClient) LookUpMKVerify(ctx context.Context,
	req *legolog_grpcint.LookUpMKVerifyRequest) (
	*legolog_grpcint.LookUpMKVerifyResponse, error) {
	return m.client.LookUpMKVerify(ctx, req)
}

func (m *legologClient) LookUpPKVerify(ctx context.Context,
	req *legolog_grpcint.LookUpPKVerifyRequest) (
	*legolog_grpcint.LookUpPKVerifyResponse, error) {
	return m.client.LookUpPKVerify(ctx, req)
}

func (m *legologClient) GetNewCheckPoint(ctx context.Context,
	req *legolog_grpcint.GetNewCheckPointRequest) (
	*legolog_grpcint.GetNewCheckPointResponse, error) {
	return m.client.GetNewCheckPoint(ctx, req)
}

func (m *legologClient) GetNewUpdateCheckPoint(ctx context.Context,
	req *legolog_grpcint.GetNewCheckPointRequest) (
	*legolog_grpcint.GetNewCheckPointResponse, error) {
	return m.client.GetNewUpdateCheckPoint(ctx, req)
}

func (m *legologClient) GetNewVerifyCheckPoint(ctx context.Context,
	req *legolog_grpcint.GetNewCheckPointRequest) (
	*legolog_grpcint.GetNewCheckPointResponse, error) {
	return m.client.GetNewVerifyCheckPoint(ctx, req)
}

// func (m *legologClient) GetMasterKeyProof(ctx context.Context,
// 	req *legolog_grpcint.GetMasterKeyProofRequest) (
// 	*legolog_grpcint.GetMasterKeyProofResponse, error) {
// 	return m.client.GetMasterKeyProof(ctx, req)
// }

func (m *legologClient) GetPublicKeyProof(ctx context.Context,
	req *legolog_grpcint.GetPublicKeyProofRequest) (
	*legolog_grpcint.GetPublicKeyProofResponse, error) {
	return m.client.GetPublicKeyProof(ctx, req)
}

// func (m *legologClient) GetLookUpProof(ctx context.Context,
// 	req *legolog_grpcint.GetLookUpProofRequest) (
// 	*legolog_grpcint.GetLookUpProofResponse, error) {
// 	return m.client.GetLookUpProof(ctx, req)
// }

// func (m *legologClient) GetMonitoringProofForTest(ctx context.Context,
// 	req *legolog_grpcint.GetMonitoringProofForTestRequest) (
// 	*legolog_grpcint.GetMonitoringProofForTestResponse, error) {
// 	return m.client.GetMonitoringProofForTest(ctx, req)
// }
