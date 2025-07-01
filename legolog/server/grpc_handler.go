package legolog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/huyuncong/MerkleSquare/core"
	legolog_grpcint "github.com/huyuncong/MerkleSquare/legolog/legolog-grpcint"
	"github.com/immesys/bw2/crypto"
)

func (s *Server) Register(ctx context.Context, req *legolog_grpcint.RegisterRequest) (
	*legolog_grpcint.RegisterResponse, error) {
	// var p core.Partition

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	position, err := s.RegisterUserKey(ctx, req.GetUsr().GetUsername(),
		req.GetKey().GetMk(), req.GetSignature(), true)
	if err != nil {
		return nil, err
	}

	return &legolog_grpcint.RegisterResponse{
		Pos: &legolog_grpcint.Position{Pos: position},
		// VrfKey: s.vrfPrivKey.Compute(req.GetUsr().GetUsername()),
	}, nil
}

func (s *Server) Append(stream legolog_grpcint.LegoLog_AppendServer) error {
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	ctx := context.Background()
	user, identifier, value := req.GetUsr().Username, req.Identifier.GetIdentifier(), req.Value.GetValue()
	partitionServer := s.GetPartitionForIdentifier(identifier)
	mkSerialized, _ := partitionServer.Storage.Get(ctx, append(user, []byte("MK")...))
	if mkSerialized == nil {
		return errors.New("User is not registered")
	}
	var mk ValueRecord
	err = json.Unmarshal(mkSerialized, &mk)
	if err != nil {
		return err
	}

	partitionServer.LastPosLock.Lock()
	partitionServer.AppendLock.Lock()
	// fmt.Println("AKSHAYR: Starting with lastpos equal to", partitionServer.LastPos, "at", time.Now().UnixNano())
	position := partitionServer.LastPos
	partitionServer.LastPos += 1
	// fmt.Println("Ending with lastpos equal to", partitionServer.LastPos, "at", time.Now().UnixNano())
	// fmt.Println("About to send position ", position)

	//Send position
	var response = &legolog_grpcint.AppendResponse{
		Pos: &legolog_grpcint.Position{Pos: position},
	}
	stream.Send(response)
	req, err = stream.Recv()
	if err != nil {
		return err
	}
	signature := req.GetSignature()
	//Verify
	if !crypto.VerifyBlob(mk.Value, signature,
		append(value, []byte(strconv.Itoa(int(position)))...)) {
		return errors.New("Verification failed")
	}
	//Add to merkle tree
	partitionServer.Partition.Append(user, identifier, value, signature)
	partitionServer.AppendLock.Unlock()
	partitionServer.LastPosLock.Unlock()

	//4. Add to K-V store
	var serializedValue []byte
	// Prepend to existing entry
	original, _ := partitionServer.Storage.Get(ctx, identifier)
	valueRecord := make([]ValueRecord, 1)
	valueRecord[0] = ValueRecord{
		Position:  position,
		Signature: signature,
		Value:     value,
	}
	if original == nil {
		serializedValue, _ = json.Marshal(valueRecord)
	} else {
		var deserialized []ValueRecord
		json.Unmarshal(original, &deserialized)
		serializedValue, _ = json.Marshal(append(valueRecord, deserialized...))
	}
	partitionServer.Storage.Put(ctx, identifier, serializedValue)
	// response.VrfKey = s.vrfPrivKey.Compute(req.GetUsr().GetUsername())
	response.Completed = true
	stream.Send(response)
	return nil

	/*

		MERKLE^2 Implementation

		req, err := stream.Recv()
		if err != nil {
			return err
		}
		ctx := context.Background()

		user, key := req.GetUsr().GetUsername(), req.GetEk().GetEk()
		mkSerialized, _ := s.Storage.Get(ctx, append(user, []byte("MK")...))
		queryString := append(user, []byte("PK")...)
		if mkSerialized == nil {
			return errors.New("User is not registered")
		}
		var mk KeyRecord
		err = json.Unmarshal(mkSerialized, &mk)
		if err != nil {
			return err
		}

		s.LastPosLock.Lock()
		s.appendLock.Lock()
		position := s.LastPos
		s.LastPos++
		//Send position
		var response = &legolog_grpcint.AppendResponse{
			Pos: &legolog_grpcint.Position{Pos: position},
		}
		stream.Send(response)
		req, err = stream.Recv()
		if err != nil {
			return err
		}
		signature := req.GetSignature()
		//Verify
		if !crypto.VerifyBlob(mk.Key, signature,
			append(key, []byte(strconv.Itoa(int(position)))...)) {
			return errors.New("Verification failed")
		}
		//Add to merkle tree
		s.MerkleSquare.Append(s.vrfPrivKey.Compute(user), key, signature)
		s.appendLock.Unlock()
		s.LastPosLock.Unlock()

		//4. Add to K-V store
		var serializedKey []byte
		// Prepend to existing entry
		original, _ := s.Storage.Get(ctx, queryString)
		keyrecord := make([]KeyRecord, 1)
		keyrecord[0] = KeyRecord{
			Position:  position,
			Signature: signature,
			Key:       key,
		}
		if original == nil {
			serializedKey, _ = json.Marshal(keyrecord)
		} else {
			var deserialized []KeyRecord
			json.Unmarshal(original, &deserialized)
			serializedKey, _ = json.Marshal(append(keyrecord, deserialized...))
		}
		s.Storage.Put(ctx, queryString, serializedKey)
		response.VrfKey = s.vrfPrivKey.Compute(req.GetUsr().GetUsername())
		response.Completed = true
		stream.Send(response)
		return nil
	*/
}

func (s *Server) LookUpMK(ctx context.Context, req *legolog_grpcint.LookUpMKRequest) (
	*legolog_grpcint.LookUpMKResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	queryString := append(req.GetUsr().GetUsername(), []byte("MK")...)
	partitionServer := s.GetPartitionForIdentifier(queryString)

	var MK ValueRecord
	serializedMK, err := partitionServer.Storage.Get(ctx, queryString)
	json.Unmarshal(serializedMK, &MK)

	// key, sign, pos, err := s.GetUserKey(ctx, req.GetUsr().GetUsername(), true, 0)
	if err != nil {
		return nil, err
	}

	return &legolog_grpcint.LookUpMKResponse{
		Imk: &legolog_grpcint.IndexedMK{
			Pos:       &legolog_grpcint.Position{Pos: MK.Position},
			MasterKey: &legolog_grpcint.MasterKey{Mk: MK.Value},
		},
		Signature: MK.Signature,
		// VrfKey:    s.vrfPrivKey.Compute(req.GetUsr().GetUsername()),
	}, nil
}

// TODO: modify to work with various positions instead of just latest PK
func (s *Server) LookUpPK(ctx context.Context, req *legolog_grpcint.LookUpPKRequest) (
	*legolog_grpcint.LookUpPKResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var identifier []byte = req.Identifier.GetIdentifier()
	partitionServer := s.GetPartitionForIdentifier(identifier)

	var keys []ValueRecord

	var serializedKey, err = partitionServer.Storage.Get(ctx, identifier)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(serializedKey, &keys)

	if keys == nil || len(keys) == 0 {
		return nil, errors.New("No keys found in storage for identifier " + string(identifier[:]))
	}

	var latest_key ValueRecord = keys[0] // convention is to store most recent element at the beginning

	return &legolog_grpcint.LookUpPKResponse{
		IndexedValue: &legolog_grpcint.IndexedValue{
			Value: &legolog_grpcint.Value{Value: latest_key.Value},
			Pos:   &legolog_grpcint.Position{Pos: latest_key.Position},
		},
		Signature: latest_key.Signature,
	}, nil

	/*
		MERKLE^2 Implementation

		var reqPos uint64
		if req.GetPos() == nil {
			s.epochLock.RLock()
			reqPos = s.PublishedPos
			s.epochLock.RUnlock()
		} else {
			reqPos = req.GetPos().GetPos()
		}

		key, sign, pos, err := s.GetUserKey(ctx, req.GetUsr().GetUsername(), false, reqPos)
		if err != nil {
			return nil, err
		}

		return &legolog_grpcint.LookUpPKResponse{
			Iek: &legolog_grpcint.IndexedEK{
				Pos:       &legolog_grpcint.Position{Pos: pos},
				PublicKey: &legolog_grpcint.EncryptionKey{Ek: key},
			},
			Signature: sign,
			VrfKey:    s.vrfPrivKey.Compute(req.GetUsr().GetUsername()),
		}, nil
	*/
}

func (s *Server) LookUpMKVerify(ctx context.Context,
	req *legolog_grpcint.LookUpMKVerifyRequest) (
	*legolog_grpcint.LookUpMKVerifyResponse, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// TODO: check what this is for...
	// if req.Size == 0 {
	// 	s.epochLock.RLock()
	// 	req.Size = s.PublishedPos
	// 	s.epochLock.RUnlock()
	// }
	identifier := append(req.GetUsr().GetUsername(), []byte("MK")...)
	partitionServer := s.GetPartitionForIdentifier(identifier)

	lookupMKResponse, err := s.LookUpMK(ctx, &legolog_grpcint.LookUpMKRequest{Usr: req.Usr}) // s.GetUserKey(ctx, req.GetUsr().GetUsername(), false, req.Size)

	masterKey, pos, sign := lookupMKResponse.Imk.MasterKey, lookupMKResponse.Imk.Pos, lookupMKResponse.Signature

	proof := partitionServer.Partition.GenerateExistenceProof(identifier, masterKey.Mk, sign)

	marshaledProof, err := json.Marshal(proof)
	return &legolog_grpcint.LookUpMKVerifyResponse{
		IndexedValue: &legolog_grpcint.IndexedValue{
			Pos:   &legolog_grpcint.Position{Pos: pos.Pos},
			Value: &legolog_grpcint.Value{Value: masterKey.Mk},
		},
		Signature: sign,
		Proof:     marshaledProof,
	}, err
}

func (s *Server) LookUpPKVerify(ctx context.Context, req *legolog_grpcint.LookUpPKVerifyRequest) (
	*legolog_grpcint.LookUpPKVerifyResponse, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	partitionServer := s.GetPartitionForIdentifier(req.Identifier.Identifier)
	/* 	if req.Size == 0 {
		partitionServer.epochLock.RLock()
		req.Size = partitionServer.PublishedPos
		partitionServer.epochLock.RUnlock()
	} */

	lookupPKResponse, err := s.LookUpPK(ctx, &legolog_grpcint.LookUpPKRequest{Identifier: req.Identifier}) // s.GetUserKey(ctx, req.GetUsr().GetUsername(), false, req.Size)
	// vrfKey := s.vrfPrivKey.Compute(req.GetUsr().GetUsername())

	/* 	key, sign, pos, err */
	indexedValue, sign := lookupPKResponse.IndexedValue, lookupPKResponse.Signature
	if err != nil {
		return nil, err
	}

	proof := partitionServer.Partition.GenerateExistenceProof(req.Identifier.Identifier, indexedValue.Value.Value, sign)

	//fmt.Println("generated existence proof ", proof)
	/* 	proof := s.MerkleSquare.ProveLatest(vrfKey, key, uint32(pos), uint32(req.Size))*/
	marshaledProof, err := json.Marshal(proof)

	return &legolog_grpcint.LookUpPKVerifyResponse{
		IndexedValue: indexedValue,
		Signature:    sign,
		/* 		VrfKey:       vrfKey, */
		Proof: marshaledProof,
	}, err
}

// TODO: for now will just return digest, but needs to be modified later
func (s *Server) GetNewCheckPoint(ctx context.Context,
	req *legolog_grpcint.GetNewCheckPointRequest) (
	*legolog_grpcint.GetNewCheckPointResponse, error) {

	// return partition's diges
	if int(req.PartitionIndex) >= len(s.PartitionServers) {
		return nil, fmt.Errorf("Partition out of bounds: %d", req.PartitionIndex)
	}
	partitionServer := s.PartitionServers[req.PartitionIndex]

	/* partitionServer.Partition.PublishedDigest

	partitionServer.Partition.GetDigest()

	*/
	d := partitionServer.PublishedDigest
	marshalledDigest, _ := json.Marshal(d)
	// fmt.Println("the marshalled digest is", marshalledDigest)
	var unmarshalledDigest core.LegologDigest
	err := json.Unmarshal(marshalledDigest, &unmarshalledDigest)
	if err != nil {
		fmt.Println("Error unmarshalling digest in getnewcheckpoint", err)
	}
	// fmt.Println("unmarshalled digest is ", unmarshalledDigest)

	return &legolog_grpcint.GetNewCheckPointResponse{
		Checkpoint: &legolog_grpcint.CheckPoint{
			MarshaledDigest: marshalledDigest,
		},
	}, nil

	// if err := ctx.Err(); err != nil {
	// 	return nil, err
	// }

	// if req.OldSize > s.PublishedPos {
	// 	return nil, errors.New("Auditor expects more leaves than what server published")
	// }

	// digest, err := json.Marshal(s.PublishedDigest)
	// if err != nil {
	// 	return nil, err
	// }

	// var marshaledProof []byte
	// proof, ok := s.extensionProofCache[ExtensionProofKey{req.OldSize, s.PublishedPos}]
	// if s.CacheExtensionProofs && ok {
	// 	marshaledProof = proof
	// } else {
	// 	extensionProof := s.MerkleSquare.GenerateExtensionProof(
	// 		uint32(req.OldSize), uint32(s.PublishedPos))
	// 	marshaledProof, err = json.Marshal(extensionProof)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if s.CacheExtensionProofs {
	// 		s.extensionProofCache[ExtensionProofKey{req.OldSize, s.PublishedPos}] = marshaledProof
	// 	}
	// }

	// return &legolog_grpcint.GetNewCheckPointResponse{
	// 	CkPoint: &legolog_grpcint.CheckPoint{
	// 		MarshaledDigest: digest,
	// 		NumLeaves:       s.PublishedPos,
	// 		Epoch:           s.epoch,
	// 	},
	// 	Proof: marshaledProof,
	// }, nil
}

// TODO: for now will just return digest, but needs to be modified later
func (s *Server) GetNewUpdateCheckPoint(ctx context.Context,
	req *legolog_grpcint.GetNewCheckPointRequest) (
	*legolog_grpcint.GetNewCheckPointResponse, error) {

	// return partition's diges
	if int(req.PartitionIndex) >= len(s.PartitionServers) {
		return nil, fmt.Errorf("Partition out of bounds: %d", req.PartitionIndex)
	}
	partitionServer := s.PartitionServers[req.PartitionIndex]

	/* partitionServer.Partition.PublishedDigest

	partitionServer.Partition.GetDigest()

	*/
	d := partitionServer.PublishedDigest
	marshalledDigest, _ := json.Marshal(d)
	// fmt.Println("the marshalled digest is", marshalledDigest)
	var unmarshalledDigest core.LegologDigest
	err := json.Unmarshal(marshalledDigest, &unmarshalledDigest)
	if err != nil {
		fmt.Println("Error unmarshalling digest in getnewcheckpoint", err)
		return nil, err
	}
	// fmt.Println("unmarshalled digest is ", unmarshalledDigest)
	proof := partitionServer.Partition.GetUpdateEpochConsistencyProof(uint32(req.OldSize))
	marshalledProof, err := json.Marshal(proof)
	if err != nil {
		fmt.Println("Error marshalling proof in getnewcheckpoint", err)
		return nil, err
	}
	return &legolog_grpcint.GetNewCheckPointResponse{
		Checkpoint: &legolog_grpcint.CheckPoint{
			MarshaledDigest: marshalledDigest,
		},
		Proof: marshalledProof,
	}, nil
}

// TODO: for now will just return digest, but needs to be modified later
func (s *Server) GetNewVerifyCheckPoint(ctx context.Context,
	req *legolog_grpcint.GetNewCheckPointRequest) (
	*legolog_grpcint.GetNewCheckPointResponse, error) {

	// return partition's diges
	if int(req.PartitionIndex) >= len(s.PartitionServers) {
		return nil, fmt.Errorf("Partition out of bounds: %d", req.PartitionIndex)
	}
	partitionServer := s.PartitionServers[req.PartitionIndex]

	/* partitionServer.Partition.PublishedDigest

	partitionServer.Partition.GetDigest()

	*/
	d := partitionServer.PublishedDigest
	marshalledDigest, _ := json.Marshal(d)
	// fmt.Println("the marshalled digest is", marshalledDigest)
	var unmarshalledDigest core.LegologDigest
	err := json.Unmarshal(marshalledDigest, &unmarshalledDigest)
	if err != nil {
		fmt.Println("Error unmarshalling digest in getnewcheckpoint", err)
	}
	// fmt.Println("unmarshalled digest is ", unmarshalledDigest)

	return &legolog_grpcint.GetNewCheckPointResponse{
		Checkpoint: &legolog_grpcint.CheckPoint{
			MarshaledDigest: marshalledDigest,
		},
	}, nil
}

func (s *Server) GetMasterKeyProof(ctx context.Context,
	req *legolog_grpcint.GetMasterKeyProofRequest) (
	*legolog_grpcint.GetMasterKeyProofResponse, error) {
	/*
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// copy this logic from below getpk proof
		proof := s.Partition.ProveNonexistence(
			s.vrfPrivKey.Compute(req.GetUsr().GetUsername()),
			uint32(req.GetPos().GetPos()), uint32(req.Size))
		marshaledProof, err := json.Marshal(proof)
		if err != nil {
			return nil, err
		}

		return &legolog_grpcint.GetMasterKeyProofResponse{
			Proof: marshaledProof,
		}, nil
	*/
	return &legolog_grpcint.GetMasterKeyProofResponse{
		Proof: nil,
	}, nil
}

func (s *Server) GetPublicKeyProof(ctx context.Context,
	req *legolog_grpcint.GetPublicKeyProofRequest) (*legolog_grpcint.GetPublicKeyProofResponse, error) {
	/*
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// what is the key for generate existence proof?
		proof := s.Partition.GenerateExistenceProof(
			req.GetIdentifier(), req.GetHeight(), uint32(req.Size))
		marshaledProof, err := json.Marshal(proof)
		if err != nil {
			return nil, err
		}

		// if req.Size == 0 {
		// 	s.epochLock.RLock()
		// 	req.Size = s.PublishedPos
		// 	s.epochLock.RUnlock()
		// }
		// proof := s.MerkleSquare.GenerateExistenceProof(
		// 	s.vrfPrivKey.Compute(req.GetUsr().GetUsername()), uint32(req.GetPos().GetPos()),
		// 	req.GetHeight(), uint32(req.Size))
		// marshaledProof, err := json.Marshal(proof)
		// if err != nil {
		// 	return nil, err
		// }

		return &legolog_grpcint.GetPublicKeyProofResponse{
			Proof: marshaledProof,
		}, nil
	*/
	return &legolog_grpcint.GetPublicKeyProofResponse{
		Proof: nil,
	}, nil
}

// func (s *Server) GetMonitoringProofForTest(ctx context.Context,
// 	req *legolog_grpcint.GetMonitoringProofForTestRequest) (
// 	*legolog_grpcint.GetMonitoringProofForTestResponse, error) {
// if err := ctx.Err(); err != nil {
// 	return nil, err
// }
// if req.Size == 0 {
// 	s.epochLock.RLock()
// 	req.Size = s.PublishedPos
// 	s.epochLock.RUnlock()
// }

// proof := s.MerkleSquare.GenerateExistenceProof(
// 	s.vrfPrivKey.Compute(req.GetUsr().GetUsername()),
// 	uint32(req.GetPos().GetPos()), req.GetHeight(), uint32(req.Size))
// marshaledProof, err := json.Marshal(proof)
// if err != nil {
// 	fmt.Println(uint32(req.GetPos().GetPos()))
// 	fmt.Println(req.GetHeight())
// 	fmt.Println(uint32(req.Size))
// 	fmt.Println(err.Error())
// 	return nil, err
// }

// return &legolog_grpcint.GetMonitoringProofForTestResponse{
// 	Proof: marshaledProof,
// }, nil

// 	return nil, nil
// }

/*

func (s *Server) GetLookUpProof(ctx context.Context,
	req *legolog_grpcint.GetLookUpProofRequest) (
	*legolog_grpcint.GetLookUpProofResponse, error) {
	// var err error
	// if err := ctx.Err(); err != nil {
	// 	return nil, err
	// }
	// var marshaledProof []byte
	// if req.GetIsMasterKey() {
	// 	proof := s.MerkleSquare.ProveFirst(
	// 		s.vrfPrivKey.Compute(req.GetUsr().GetUsername()),
	// 		req.GetMasterKey().GetMk(), uint32(req.GetPos().GetPos()), uint32(req.Size))
	// 	marshaledProof, err = json.Marshal(proof)
	// } else {
	// 	proof := s.MerkleSquare.ProveLatest(
	// 		s.vrfPrivKey.Compute(req.GetUsr().GetUsername()),
	// 		req.GetEncryptionKey().GetEk(), uint32(req.GetPos().GetPos()), uint32(req.Size))
	// 	marshaledProof, err = json.Marshal(proof)
	// }
	// if err != nil {
	// 	return nil, err
	// }

	// return &legolog_grpcint.GetLookUpProofResponse{
	// 	Proof: marshaledProof,
	// }, nil

	return nil, nil
}

*/
