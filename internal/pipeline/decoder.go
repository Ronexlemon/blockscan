package pipeline


import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)


type TransferArgs struct {
	To     common.Address
	Amount *big.Int
}

type TransferFromArgs struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
}

type ApproveArgs struct {
	Spender common.Address
	Amount  *big.Int
}

type MintArgs struct {
	To     common.Address
	Amount *big.Int
}


type DecodedCall struct {
	Method       string
	TxHash       common.Hash
	From         common.Address 
	ContractAddr common.Address 
	Args         interface{}    
}



func DecodeMethodCall(call MethodCall) (*DecodedCall, error) {
	selHex := hex.EncodeToString(call.Selector[:]) // the first 4 bytes
	//fmt.Println("decoder Starting",call.From)

	switch selHex {
	case "a9059cbb":
		args, err := decodeTransfer(call.Input)
		if err != nil {
			return nil, err
		}
		return &DecodedCall{
			Method:       "transfer",
			TxHash:       call.TxHash,
			From:         call.From,
			ContractAddr: call.To,
			Args:         args,
		}, nil

	case "23b872dd":
		args, err := decodeTransferFrom(call.Input)
		if err != nil {
			return nil, err
		}
		return &DecodedCall{
			Method:       "transferFrom",
			TxHash:       call.TxHash,
			From:         call.From,
			ContractAddr: call.To,
			Args:         args,
		}, nil

	case "095ea7b3":
		args, err := decodeApprove(call.Input)
		if err != nil {
			return nil, err
		}
		return &DecodedCall{
			Method:       "approve",
			TxHash:       call.TxHash,
			From:         call.From,
			ContractAddr: call.To,
			Args:         args,
		}, nil

	case "40c10f19":
		args, err := decodeMint(call.Input)
		if err != nil {
			return nil, err
		}
		return &DecodedCall{
			Method:       "mint",
			TxHash:       call.TxHash,
			From:         call.From,
			ContractAddr: call.To,
			Args:         args,
		}, nil

	default:
		return nil, fmt.Errorf("unknown selector: %s", selHex)
	}
}

func readAddress(slot []byte) common.Address {
	return common.BytesToAddress(slot[12:32])
}


func readUint256(slot []byte) *big.Int {
	return new(big.Int).SetBytes(slot[0:32])
}

// checkLength validates the input has the expected number of 32-byte slots
func checkLength(input []byte, slots int) error {
	expected := 4 + slots*32
	if len(input) < expected {
		return fmt.Errorf("input too short: got %d bytes, want %d", len(input), expected)
	}
	return nil
}

// -----------------------------------------------------------------------
// transfer(address to, uint256 amount)
// slot 0 = to
// slot 1 = amount
// -----------------------------------------------------------------------

func decodeTransfer(input []byte) (*TransferArgs, error) {
	if err := checkLength(input, 2); err != nil {
		return nil, fmt.Errorf("transfer: %w", err)
	}
	data := input[4:] // strip selector
	return &TransferArgs{
		To:     readAddress(data[0:32]),
		Amount: readUint256(data[32:64]),
	}, nil
}

// -----------------------------------------------------------------------
// transferFrom(address from, address to, uint256 amount)
// slot 0 = from
// slot 1 = to
// slot 2 = amount
// -----------------------------------------------------------------------

func decodeTransferFrom(input []byte) (*TransferFromArgs, error) {
	if err := checkLength(input, 3); err != nil {
		return nil, fmt.Errorf("transferFrom: %w", err)
	}
	data := input[4:]
	return &TransferFromArgs{
		From:   readAddress(data[0:32]),
		To:     readAddress(data[32:64]),
		Amount: readUint256(data[64:96]),
	}, nil
}

// -----------------------------------------------------------------------
// approve(address spender, uint256 amount)
// slot 0 = spender
// slot 1 = amount
// -----------------------------------------------------------------------

func decodeApprove(input []byte) (*ApproveArgs, error) {
	if err := checkLength(input, 2); err != nil {
		return nil, fmt.Errorf("approve: %w", err)
	}
	data := input[4:]
	return &ApproveArgs{
		Spender: readAddress(data[0:32]),
		Amount:  readUint256(data[32:64]),
	}, nil
}

// -----------------------------------------------------------------------
// mint(address to, uint256 amount)
// slot 0 = to
// slot 1 = amount
// -----------------------------------------------------------------------

func decodeMint(input []byte) (*MintArgs, error) {
	if err := checkLength(input, 2); err != nil {
		return nil, fmt.Errorf("mint: %w", err)
	}
	data := input[4:]
	return &MintArgs{
		To:     readAddress(data[0:32]),
		Amount: readUint256(data[32:64]),
	}, nil
}

// -----------------------------------------------------------------------
// Pretty printer
// -----------------------------------------------------------------------

func (d *DecodedCall) String() string {
	base := fmt.Sprintf("tx=%s  contract=%s  caller=%s\n",
		d.TxHash.Hex()[:10],
		d.ContractAddr.Hex(),
		d.From.Hex(),
	)

	switch args := d.Args.(type) {
	case *TransferArgs:
		return base + fmt.Sprintf("  transfer → to=%s  amount=%s",
			args.To.Hex(), args.Amount.String())

	case *TransferFromArgs:
		return base + fmt.Sprintf("  transferFrom → from=%s  to=%s  amount=%s",
			args.From.Hex(), args.To.Hex(), args.Amount.String())

	case *ApproveArgs:
		return base + fmt.Sprintf("  approve → spender=%s  amount=%s",
			args.Spender.Hex(), args.Amount.String())

	case *MintArgs:
		return base + fmt.Sprintf("  mint → to=%s  amount=%s",
			args.To.Hex(), args.Amount.String())

	default:
		return base + "  (unknown args)"
	}
}