package pipeline

import (
	//"context"

	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)


type BlockProcessor struct{
	Client *ethclient.Client
}

var (
    TransferTopic     = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f164f4f5b3cd39f0c5")
    ApprovalTopic     = common.HexToHash("0x8c5be1e5ebec7d5bd14f714f3e3d65b70a5a3e5a4c238e5b236e4d5e4c1d49a7")
    MintTopic         = common.HexToHash("0x40c10f1908d2a7b70c65a1ce0e72a5a742a5190d5b9dc159f3d34f50b2d8c5d9")
)

func (p *BlockProcessor) ProcessBlock(ctx context.Context,header *types.Header)([]types.Log,error){
	fmt.Println("THE BLOCK HASH",header.Number)
 logs,err := p.Client.FilterLogs(ctx,ethereum.FilterQuery{
	FromBlock: header.Number,
	ToBlock: header.Number,
	//Topics: [][]common.Hash{{TransferTopic, ApprovalTopic, MintTopic}},
  })

  if err != nil{
	fmt.Println("Failed")
	return nil, nil
  }
	fmt.Println("EVENTS",logs)
  return logs,nil

  
}