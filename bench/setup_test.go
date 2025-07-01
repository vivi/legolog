package bench

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/huyuncong/MerkleSquare/client"

	"github.com/huyuncong/MerkleSquare/lib/storage"

	"github.com/immesys/bw2/crypto"
)

func SetUp(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ctx := context.Background()

	//Setup servers
	dir, err := ioutil.TempDir("", "teststore")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	db := storage.OpenFile(dir)
	if err != nil {
		panic(err)
	}
	defer db.Close(ctx)

	serv := setupServer(db)

	//Client test
	c, err := client.NewClient(serverAddr, "", "")
	if err != nil {
		t.Error(errors.New("Failed to start client: " + err.Error()))
	}

	masterSK, masterVK := crypto.GenerateKeypair()
	_, publicVK := crypto.GenerateKeypair()
	c.Register(ctx, []byte("testuser"), masterSK, masterVK)

	PopulateStorage(serv, 1000000, 100)

	start := time.Now()
	c.Append(ctx, []byte("testuser"), publicVK)
	c.Register(ctx, []byte("testuser2"), masterSK, masterVK)
	fmt.Println(time.Since(start))
}
