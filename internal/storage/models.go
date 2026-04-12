package storage

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type BlockRecord struct{
	ID    uint64  `db:"id"`
	BlockNumber  uint64 `db:"block_number"`
	TotalTxs   int  `db:"total_txs"`
	SkippedTxs  int `db:"skipped_txs"`
	EventCount   int `db:"event_count"`
	CallCount    int `db:"call_count"`
	ProcessedAt  time.Time `db:"processed_at"`
}


type DecodedCallRecord struct{
	ID     uint64 `db:"id"`
	BlockNumber  uint64 `db:"block_number"`
	TxHash     common.Hash `db:"tx_hash"`
	Method   string `db:"method"`
	CallerAddr  common.Address `db:"caller_addr"`
	ContractAddr  common.Address  `db:"contract_addr"`
	From         common.Address `db:"from"`
	To           common.Address `db:"to"`
	Amount       big.Int  `db:"amount"`
	TxType      uint64 `db:"tx_type"`
	CreatedAt  time.Time  `db:"created_at"`
}

