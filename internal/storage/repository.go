package storage

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"time"
)

type Repository struct{
	Db *sql.DB
}

type DbRepository interface{
	SaveBlock(ctx context.Context,block *BlockRecord)error
	SaveDecodedCall(ctx context.Context,call *DecodedCallRecord)error
	SaveDecodedCallBatch(ctx context.Context, calls []*DecodedCallRecord)error
	GetLastProcessedBlock(ctx context.Context)(uint64,error)
	IsBlockProcessed(ctx context.Context,blockNumber uint64)(bool,error)
}


func (r *Repository) Migrate() error{

	queries :=[]string{
		`CREATE TABLE IF NOT EXISTS blocks(
		id      SERIAL PRIMARY KEY,
		block_number  BIGINT UNIQUE NOT NULL,
		total_txs   INT DEFAULT 0,
		skipped_txs  INT DEFAULT 0,
		event_count  INT DEFAULT 0,
		call_count   INT DEFAULT 0,
		processed_at  TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS decoded_calls(
		id            SERIAL PRIMARY KEY,
			block_number  BIGINT NOT NULL,
			tx_hash       VARCHAR(66) NOT NULL,
			method        VARCHAR(20) NOT NULL,
			caller_addr   VARCHAR(42) NOT NULL,
			contract_addr VARCHAR(42) NOT NULL,
			from_addr     VARCHAR(42) NOT NULL,
			to_addr       VARCHAR(42) NOT NULL,
			amount        NUMERIC(78, 0) NOT NULL DEFAULT 0,
			tx_type       INT NOT NULL DEFAULT 0,
			created_at    TIMESTAMPTZ DEFAULT NOW(),
			FOREIGN KEY (block_number) REFERENCES blocks(block_number)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_decoded_calls_block   ON decoded_calls(block_number)`,
		`CREATE INDEX IF NOT EXISTS idx_decoded_calls_method  ON decoded_calls(method)`,
		`CREATE INDEX IF NOT EXISTS idx_decoded_calls_from    ON decoded_calls(from_addr)`,
		`CREATE INDEX IF NOT EXISTS idx_decoded_calls_to      ON decoded_calls(to_addr)`,
		`CREATE INDEX IF NOT EXISTS idx_decoded_calls_tx      ON decoded_calls(tx_hash)`,
	}

	for _,q := range queries{
		if _,err:= r.Db.Exec(q); err !=nil{
			return fmt.Errorf("Migrate: %w",err)
		}
	}
fmt.Println("DB migrations complete")
return nil
}

func (r *Repository) SaveBlock(ctx context.Context, block *BlockRecord) error {
    query := `
        INSERT INTO blocks (block_number, total_txs, skipped_txs, event_count, call_count, processed_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (block_number) DO NOTHING`

    _, err := r.Db.ExecContext(ctx, query,
        block.BlockNumber, 
        block.TotalTxs,    
        block.SkippedTxs,  
        block.EventCount,  
        block.CallCount,   
        time.Now(),        
    )
    return err
}
func (r *Repository) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	var blockNumber uint64
	err := r.Db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(block_number), 0) FROM blocks`,
	).Scan(&blockNumber)
	return blockNumber, err
}

func (r *Repository) IsBlockProcessed(ctx context.Context, blockNumber uint64) (bool, error) {
	var exists bool
	err := r.Db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM blocks WHERE block_number = $1)`,
		blockNumber,
	).Scan(&exists)
	return exists, err
}

func (r *Repository) SaveDecodedCall(ctx context.Context, call *DecodedCallRecord) error {
	_, err := r.Db.ExecContext(ctx, `
		INSERT INTO decoded_calls
			(block_number, tx_hash, method, caller_addr, contract_addr, from_addr, to_addr, amount, tx_type)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`,
		call.BlockNumber,
		call.TxHash.Hex(),
		call.Method,
		call.CallerAddr.Hex(),
		call.ContractAddr.Hex(),
		call.From.Hex(),
		call.To.Hex(),
		amountToString(&call.Amount),
		call.TxType,
	)
	return err
}

func amountToString(amount *big.Int) string {
	if amount == nil {
		return "0"
	}
	return amount.String()
}

func (r *Repository) SaveDecodedCallBatch(ctx context.Context, calls []*DecodedCallRecord) error {
	if len(calls) == 0 {
		return nil
	}

	tx, err := r.Db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO decoded_calls
			(block_number, tx_hash, method, caller_addr, contract_addr, from_addr, to_addr, amount, tx_type)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, call := range calls {
		_, err := stmt.ExecContext(ctx,
			call.BlockNumber,
			call.TxHash.Hex(),
			call.Method,
			call.CallerAddr.Hex(),
			call.ContractAddr.Hex(),
			call.From.Hex(),
			call.To.Hex(),
			amountToString(&call.Amount),
			call.TxType,
		)
		if err != nil {
			return fmt.Errorf("insert decoded call: %w", err)
		}
	}

	return tx.Commit()
}