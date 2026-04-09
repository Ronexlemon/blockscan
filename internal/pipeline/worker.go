package pipeline

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
)


func BlockWorkerSubscriber(ctx context.Context,workercount int,in <-chan *types.Header,handler func(blocs *types.Header))error{
	for i:=0;i<workercount; i++{
		go func(id int){
			for{
				select{
				case blocks:= <-in:
					handler(blocks)
				case <-ctx.Done():
					log.Printf("Worker %d stopped\n", id)
                    return

				}
			}

		}(i)

	}
	return nil
}



func LogWorkerSubscriber(ctx context.Context,workercount int,in <-chan types.Log,handler func(blocs types.Log))error{
	for i:=0;i<workercount; i++{
		go func(id int){
			for{
				select{
				case logsEvent := <-in:
					handler(logsEvent)
				case <-ctx.Done():
					log.Printf("Worker %d stopped\n", id)
                    return

				}
			}

		}(i)

	}
	return nil
}