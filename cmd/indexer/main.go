package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/joho/godotenv"
	//"github.com/ronexlemon/blockscan/internal/api"
	"github.com/ronexlemon/blockscan/internal/broadcast"
	"github.com/ronexlemon/blockscan/internal/chain"
	"github.com/ronexlemon/blockscan/internal/pipeline"
	"github.com/ronexlemon/blockscan/internal/storage"
)

func main() {
    // Load .env before anything else
    if err := godotenv.Load(); err != nil {
        log.Println("[WARN] no .env file found, using system env")
    }
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
     db := storage.NewDataBaseConnection()
     repo := storage.Repository{Db: db.DB}
     if err := repo.Migrate(); err != nil {
		log.Fatalf("[FATAL] migrate: %v", err)
	}
    // hub:= api.NewSSEHub()
    publisher := broadcast.NewPublisher()
    //go hub.Run()
    wss := "https://forno.celo.org"
    fmt.Println("Starting to listen to blocks")

    client := chain.NewchainConfig(wss, "Celo")
    if client == nil {
        log.Fatal("[FATAL] failed to connect to chain")
    }

    resultChan := make(chan *pipeline.BlockResult,500)
    blocklistener := chain.BlockListener{Client: client.Client}

    processor := &pipeline.BlockProcessor{
        Client:    client.Client,
        RpcClient: client.RpcClient,
        RpcURL:    wss,
        Sem:       make(chan struct{}, 5), // max 5 concurrent RPC calls
        OnResult: func(result *pipeline.BlockResult) {
            select{
            case resultChan <- result:
            default:
                fmt.Printf("[WARN] result channel full, dropping block %d\n",
                result.BlockNumber)
            }
            publisher.Publish(ctx,"block",map[string]any{
                "block_number": result.BlockNumber,
                "total_txs":result.TotalTxs,
                "call_count":len(result.MethodCalls),
            })
            for _,decoded := range result.DecodedCalls{
                publisher.Publish(ctx,"transaction",decoded)
            }
        },
    }

    blockChan := make(chan *types.Header, 1000)

    if err := blocklistener.SubscribeToBlocksPolling(ctx, blockChan); err != nil {
        log.Fatal(err)
    }

    pipeline.BlockWorkerSubscriber(ctx, 10, blockChan, processor)
    storage.StartDbWorker(ctx,resultChan,repo,5)
    storage.StartRetentionWorker(ctx,&repo)
    select {}
}