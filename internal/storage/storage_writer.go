package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ronexlemon/blockscan/internal/pipeline"
)

const(
	 MaxBlocks = 2000
	 RetentionInterval = 5 *time.Minute

)
//Saving to db worker

func StartDbWorker(ctx context.Context,ch <- chan *pipeline.BlockResult,repo Repository,n int){

	for i := 0; i < n; i++ {
		go func (workerId int)  {
			for{
				select {
				case result,ok := <-ch:
					if !ok {
						return
					}
					if err := persistResult(ctx,repo,result); err !=nil{
						fmt.Printf("[DB Worker %d] error block %d: %v\n",workerId,result.BlockNumber,err)
					}
					fmt.Println("Received results",result.DecodedCalls)
				case <- ctx.Done():
					return

					
					
				}
			}
			
		}(i)
	}

	fmt.Printf("[DB] started %d workers\n",n)

}


//retention worker

func StartRetentionWorker(ctx context.Context,repo *Repository){

	err := repo.RunRetention(ctx,MaxBlocks)
	if err !=nil{
		log.Printf("[Retention] initial run error: %v\n", err)

	}

	ticker := time.NewTicker(RetentionInterval)
	defer ticker.Stop()

	for{
		select{
		case <-ticker.C:
			if err := repo.RunRetention(ctx,MaxBlocks); err !=nil{
				log.Printf("[Retention] error: %v\n", err)

			}
		case <- ctx.Done():
			log.Println("[Retention] worker stopped")
			return


		}
	}

}

func persistResult(ctx context.Context, repo  Repository, result *pipeline.BlockResult) error {
	exists, err := repo.IsBlockProcessed(ctx, result.BlockNumber)
	if err != nil {
		return fmt.Errorf("check block: %w", err)
	}
	if exists {
		fmt.Printf("[DB] block %d already stored, skipping\n", result.BlockNumber)
		return nil
	}

	block, calls := BlockResultToRecords(result)

	
	if err := repo.SaveBlock(ctx, block); err != nil {
		return fmt.Errorf("save block: %w", err)
	}

	
	if err := repo.SaveDecodedCallBatch(ctx, calls); err != nil {
		return fmt.Errorf("save calls: %w", err)
	}

	

	fmt.Printf("[DB] saved block %d — calls=%d \n",
		result.BlockNumber, len(calls),)

	return nil
}