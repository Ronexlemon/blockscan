package storage


import (
	"math/big"

	"github.com/ronexlemon/blockscan/internal/pipeline"
)

// -----------------------------------------------------------------------
// Maps pipeline results → DB records
// -----------------------------------------------------------------------

func BlockResultToRecords(result *pipeline.BlockResult) (
	block *BlockRecord,
	calls []*DecodedCallRecord,
) {
	block = &BlockRecord{
		BlockNumber: result.BlockNumber,
		TotalTxs:    result.TotalTxs,
		SkippedTxs:  result.SkippedTxs,
		EventCount:  len(result.EventLogs),
		CallCount:   len(result.DecodedCalls),
	}

	
	for _, d := range result.DecodedCalls {
		rec := &DecodedCallRecord{
			BlockNumber:  result.BlockNumber,
			TxHash:       d.TxHash,
			Method:       d.Method,
			CallerAddr:   d.From,
			ContractAddr: d.ContractAddr,
			TxType:       0,
		}

		switch args := d.Args.(type) {
		case *pipeline.TransferArgs:
			rec.From = d.From
			rec.To = args.To
			rec.Amount = *args.Amount

		case *pipeline.TransferFromArgs:
			rec.From = args.From
			rec.To = args.To
			rec.Amount = *args.Amount

		case *pipeline.ApproveArgs:
			rec.From = d.From
			rec.To = args.Spender
			rec.Amount = *args.Amount

		case *pipeline.MintArgs:
			rec.From = d.From
			rec.To = args.To
			rec.Amount = *args.Amount

		default:
			rec.Amount = *big.NewInt(0)
		}

		calls = append(calls, rec)
	}

	

	return block, calls
}