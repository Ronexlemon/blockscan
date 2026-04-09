package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ronexlemon/blockscan/internal/chain"
    "github.com/ronexlemon/blockscan/internal/pipeline"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    wss := "https://forno.celo.org"
    fmt.Println("Starting to listen to blocks")

    client := chain.NewchainConfig(wss, "Celo")
    if client == nil {
        log.Fatal("[FATAL] failed to connect to chain")
    }

    blocklistener := chain.BlockListener{Client: client.Client}

    processor := &pipeline.BlockProcessor{
        Client:    client.Client,
        RpcClient: client.RpcClient,
        RpcURL:    wss,
        Sem:       make(chan struct{}, 5), // max 5 concurrent RPC calls
    }

    blockChan := make(chan *types.Header, 1000)

    if err := blocklistener.SubscribeToBlocksPolling(ctx, blockChan); err != nil {
        log.Fatal(err)
    }

    pipeline.BlockWorkerSubscriber(ctx, 10, blockChan, processor)
    select {}
}