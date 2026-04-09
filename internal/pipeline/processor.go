package pipeline

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)


type BlockProcessor struct{
	Client *ethclient.Client
}

func (p *BlockProcessor) BlockProcessor(ctx context.Context,header *types.Header)([]types.Log)