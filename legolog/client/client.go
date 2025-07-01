package legolog

import (
	"context"
	"errors"

	"github.com/golang/protobuf/proto"
	legolog_grpcint "github.com/huyuncong/MerkleSquare/legolog/legolog-grpcint"

	"github.com/immesys/bw2/crypto"
)

// Client represents a connection to the server and the verification daemon,
// over which operation and verification can be performed. It contains references
// to multiple possible clinet types, including a basic read-and-write client, auditor
// client, and verifier client.
type Client struct {
	legologClient  BasicClient
	auditorClient  *struct{} //auditorclt.Client
	verifierClient *struct{} // verifierclt.Client

	masterKeys map[string]MasterKeyRecord
}

type MasterKeyRecord struct {
	masterSK []byte
	masterVK []byte
}

// NewClient connects to the server and the daemon, and returns a Client
// representing that connection.
func NewClient(serverAddr string, auditorAddr string, verifierAddr string) (
	*Client, error) {

	var err error
	c := Client{
		masterKeys: make(map[string]MasterKeyRecord),
	}
	c.legologClient, err = NewLegologClient(serverAddr)
	if err != nil {
		return nil, err
	}
	// if auditorAddr != "" {
	// 	c.auditorClient, err = auditorclt.NewAuditorClient(auditorAddr)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	// if verifierAddr != "" {
	// 	c.verifierClient, err = verifierclt.NewVerifierClient(verifierAddr)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return &c, nil
}

/*
func NewClientForUserThroughput(serverAddr string, maxMsgSize int) (
	*Client, error) {

	c := Client{
		masterKeys: make(map[string]MasterKeyRecord),
	}
	var err error
	c.legologClient, err = merkleclt.NewMerkleClientWithMaxMsgSize(serverAddr, maxMsgSize)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
*/

// Register user and the corresponding public master key to the server.
func (c *Client) Register(ctx context.Context, username []byte,
	masterSK []byte, masterVK []byte) (uint64, error) {
	response, err := c.registerInt(ctx, username, masterSK, masterVK)
	return response.GetPos().GetPos(), err
}

func (c *Client) RegisterForSize(ctx context.Context, username []byte,
	masterSK []byte, masterVK []byte) (int, error) {
	response, err := c.registerInt(ctx, username, masterSK, masterVK)
	return proto.Size(response), err
}

func (c *Client) RegisterForThroughput(ctx context.Context, username []byte,
	masterSK []byte, masterVK []byte) (uint64, error) {
	signature := make([]byte, 64)
	crypto.SignBlob(masterSK, masterVK, signature, masterVK)
	request := legolog_grpcint.RegisterRequest{
		Usr:       &legolog_grpcint.Username{Username: username},
		Key:       &legolog_grpcint.MasterKey{Mk: masterVK},
		Signature: signature,
	}
	response, err := c.legologClient.Register(ctx, &request)
	return response.GetPos().GetPos(), err
}

// registerInt is the internal implementation to register a user
func (c *Client) registerInt(ctx context.Context, username []byte,
	masterSK []byte, masterVK []byte) (
	*legolog_grpcint.RegisterResponse, error) {
	c.masterKeys[string(username)] = MasterKeyRecord{
		masterSK: masterSK,
		masterVK: masterVK,
	}

	signature := make([]byte, 64)
	crypto.SignBlob(masterSK, masterVK, signature, masterVK)

	request := &legolog_grpcint.RegisterRequest{
		Usr:       &legolog_grpcint.Username{Username: username},
		Key:       &legolog_grpcint.MasterKey{Mk: masterVK},
		Signature: signature,
	}
	response, err := c.legologClient.Register(ctx, request)
	if err != nil {
		return nil, err
	}

	var verifierErr error
	/*
		if c.verifierClient != nil {
			verifierErr = c.verifierClient.VerifyRegisterAsync(ctx, request, response)
		}
	*/

	return response, verifierErr
}

// Append adds a public key for the user to the key transparency infrastructure.
func (c *Client) Append(ctx context.Context, username []byte, identifier []byte, key []byte) (
	uint64, *legolog_grpcint.VerifyAppendRequest, error) {
	response /*verifierReq,*/, err := c.appendInt(ctx, username, identifier, key)
	return response.GetPos().GetPos(), nil /*verifierReq*/, err
}

// AppendForSize adds a public key for the user to the key transparency infrastructure
// and returns the size of response returned by the server.
func (c *Client) AppendForSize(ctx context.Context, username []byte, identifier []byte, key []byte) (int, error) {
	response /*_,*/, err := c.appendInt(ctx, username, identifier, key)
	return proto.Size(response), err
}

// appendInt is the internal implementation to add a public key for the user.
func (c *Client) appendInt(ctx context.Context, username []byte, identifier []byte, value []byte) (
	*legolog_grpcint.AppendResponse /* *legolog_grpcint.VerifyAppendRequest,*/, error) {
	masterKeyInfo, ok := c.masterKeys[string(username)]
	if !ok {
		return nil /*, nil*/, errors.New("masterkey does not exist")
	}

	request := &legolog_grpcint.AppendRequest{
		Usr:        &legolog_grpcint.Username{Username: username},
		Identifier: &legolog_grpcint.Identifier{Identifier: identifier},
		Value:      &legolog_grpcint.Value{Value: value},
	}
	response, _ /*signature*/, err := c.legologClient.Append(ctx, request,
		masterKeyInfo.masterSK, masterKeyInfo.masterVK)
	if err != nil {
		return nil /* nil,*/, err
	}

	// var verifierErr error
	// var verifierRequest *legolog_grpcint.VerifyAppendRequest
	// if c.verifierClient != nil {
	// 	verifierRequest = &legolog_grpcint.VerifyAppendRequest{
	// 		Usr:       request.GetUsr(),
	// 		VrfKey:    response.GetVrfKey(),
	// 		Key:       request.GetEk(),
	// 		Signature: signature,
	// 		Pos:       response.GetPos(),
	// 	}
	// 	verifierErr = c.verifierClient.VerifyAppendAsync(ctx, verifierRequest)
	// }
	return response, nil //verifierRequest, verifierErr
}

// LookUpMK takes a name and returns the associated master key of the user, if user exists.
// Verification of the response is done asynchronously by the verifier daemon.
func (c *Client) LookUpMK(ctx context.Context, username []byte) ([]byte, uint64, error) {
	var request = &legolog_grpcint.LookUpMKRequest{
		Usr: &legolog_grpcint.Username{Username: username},
	}
	response, err := c.legologClient.LookUpMK(ctx, request)
	if err != nil {
		return response.GetImk().GetMasterKey().GetMk(), response.GetImk().GetPos().GetPos(), err
	}

	var verifierErr error = nil
	/*
		if c.verifierClient != nil {
			verifierErr = c.verifierClient.VerifyLookUpMKAsync(ctx, request, response)
		}
	*/
	return response.GetImk().GetMasterKey().GetMk(), response.GetImk().GetPos().GetPos(), verifierErr
}

// LookUpMKVerify takes a name and looks up the associated key/proof.
// Verification of the response is done synchronously during the API call.
func (c *Client) LookUpMKVerify(ctx context.Context, username []byte) ([]byte, uint64, []byte, []byte, error) {
	response, err := c.lookUpMKVerifyInt(ctx, username)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	return response.IndexedValue.Value.Value,
		response.IndexedValue.Pos.Pos, response.Signature, response.Proof, nil
}

// LookUpMKVerifyForTest takes a name and looks up the associated key/proof.
// Verification of the response is done synchronously during the API call.
// This function returns more information relevant to testing compared to
// LookUpMKVerify.
func (c *Client) LookUpMKVerifyForTest(ctx context.Context, username []byte) (
	[]byte, uint64 /*[]byte, */, []byte, error) {
	response, err := c.lookUpMKVerifyInt(ctx, username)
	if err != nil {
		return nil, 0, nil, err
	}
	return response.IndexedValue.Value.Value,
		response.IndexedValue.Pos.Pos /*response.GetVrfKey() ,*/, response.Signature, nil
}

// LookUpMKVerifyForSize takes a name and looks up the associated key/proof.
// Verification of the response is done synchronously during the API call.
// This function returns the size of the response returned by the server.
func (c *Client) LookUpMKVerifyForSize(ctx context.Context, username []byte) (int, error) {
	response, err := c.lookUpMKVerifyInt(ctx, username)
	if err != nil {
		return 0, err
	}
	return proto.Size(response), nil
}

// LookUpMKVerifyForThroughput takes a name and looks up the associated key/proof.
// This function skips the actual verification as to only measure the time
// taken for the server to generate the proof, not including the time it takes
// to verify the proof.
/*
func (c *Client) LookUpMKVerifyForThroughput(ctx context.Context, username []byte) ([]byte, uint64, error) {
	var serverRequest = &legolog_grpcint.LookUpMKVerifyRequest{
		Size: 0,
		Usr:  &legolog_grpcint.Username{Username: username},
	}
	response, err := c.legologClient.LookUpMKVerify(ctx, serverRequest)
	return response.GetImk().GetMasterKey().GetMk(), response.GetImk().GetPos().GetPos(), err
}
*/

// lookUpMKVerifyInt is the internal function to look up the master key
// and verify the response synchronously.
func (c *Client) lookUpMKVerifyInt(ctx context.Context, username []byte) (
	*legolog_grpcint.LookUpMKVerifyResponse, error) {
	// var numLeaves uint64
	// var digest = &core.Digest{}

	if c.auditorClient != nil {
		/*
			auditorResponse, err := c.auditorClient.GetEpochUpdate(ctx)
			if err != nil {
				return nil, err
			}
			numLeaves = auditorResponse.CkPoint.GetNumLeaves()
			marshaledDigest := auditorResponse.GetCkPoint().GetMarshaledDigest()
			json.Unmarshal(marshaledDigest, &digest)
		*/
	} else {
		// numLeaves = 0
	}

	var request = &legolog_grpcint.LookUpMKVerifyRequest{
		// Size: numLeaves,
		Usr: &legolog_grpcint.Username{Username: username},
	}
	response, err := c.legologClient.LookUpMKVerify(ctx, request)
	if err != nil {
		return nil, err
	}

	// masterKey := response.IndexedValue.Value.Value
	// position := response.IndexedValue.Pos.Pos

	/*
		if c.auditorClient != nil {
			var mkProof core.MKProof
			json.Unmarshal(response.GetProof(), &mkProof)
			proven := core.VerifyMKProof(digest,
			masterKey, response.Signature, uint32(position), &mkProof) TODO:
			if !proven {
				fmt.Printf("WARNING: could not prove lookup at position %v \n", position)
				return nil, errors.New("WARNING: could not prove lookup")
			}
		}
	*/
	return response, nil
}

// LookUpPK takes a name and returns the latest published PK up until the last epoch, if one exists.
// This function will not return the master key if master key is the latest PK.
func (c *Client) LookUpPK(ctx context.Context, identifier []byte) ([]byte, uint64, error) {
	var serverRequest *legolog_grpcint.LookUpPKRequest

	if c.auditorClient != nil {
		/*
			auditorResponse, err := c.auditorClient.GetEpochUpdate(ctx)
			if err != nil {
				return nil, 0, err
			}
			serverRequest = &legolog_grpcint.LookUpPKRequest{
				// Pos: &legolog_grpcint.Position{
				// 	Pos: auditorResponse.CkPoint.GetNumLeaves(),
				// },
				Identifier: &legolog_grpcint.Identifier{Identifier: identifier},
				Usr:        &legolog_grpcint.Username{Username: username},
			}
		*/
	} else {
		serverRequest = &legolog_grpcint.LookUpPKRequest{
			// Pos: nil,
			Identifier: &legolog_grpcint.Identifier{Identifier: identifier},
		}
	}

	serverResponse, err := c.legologClient.LookUpPK(ctx, serverRequest)

	if err != nil {
		// TODO: fix pos
		return serverResponse.IndexedValue.Value.Value, serverResponse.IndexedValue.Pos.Pos, err
	}

	var verifierErr error
	/*
		if c.verifierClient != nil {
			verifierErr = c.verifierClient.VerifyLookUpPKAsync(ctx, serverRequest, serverResponse)
		}
	*/

	return serverResponse.IndexedValue.Value.Value,
		serverResponse.IndexedValue.Pos.Pos, verifierErr // TODO: pos is incorrect. We need to implement using positions first.
}

// LookUpPKVerify takes a name and looks up the associated key/proof.
// Verification of the response is done synchronously during the API call.
func (c *Client) LookUpPKVerify(ctx context.Context, username []byte, identifier []byte) ([]byte, uint64, []byte, []byte, error) {
	response, err := c.lookUpPKVerifyInt(ctx, username, identifier)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	return response.IndexedValue.Value.Value, response.IndexedValue.Pos.Pos, response.Signature, response.Proof, nil
}

// LookUpPKVerifyForTest takes a name and looks up the associated key/proof.
// Verification of the response is done synchronously during the API call.
// This function returns more information relevant to testing compared to
// LookUpPKVerify.
func (c *Client) LookUpPKVerifyForTest(ctx context.Context, username []byte, identifer []byte) ([]byte, uint64, []byte /*[]byte,*/, error) {
	response, err := c.lookUpPKVerifyInt(ctx, username, identifer)
	if err != nil {
		return nil, 0, nil, err
	}
	return response.IndexedValue.Value.Value, response.IndexedValue.Pos.Pos,
		/*response.GetVrfKey() ,*/ response.Signature, nil
}

// LookUpPKVerifyForSize takes a name and looks up the associated key/proof.
// Verification of the response is done synchronously during the API call.
// This function returns the size of the response returned by the server.
func (c *Client) LookUpPKVerifyForSize(ctx context.Context, username []byte, identifer []byte) (int, error) {
	response, err := c.lookUpPKVerifyInt(ctx, username, identifer)
	if err != nil {
		return 0, err
	}
	return proto.Size(response), nil
}

// lookUpPKVerifyInt is the internal function to look up the latest public key
// for the user and verify the response synchronously.
func (c *Client) lookUpPKVerifyInt(ctx context.Context, username []byte, identifer []byte) (
	*legolog_grpcint.LookUpPKVerifyResponse, error) {
	// var numLeaves uint64
	// var digest = new(core.Digest)

	/*
		if c.auditorClient != nil {
			auditorResponse, err := c.auditorClient.GetEpochUpdate(ctx)
			if err != nil {
				return nil, err
			}
			numLeaves = auditorResponse.CkPoint.GetNumLeaves()
			marshaledDigest := auditorResponse.GetCkPoint().GetMarshaledDigest()
			json.Unmarshal(marshaledDigest, &digest)
		} else {
			numLeaves = 0
		}
	*/

	serverRequest := &legolog_grpcint.LookUpPKVerifyRequest{
		// Size: numLeaves,
		Identifier: &legolog_grpcint.Identifier{Identifier: identifer},
	}
	serverResponse, err := c.legologClient.LookUpPKVerify(ctx, serverRequest)
	if err != nil {
		return serverResponse, err
	}

	// encryptionKey := serverResponse.IndexedValue.Value.Value
	// position := serverResponse.IndexedValue.Pos.Pos

	/*
		if c.auditorClient != nil {
			var pkProof core.LatestPKProof
			json.Unmarshal(serverResponse.GetProof(), &pkProof)
			proven := core.VerifyPKProof(digest, serverResponse.GetVrfKey(), encryptionKey, serverResponse.Signature, uint32(position), &pkProof)
			if !proven {
				fmt.Printf("WARNING: could not prove lookup at position %v \n", position)
				return serverResponse, errors.New("WARNING: could not prove lookup")
			}
		}
	*/
	return serverResponse, nil
}

// // LookUpPKVerifyForThroughput takes a name and looks up the associated key/proof.
// // This function skips the actual verification as to only measure the time
// // taken for the server to generate the proof, not including the time it takes
// // to verify the proof.
// func (c *Client) LookUpPKVerifyForThroughput(ctx context.Context,
// 	username []byte) ([]byte, uint64, error) {
// 	serverRequest := &legolog_grpcint.LookUpPKVerifyRequest{
// 		Size: 0,
// 		Usr:  &legolog_grpcint.Username{Username: username},
// 	}
// 	_, err := c.legologClient.LookUpPKVerify(ctx, serverRequest)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return nil, 0, err
// 	}
// 	return nil, 0, nil
// }

// func (c *Client) MonitoringForLatency(ctx context.Context, username []byte,
// 	pos int, height int, vrf []byte, Ek []byte, signature []byte,
// 	keyhash []core.KeyHash) error {
// 	_, err := c.monitoringInt(ctx, username, pos, height, vrf,
// 		Ek, signature, keyhash)
// 	return err
// }

// func (c *Client) MonitoringForSize(ctx context.Context, username []byte,
// 	pos int, height int, vrf []byte, Ek []byte, signature []byte,
// 	keyhash []core.KeyHash) (int, error) {
// 	response, err := c.monitoringInt(ctx, username, pos, height, vrf,
// 		Ek, signature, keyhash)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return proto.Size(response), nil
// }

// func (c *Client) MonitoringForThroughput(ctx context.Context, username []byte,
// 	pos int, height int) error {
// 	proofrequest := &legolog_grpcint.GetMonitoringProofForTestRequest{
// 		Usr: &legolog_grpcint.Username{
// 			Username: username,
// 		},
// 		Size: 0,
// 		Pos: &legolog_grpcint.Position{
// 			Pos: uint64(pos),
// 		},
// 		Height: uint32(height),
// 	}
// 	_, err := c.legologClient.GetMonitoringProofForTest(ctx, proofrequest)
// 	if err != nil {
// 		fmt.Println("could not get proof, trying again in next verify iteration")
// 		return err
// 	}
// 	return nil
// }

// func (c *Client) monitoringInt(ctx context.Context, username []byte,
// 	pos int, height int, vrf []byte, Ek []byte, signature []byte,
// 	keyhash []core.KeyHash) (*legolog_grpcint.GetPublicKeyProofResponse, error) {
// 	auditorResponse, err := c.auditorClient.GetEpochUpdate(ctx)
// 	if err != nil {
// 		fmt.Println("Could not query auditor server, skipping this verify cycle")
// 		return nil, err
// 	}
// 	var digest = new(core.Digest)
// 	marshaledDigest := auditorResponse.GetCkPoint().GetMarshaledDigest()
// 	json.Unmarshal(marshaledDigest, &digest)

// 	proofrequest := &legolog_grpcint.GetPublicKeyProofRequest{
// 		Usr: &legolog_grpcint.Username{
// 			Username: username,
// 		},
// 		Size: uint64(digest.Size),
// 		Pos: &legolog_grpcint.Position{
// 			Pos: uint64(pos),
// 		},
// 		Height: uint32(height),
// 	}

// 	proofResponse, err := c.legologClient.GetPublicKeyProof(ctx, proofrequest)
// 	if err != nil {
// 		fmt.Println("could not get proof, trying again in next verify iteration")
// 		return nil, err
// 	}

// 	if err != nil {
// 		fmt.Println("could not get proof, trying again in next verify iteration")
// 		return nil, err
// 	}

// 	var appendProof core.MerkleExistenceProof
// 	json.Unmarshal(proofResponse.GetProof(), &appendProof)

// 	nodeHash := core.ComputeLeafNodeHash(vrf, Ek, signature, uint32(pos))
// 	oldHashes := keyhash
// 	proven, _, _ := core.VerifyExistenceProof(digest, nodeHash, vrf, uint32(pos),
// 		uint32(height), &appendProof, oldHashes)
// 	if !proven {
// 		fmt.Printf("WARNING: could not prove existence at position %v \n", pos)
// 	}
// 	return proofResponse, nil
// }

// func (c *Client) RequestAuditorForSize(ctx context.Context) (int, error) {
// 	auditorResponse, err := c.auditorClient.GetEpochUpdate(ctx)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return proto.Size(auditorResponse), nil
// }
