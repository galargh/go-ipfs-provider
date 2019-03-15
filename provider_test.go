package provider

import (
	"context"
	"math/rand"
	"testing"
	"time"

	blocksutil "github.com/ipfs/go-ipfs-blocksutil"
	cid "github.com/ipfs/go-cid"
	datastore "github.com/ipfs/go-datastore"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	sync "github.com/ipfs/go-datastore/sync"
)

var blockGenerator = blocksutil.NewBlockGenerator()

type mockRouting struct {
	provided chan cid.Cid
}

func mockContentRouting() *mockRouting {
	r := mockRouting{}
	r.provided = make(chan cid.Cid)
	return &r
}

func TestAnnouncement(t *testing.T) {
	ctx := context.Background()
	defer ctx.Done()

	ds := sync.MutexWrap(datastore.NewMapDatastore())
	queue, err := NewQueue(ctx, "test", ds)
	if err != nil {
		t.Fatal(err)
	}

	r := mockContentRouting()

	provider := NewProvider(ctx, queue, r)
	provider.Run()

	cids := cid.NewSet()

	for i := 0; i < 1000; i++ {
		c := blockGenerator.Next().Cid()
		cids.Add(c)
	}

	go func() {
		for _, c := range cids.Keys() {
			err = provider.Provide(c)
			// A little goroutine stirring to exercise some different states
			r := rand.Intn(10)
			time.Sleep(time.Microsecond * time.Duration(r))
		}
	}()

	for cids.Len() > 0 {
		select {
			case cp := <-r.provided:
				if !cids.Has(cp) {
					t.Fatal("Wrong CID provided")
				}
				cids.Remove(cp)
			case <-time.After(time.Second * 5):
				t.Fatal("Timeout waiting for cids to be provided.")
		}
	}
}

func (r *mockRouting) Provide(ctx context.Context, cid cid.Cid, recursive bool) error {
	r.provided <- cid
	return nil
}

// Search for peers who are able to provide a given key
func (r *mockRouting) FindProvidersAsync(ctx context.Context, cid cid.Cid, timeout int) <-chan pstore.PeerInfo {
	return nil
}