package main

import (
	"context"
	"fmt"
	"log"

	//"github.com/ethereum/go-ethereum"
	//"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ronexlemon/blockscan/internal/chain"
	"github.com/ronexlemon/blockscan/internal/pipeline"
)


func main(){
	ctx,cancel := context.WithCancel(context.Background())

	defer cancel()

	wss:= "wss://forno.celo-sepolia.celo-testnet.org/ws"
	fmt.Println("Staring to ;listen to blocks")

	client := chain.NewchainConfig(wss,"celo")

	blocklistener := chain.BlockListener{Client: client.Client}
	processor := pipeline.BlockProcessor{Client: client.Client}
	// logListener   := chain.LogListener{Client: client.Client}
	// logChan := make(chan types.Log, 1000)
	blockChan := make(chan *types.Header,1000)
	// query := ethereum.FilterQuery{
    //     Addresses: []common.Address{common.HexToAddress("0xYourContract")},
    // }

	// if err := logListener.SubscribeToLogs(ctx,query,logChan);err !=nil{
	// 	log.Fatal(err)
	// }

	if err := blocklistener.SubscribeToBlocks(ctx,blockChan);err !=nil{
		log.Fatal(err)
	}

	// pipeline.LogWorkerSubscriber(ctx,10,logChan,func(ev types.Log){
	// 	  fmt.Println("Processing Log:", ev.TxHash.Hex())
	// })

	pipeline.BlockWorkerSubscriber(ctx,10,blockChan,&processor)

	select {}


}