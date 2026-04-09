package pipeline

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	TransferTopic = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163f4f5b3cd39f0c5b3f6f1d4")
	ApprovalTopic = common.HexToHash("0x8c5be1e5ebec7d5bd14f714f3e3d65b70a5a3e5a4c238e5b236e4d5e4c1d49a7")
	MintTopic     = common.HexToHash("0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885")
)

var knownSelectors = map[[4]byte]string{
	mustSel("a9059cbb"): "transfer",
	mustSel("23b872dd"): "transferFrom",
	mustSel("095ea7b3"): "approve",
	mustSel("40c10f19"): "mint",
}

func mustSel(s string) [4]byte {
	b, _ := hex.DecodeString(s)
	var out [4]byte
	copy(out[:], b)
	return out
}



type rawBlock struct {
	Hash         common.Hash   `json:"hash"`
	Transactions []common.Hash `json:"transactions"`
}

type rawTx struct {
	Hash     common.Hash     `json:"hash"`
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Input    hexutil.Bytes   `json:"input"`
	Type     hexutil.Uint64  `json:"type"`
	BlockNum *hexutil.Big    `json:"blockNumber"`
}


type MethodCall struct {
	TxHash   common.Hash
	From     common.Address
	To       common.Address
	Selector [4]byte
	Name     string
	Input    []byte
	TxType   uint64
}

type BlockResult struct {
	BlockNumber uint64
	EventLogs   []types.Log
	MethodCalls []MethodCall
	TotalTxs    int
	SkippedTxs  int
}


type BlockProcessor struct {
	Client    *ethclient.Client
	RpcClient *rpc.Client
	RpcURL    string
	Sem       chan struct{} // caps concurrent RPC calls — init with make(chan struct{}, N)
}



func (p *BlockProcessor) ProcessBlock(ctx context.Context, header *types.Header) (*BlockResult, error) {
	result := &BlockResult{BlockNumber: header.Number.Uint64()}

	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	if p.RpcClient == nil {
		return result, fmt.Errorf("RpcClient is nil — set BlockProcessor.RpcClient before use")
	}
	if p.Sem == nil {
		return result, fmt.Errorf("Sem is nil — init with make(chan struct{}, 5)")
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		logs, err := p.fetchEventLogs(ctx, header.Number)
		if err != nil {
			errCh <- fmt.Errorf("event logs: %w", err)
			return
		}
		result.EventLogs = logs
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		calls, total, skipped, err := p.fetchMethodCalls(ctx, header.Number)
		if err != nil {
			errCh <- fmt.Errorf("method calls: %w", err)
			return
		}
		result.MethodCalls = calls
		result.TotalTxs = total
		result.SkippedTxs = skipped
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			fmt.Printf("[WARN] block %d: %v\n", result.BlockNumber, err)
		}
	}

	if len(result.EventLogs) > 0 || len(result.MethodCalls) > 0 {
		fmt.Printf("[Block %d] events=%d calls=%d (txs=%d skipped=%d)\n",
			result.BlockNumber,
			len(result.EventLogs),
			len(result.MethodCalls),
			result.TotalTxs,
			result.SkippedTxs,
		)
		for _, call := range result.MethodCalls {
			fmt.Printf("  [%s] from=%s to=%s\n", call.Name, call.From.Hex(), call.To.Hex())
		}
	} else {
		fmt.Printf("[Block %d] no ERC20 activity (txs=%d skipped=%d)\n",
			result.BlockNumber, result.TotalTxs, result.SkippedTxs)
	}

	return result, nil
}



func (p *BlockProcessor) fetchEventLogs(ctx context.Context, blockNum *big.Int) ([]types.Log, error) {
	logs, err := p.Client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: blockNum,
		ToBlock:   blockNum,
		Topics:    [][]common.Hash{{TransferTopic, ApprovalTopic, MintTopic}},
	})
	if err != nil {
		fmt.Printf("[WARN] FilterLogs failed (%v) — using per-tx fallback\n", err)
		return p.fetchEventLogsFallback(ctx, blockNum)
	}
	return logs, nil
}

func (p *BlockProcessor) fetchEventLogsFallback(ctx context.Context, blockNum *big.Int) ([]types.Log, error) {
	hashes, err := p.fetchTxHashes(ctx, blockNum)
	if err != nil {
		return nil, err
	}
	if len(hashes) == 0 {
		return nil, nil
	}

	topicSet := map[common.Hash]struct{}{
		TransferTopic: {},
		ApprovalTopic: {},
		MintTopic:     {},
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	var results []types.Log

	for _, h := range hashes {
		wg.Add(1)
		go func(hash common.Hash) {
			defer wg.Done()

			// Semaphore
			p.Sem <- struct{}{}
			defer func() { <-p.Sem }()

			receipt, err := p.Client.TransactionReceipt(ctx, hash)
			if err != nil {
				return
			}
			for _, lg := range receipt.Logs {
				if len(lg.Topics) > 0 {
					if _, ok := topicSet[lg.Topics[0]]; ok {
						mu.Lock()
						results = append(results, *lg)
						mu.Unlock()
					}
				}
			}
		}(h)
	}

	wg.Wait()
	return results, nil
}


var systemSelectors = map[[4]byte]bool{
	mustSel("3db6be2b"): true, // CELO / OP system
	mustSel("d764ad0b"): true, // relayMessage
	mustSel("1635f5fd"): true, // depositTransaction
}

// System addresses to skip
var systemAddresses = map[common.Address]bool{
	common.HexToAddress("0x0000000000000000000000000000000000000000"): true,
	common.HexToAddress("0x4200000000000000000000000000000000000015"): true,
	common.HexToAddress("0x4200000000000000000000000000000000000016"): true,
	common.HexToAddress("0xDeadDeAddeAddEAddeadDEaDDEAdDeaDDeAD0000"): true,
}

func (p *BlockProcessor) fetchMethodCalls(ctx context.Context, blockNum *big.Int) ([]MethodCall, int, int, error) {
	hashes, err := p.fetchTxHashes(ctx, blockNum)
	if err != nil {
		return nil, 0, 0, err
	}
	if len(hashes) == 0 {
		return nil, 0, 0, nil
	}

	total := len(hashes)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var results []MethodCall
	skipped := 0

	for _, h := range hashes {
		wg.Add(1)
		go func(hash common.Hash) {
			defer wg.Done()

			tx, err := p.fetchRawTx(ctx, hash)
			if err != nil {
				fmt.Printf("[SKIP] tx %s: %v\n", hash.Hex()[:10], err)
				return
			}

			if tx == nil || tx.To == nil || len(tx.Input) < 4 {
				return // contract creation or bare ETH transfer
			}

			// Skip system addresses
			if systemAddresses[*tx.To] || systemAddresses[tx.From] {
				mu.Lock()
				skipped++
				mu.Unlock()
				return
			}

			var sel [4]byte
			copy(sel[:], tx.Input[:4])

			// Skip system selectors
			if systemSelectors[sel] {
				mu.Lock()
				skipped++
				mu.Unlock()
				return
			}

			name, known := knownSelectors[sel]
			if !known {
				return
			}

			mu.Lock()
			results = append(results, MethodCall{
				TxHash:   tx.Hash,
				From:     tx.From,
				To:       *tx.To,
				Selector: sel,
				Name:     name,
				Input:    tx.Input,
				TxType:   uint64(tx.Type),
			})
			mu.Unlock()
		}(h)
	}

	wg.Wait()
	return results, total, skipped, nil
}

func (p *BlockProcessor) fetchTxHashes(ctx context.Context, blockNum *big.Int) ([]common.Hash, error) {
	if p.RpcClient == nil {
		return nil, fmt.Errorf("RpcClient is nil")
	}
	var raw rawBlock
	err := p.RpcClient.CallContext(
		ctx,
		&raw,
		"eth_getBlockByNumber",
		hexutil.EncodeBig(blockNum),
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber: %w", err)
	}
	return raw.Transactions, nil
}


func (p *BlockProcessor) fetchRawTx(ctx context.Context, hash common.Hash) (*rawTx, error) {
	p.Sem <- struct{}{}
	defer func() { <-p.Sem }()

	var result json.RawMessage
	err := p.RpcClient.CallContext(ctx, &result, "eth_getTransactionByHash", hash)
	if err != nil {
		return nil, err
	}
	if result == nil || string(result) == "null" {
		return nil, fmt.Errorf("tx not found")
	}
	var tx rawTx
	if err := json.Unmarshal(result, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}