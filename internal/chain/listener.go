package chain

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)


type LogListener struct{
	Client *ethclient.Client
}


type BlockListener struct{
	Client *ethclient.Client
}


func (b *BlockListener) SubscribeToBlocks(ctx context.Context, out chan <- *types.Header)error{
	headers := make(chan *types.Header)

	sub,err := b.Client.SubscribeNewHead(ctx,headers)

	if err !=nil{
		return err
	}

	go func ()  {
		for{
			select{
			case err:= <-sub.Err():
				log.Println("Block subscription error:", err)
                return
			case header := <-headers:
				out <- header
			case <-ctx.Done():
				log.Println("Block listener stopped")
                return
			}
		}
		
	}()
	return nil
}

func (l *LogListener) SubscribeToLogs(ctx context.Context,query ethereum.FilterQuery,out chan <- types.Log)error{
  logChan := make(chan types.Log)
	sub,err := l.Client.SubscribeFilterLogs(ctx,query,logChan)
	if err !=nil{
		return err

	}
	go func(){
		for{
			select{
			case err := <-sub.Err():
				log.Println("Log Subscription Error",err)
				return
			case logs := <- logChan:
				out <- logs
			case <-ctx.Done():
				log.Println("Logs listener stopped")
                return
			}
		}

	}()
	return nil
}

func (b *BlockListener) SubscribeToBlocksPolling(ctx context.Context, ch chan<- *types.Header) error {
    go func() {
        var lastBlock uint64

        for {
            select {
            case <-ctx.Done():
                return
            default:
            }

            header, err := b.Client.HeaderByNumber(ctx, nil) // nil = latest
            if err != nil {
                fmt.Printf("[WARN] poll error: %v — retrying in 2s\n", err)
                time.Sleep(2 * time.Second)
                continue
            }

            if header.Number.Uint64() > lastBlock {
                lastBlock = header.Number.Uint64()
                ch <- header
            }

            //time.Sleep(2 * time.Second)
        }
    }()
    return nil
}