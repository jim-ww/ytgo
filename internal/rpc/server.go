package rpc

import (
	"sync"

	"github.com/jim-ww/ytgo/internal/scraper"
	"github.com/jim-ww/ytgo/internal/types"
)

type Server struct {
	store    *types.StoreData
	mu       sync.RWMutex
	searcher scraper.YouTubeSearcher
}

func StartServer(socket string) error {
	// net.Listen("unix", socket) + rpc.ServeCodec(jsonrpc.NewServerCodec(conn), rpcServer)
	return nil
}
