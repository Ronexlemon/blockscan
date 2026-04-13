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

type Pagination struct{
	Page int
	PerPage int
}

func (p *Pagination) normalize(){
	if p.Page < 1{
		p.Page =1
	}
	if p.PerPage < 20{
		p.PerPage =20
	}
	if p.PerPage >100{
		p.PerPage =100
	}
}

func (p *Pagination) offset()int{
	return (p.Page -1)* p.PerPage
}

type PaginatedResult[T any] struct{
	Data  []T
	Total int
	Page  int
	PerPage int
	TotalPages int
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


//reads

func (r *Repository) GetLatestBlock(ctx context.Context)(*BlockRecord,error){
	query:= `SELECT id,block_number,total_txs,skipped_txs,event_count,call_count,processed_at
	         FROM blocks
			 ORDER BY block_number DESC
			 LIMIT 1`
	row := r.Db.QueryRowContext(ctx,query)
	var b BlockRecord
	err:= row.Scan(
		&b.ID,
		&b.BlockNumber,
		&b.TotalTxs,
		&b.SkippedTxs,
		&b.EventCount,
		&b.CallCount,
		&b.ProcessedAt,
	)

	if err !=nil{
		return nil,fmt.Errorf("get latest block: %w",err)
	}
	return &b,nil
}

func (r *Repository) GetBlockByNumber(ctx context.Context,blockNumber uint64)(*BlockRecord,error){
	query:= `SELECT id,block_number,total_txs,skipped_txs,event_count,call_count,processed_at
	         FROM blocks
			 WHERE block_number = $1`
	row:= r.Db.QueryRowContext(ctx,query,blockNumber)

	var b BlockRecord

	err := row.Scan(
		&b.ID,
		&b.BlockNumber,
		&b.TotalTxs,
		&b.SkippedTxs,
		&b.EventCount,
		&b.CallCount,
		&b.ProcessedAt,
	)
	if err !=nil{
		return nil, fmt.Errorf("get block %d: %w",blockNumber,err)

	}
	return &b,nil
}

func (r *Repository) GetTransactionsByBlock(ctx context.Context,blockNumber uint64)([]DecodedCallRecord,error){
	query:= `SELECT id,block_number,tx_hash,method,caller_addr,contract_addr,from_addr,to_addr,amount,tx_type,created_at
	         FROM decoded_calls
			 WHERE block_number =$1
			 ORDER BY id ASC`
	rows,err := r.Db.QueryContext(ctx,query,blockNumber)
	if err !=nil{
		return nil,fmt.Errorf("get txs by block %d: %w",blockNumber,err)

	}
	defer rows.Close()
	return scanDecodedCalls(rows)
}

func (r *Repository) GetLatestTransactions(ctx context.Context,p Pagination)(*PaginatedResult[DecodedCallRecord],error){
p.normalize()

var total int

err := r.Db.QueryRowContext(ctx,`SELECT COUNT(*) FROM decoded_calls`).Scan(&total)
if err !=nil{
	return nil,fmt.Errorf("count transactions: %w",err)
}
query:= `SELECT id, block_number,tx_hash,method,caller_addr,contract_addr,from_addr,to_addr,amount,tx_type,created_at
         FROM decoded_calls
		 ORDER BY block_number DESC,id DESC
		 LIMIT $1 OFFSET $2`
rows,err := r.Db.QueryContext(ctx,query,p.PerPage,p.offset())
if err !=nil{
	return nil,fmt.Errorf("get latest transactions")
}
defer rows.Close()
data, err := scanDecodedCalls(rows)
if err !=nil{
	return nil,err
}
return &PaginatedResult[DecodedCallRecord]{
	Data: data,
	Total: total,
	Page: p.Page,
	PerPage: p.PerPage,
	TotalPages: totalPages(total,p.PerPage),
},nil
}

func (r *Repository) GetLatestBlocks(ctx context.Context, p Pagination) (*PaginatedResult[BlockRecord], error) {
	p.normalize()

	var total int
	err := r.Db.QueryRowContext(ctx, `SELECT COUNT(*) FROM blocks`).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count blocks: %w", err)
	}

	rows, err := r.Db.QueryContext(ctx, `
		SELECT id, block_number, total_txs, skipped_txs, event_count, call_count, processed_at
		FROM blocks
		ORDER BY block_number DESC
		LIMIT $1 OFFSET $2
	`, p.PerPage, p.offset())
	if err != nil {
		return nil, fmt.Errorf("get latest blocks: %w", err)
	}
	defer rows.Close()

	data, err := scanBlocks(rows)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[BlockRecord]{
		Data:       data,
		Total:      total,
		Page:       p.Page,
		PerPage:    p.PerPage,
		TotalPages: totalPages(total, p.PerPage),
	}, nil
}

func (r *Repository) GetBlockTransactionsPaginated(ctx context.Context, blockNumber uint64, p Pagination) (*PaginatedResult[DecodedCallRecord], error) {
	p.normalize()

	var total int
	err := r.Db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM decoded_calls WHERE block_number = $1`, blockNumber,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count block txs: %w", err)
	}

	rows, err := r.Db.QueryContext(ctx, `
		SELECT id, block_number, tx_hash, method, caller_addr, contract_addr,
		       from_addr, to_addr, amount, tx_type, created_at
		FROM decoded_calls
		WHERE block_number = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3
	`, blockNumber, p.PerPage, p.offset())
	if err != nil {
		return nil, fmt.Errorf("get block txs paginated: %w", err)
	}
	defer rows.Close()

	data, err := scanDecodedCalls(rows)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[DecodedCallRecord]{
		Data:       data,
		Total:      total,
		Page:       p.Page,
		PerPage:    p.PerPage,
		TotalPages: totalPages(total, p.PerPage),
	}, nil
}

func (r *Repository) GetTransactionsByAddress(ctx context.Context, address string, p Pagination) (*PaginatedResult[DecodedCallRecord], error) {
	p.normalize()

	var total int
	err := r.Db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM decoded_calls
		WHERE from = $1 OR to = $1 OR caller_addr = $1
	`, address).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count address txs: %w", err)
	}

	rows, err := r.Db.QueryContext(ctx, `
		SELECT id, block_number, tx_hash, method, caller_addr, contract_addr,
		       from_addr, to_addr, amount, tx_type, created_at
		FROM decoded_calls
		WHERE from_addr = $1 OR to_addr = $1 OR caller_addr = $1
		ORDER BY block_number DESC
		LIMIT $2 OFFSET $3
	`, address, p.PerPage, p.offset())
	if err != nil {
		return nil, fmt.Errorf("get address txs: %w", err)
	}
	defer rows.Close()

	data, err := scanDecodedCalls(rows)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[DecodedCallRecord]{
		Data:       data,
		Total:      total,
		Page:       p.Page,
		PerPage:    p.PerPage,
		TotalPages: totalPages(total, p.PerPage),
	}, nil
}

//helpers

func scanDecodedCalls(rows interface{Scan(...any) error;Next() bool; Err() error})([]DecodedCallRecord,error){
	var results []DecodedCallRecord

	for rows.Next(){
		var c DecodedCallRecord
		var amount string
		var createdAt time.Time

		err := rows.Scan(
			&c.ID,
			&c.BlockNumber,
			&c.TxHash,
			&c.Method,
			&c.CallerAddr,
			&c.ContractAddr,
			&c.From,
			&c.To,
			&amount,
			&c.TxType,
			&createdAt,

		)
		if err !=nil{
			return nil,fmt.Errorf("scan decoded call: %w",err)
		}
		c.Amount = *parseBigInt(amount)
		results = append(results, c)
	}
	return results,rows.Err()
}

func scanBlocks(rows interface{Scan(...any) error;Next() bool;Err() error})([]BlockRecord,error){
	var results []BlockRecord

	for rows.Next(){
		var b BlockRecord
		err:= rows.Scan(
			&b.ID,
			&b.BlockNumber,
			&b.TotalTxs,
			&b.SkippedTxs,
			&b.EventCount,
			&b.CallCount,
			&b.ProcessedAt,
		)
		if err !=nil{
			return nil,fmt.Errorf("scan block %w",err)
		}
		results = append(results, b)
	}
	return results,rows.Err()
}

func totalPages(total, perPage int)int{
	if perPage == 0{
		return 0
	}
	pages := total /perPage
	if total % perPage ==0{
		pages++
	}
	return pages
}

func parseBigInt(s string) *big.Int{
	n:= new(big.Int)
	n.SetString(s, 10)
	return n
}