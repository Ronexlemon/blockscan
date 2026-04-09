package chain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ronexlemon/blockscan/pkg"
)




type ChainClient struct{
	Client  *ethclient.Client
	NAME    string
}

func NewchainConfig(rpcUrl string, name string)*ChainClient{
	client, err := connectClient(rpcUrl)

	if err !=nil{
		fmt.Println("Failed to create a connection")
		return  nil

	}
	return&ChainClient{Client: client,NAME: name}
}
func connectClient(rpcUlr string)(*ethclient.Client,error){
	client ,err := ethclient.Dial(rpcUlr)

	if err !=nil{
		fmt.Println("Failed to create a connection")
		return nil,pkg.WrapError("failed to dial RPC",err)
	}

	return client,nil
	
}


func (c *ChainClient) GetLatestBlock(ctx context.Context)(uint64,error){
	latestBlock,err := c.Client.BlockNumber(ctx)
	if err !=nil{
		return 0,pkg.WrapError("failed to get the latest block",err)
	}
	return latestBlock,nil
}

func (c *ChainClient) GetBlockByNumber(ctx context.Context , number *big.Int)(*types.Block,error){
	block,err:= c.Client.BlockByNumber(ctx,number)
	if err !=nil{
		return nil,pkg.WrapError("failed to get the blockBy Number",err)
	}
	return block,nil
}