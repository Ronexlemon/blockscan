package chain

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/rpc"
    "github.com/ronexlemon/blockscan/pkg"
)

type ChainClient struct {
    Client    *ethclient.Client
    RpcClient *rpc.Client
    NAME      string
}

func NewchainConfig(rpcUrl string, name string) *ChainClient {
    if rpcUrl == "" {
        fmt.Println("[FATAL] rpcUrl is empty")
        return nil
    }
    client, rpcClient, err := connectClient(rpcUrl)
    if err != nil {
        fmt.Printf("[FATAL] failed to connect: %v\n", err)
        return nil
    }
    fmt.Printf("[INFO] connected to %s (%s)\n", name, rpcUrl)
    return &ChainClient{Client: client, RpcClient: rpcClient, NAME: name}
}

func connectClient(rpcUrl string) (*ethclient.Client, *rpc.Client, error) {
    client, err := ethclient.Dial(rpcUrl)
    if err != nil {
        return nil, nil, pkg.WrapError("failed to dial ethclient", err)
    }
    rpcC, err := rpc.Dial(rpcUrl)
    if err != nil {
        return nil, nil, pkg.WrapError("failed to dial rpc", err)
    }
    return client, rpcC, nil
}

func (c *ChainClient) GetLatestBlock(ctx context.Context) (uint64, error) {
    latestBlock, err := c.Client.BlockNumber(ctx)
    if err != nil {
        return 0, pkg.WrapError("failed to get latest block", err)
    }
    return latestBlock, nil
}

func (c *ChainClient) GetBlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
    block, err := c.Client.BlockByNumber(ctx, number)
    if err != nil {
        return nil, pkg.WrapError("failed to get block by number", err)
    }
    return block, nil
}