package services

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ownafarm/ownafarm-backend/internal/config"
)

// Investment represents an investment from the smart contract
type OnchainInvestment struct {
	Amount     *big.Int
	TokenID    uint32
	InvestedAt uint32
	Claimed    bool
}

// BlockchainService defines the interface for blockchain operations
type BlockchainService interface {
	GetInvestmentCount(ctx context.Context, investor string) (uint64, error)
	GetInvestment(ctx context.Context, investor string, investmentId uint64) (*OnchainInvestment, error)
	GetInvoiceByTokenID(ctx context.Context, tokenId uint64) (*OnchainInvoice, error)
}

// OnchainInvoice represents an invoice from the smart contract
type OnchainInvoice struct {
	Farmer       common.Address
	TargetFund   *big.Int
	FundedAmount *big.Int
	YieldBps     uint16
	Duration     uint32
	CreatedAt    uint32
	Status       uint8 // 0: Pending, 1: Approved, 2: Rejected, 3: Funded, 4: Completed
	OfftakerId   [32]byte
}

type blockchainService struct {
	client     *ethclient.Client
	nftAddress common.Address
	abi        abi.ABI
}

// OwnaFarmNFT ABI (minimal for reading investments)
const OwnaFarmNFTABI = `[
	{
		"inputs": [{"internalType": "address", "name": "", "type": "address"}],
		"name": "investmentCount",
		"outputs": [{"internalType": "uint256", "name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "address", "name": "investor", "type": "address"},
			{"internalType": "uint256", "name": "investmentId", "type": "uint256"}
		],
		"name": "getInvestment",
		"outputs": [
			{
				"components": [
					{"internalType": "uint128", "name": "amount", "type": "uint128"},
					{"internalType": "uint32", "name": "tokenId", "type": "uint32"},
					{"internalType": "uint32", "name": "investedAt", "type": "uint32"},
					{"internalType": "bool", "name": "claimed", "type": "bool"}
				],
				"internalType": "struct OwnaFarmNFT.Investment",
				"name": "",
				"type": "tuple"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"internalType": "uint256", "name": "", "type": "uint256"}],
		"name": "invoices",
		"outputs": [
			{"internalType": "address", "name": "farmer", "type": "address"},
			{"internalType": "uint128", "name": "targetFund", "type": "uint128"},
			{"internalType": "uint128", "name": "fundedAmount", "type": "uint128"},
			{"internalType": "uint16", "name": "yieldBps", "type": "uint16"},
			{"internalType": "uint32", "name": "duration", "type": "uint32"},
			{"internalType": "uint32", "name": "createdAt", "type": "uint32"},
			{"internalType": "uint8", "name": "status", "type": "uint8"},
			{"internalType": "bytes32", "name": "offtakerId", "type": "bytes32"}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`

// NewBlockchainService creates a new BlockchainService instance
func NewBlockchainService(cfg *config.BlockchainConfig) (BlockchainService, error) {
	client, err := ethclient.Dial(cfg.MantleRPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Mantle RPC: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(OwnaFarmNFTABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse OwnaFarmNFT ABI: %w", err)
	}

	return &blockchainService{
		client:     client,
		nftAddress: common.HexToAddress(cfg.OwnaFarmNFTAddr),
		abi:        parsedABI,
	}, nil
}

// GetInvestmentCount returns the number of investments for an investor
func (s *blockchainService) GetInvestmentCount(ctx context.Context, investor string) (uint64, error) {
	investorAddr := common.HexToAddress(investor)

	data, err := s.abi.Pack("investmentCount", investorAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to pack investmentCount call: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &s.nftAddress,
		Data: data,
	}

	result, err := s.client.CallContract(ctx, msg, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call investmentCount: %w", err)
	}

	var count *big.Int
	err = s.abi.UnpackIntoInterface(&count, "investmentCount", result)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack investmentCount result: %w", err)
	}

	return count.Uint64(), nil
}

// GetInvestment returns an investment by investor address and investment ID
func (s *blockchainService) GetInvestment(ctx context.Context, investor string, investmentId uint64) (*OnchainInvestment, error) {
	investorAddr := common.HexToAddress(investor)
	investmentIdBig := new(big.Int).SetUint64(investmentId)

	data, err := s.abi.Pack("getInvestment", investorAddr, investmentIdBig)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getInvestment call: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &s.nftAddress,
		Data: data,
	}

	result, err := s.client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call getInvestment: %w", err)
	}

	// Unpack the result
	unpacked, err := s.abi.Unpack("getInvestment", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getInvestment result: %w", err)
	}

	// The result is a struct, use reflection to extract fields
	if len(unpacked) == 0 {
		return nil, fmt.Errorf("empty result from getInvestment")
	}

	// Use reflect to access fields from the anonymous struct
	invValue := reflect.ValueOf(unpacked[0])
	if invValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unexpected type from getInvestment: %T", unpacked[0])
	}

	// Extract fields by name
	amountField := invValue.FieldByName("Amount")
	tokenIdField := invValue.FieldByName("TokenId")
	investedAtField := invValue.FieldByName("InvestedAt")
	claimedField := invValue.FieldByName("Claimed")

	var amount *big.Int
	if amountField.IsValid() && !amountField.IsNil() {
		amount = amountField.Interface().(*big.Int)
	}

	var tokenId uint32
	if tokenIdField.IsValid() {
		tokenId = uint32(tokenIdField.Uint())
	}

	var investedAt uint32
	if investedAtField.IsValid() {
		investedAt = uint32(investedAtField.Uint())
	}

	var claimed bool
	if claimedField.IsValid() {
		claimed = claimedField.Bool()
	}

	return &OnchainInvestment{
		Amount:     amount,
		TokenID:    tokenId,
		InvestedAt: investedAt,
		Claimed:    claimed,
	}, nil
}

// GetInvoiceByTokenID returns an invoice by token ID from the smart contract
func (s *blockchainService) GetInvoiceByTokenID(ctx context.Context, tokenId uint64) (*OnchainInvoice, error) {
	tokenIdBig := new(big.Int).SetUint64(tokenId)

	data, err := s.abi.Pack("invoices", tokenIdBig)
	if err != nil {
		return nil, fmt.Errorf("failed to pack invoices call: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &s.nftAddress,
		Data: data,
	}

	result, err := s.client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call invoices: %w", err)
	}

	unpacked, err := s.abi.Unpack("invoices", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack invoices result: %w", err)
	}

	if len(unpacked) < 8 {
		return nil, fmt.Errorf("invalid invoice data")
	}

	return &OnchainInvoice{
		Farmer:       unpacked[0].(common.Address),
		TargetFund:   unpacked[1].(*big.Int),
		FundedAmount: unpacked[2].(*big.Int),
		YieldBps:     unpacked[3].(uint16),
		Duration:     unpacked[4].(uint32),
		CreatedAt:    unpacked[5].(uint32),
		Status:       unpacked[6].(uint8),
		OfftakerId:   unpacked[7].([32]byte),
	}, nil
}
