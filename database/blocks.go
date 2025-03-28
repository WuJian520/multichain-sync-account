package database

import (
	"errors"
	"math/big"

	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/dapplink-labs/multichain-sync-account/rpcclient"
)

type Blocks struct {
	Hash       common.Hash `gorm:"primaryKey;serializer:bytes"`
	ParentHash common.Hash `gorm:"serializer:bytes"`
	Number     *big.Int    `gorm:"serializer:u256"`
	Timestamp  uint64
}

func BlockHeaderFromHeader(header *types.Header) rpcclient.BlockHeader {
	return rpcclient.BlockHeader{
		Hash:       header.Hash(),
		ParentHash: header.ParentHash,
		Number:     header.Number,
		Timestamp:  header.Time,
	}
}

type BlocksView interface {
	LatestBlocks() (*rpcclient.BlockHeader, error)
	QueryBlocksByNumber(*big.Int) (*rpcclient.BlockHeader, error)
}

type BlocksDB interface {
	BlocksView

	StoreBlockss([]Blocks) error
	DeleteBlocksByNumber(blockHeader []Blocks) error
}

type blocksDB struct {
	gorm *gorm.DB
}

func NewBlocksDB(db *gorm.DB) BlocksDB {
	return &blocksDB{gorm: db}
}

func (db *blocksDB) StoreBlockss(headers []Blocks) error {
	result := db.gorm.CreateInBatches(&headers, len(headers))
	return result.Error
}

func (db *blocksDB) LatestBlocks() (*rpcclient.BlockHeader, error) {
	var header Blocks
	result := db.gorm.Order("number DESC").Take(&header)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return (*rpcclient.BlockHeader)(&header), nil
}

func (db *blocksDB) QueryBlocksByNumber(queryNumber *big.Int) (*rpcclient.BlockHeader, error) {
	var header Blocks
	result := db.gorm.Table("blocks").Where("number = ?", queryNumber.Uint64()).Take(&header)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, result.Error
	}
	return (*rpcclient.BlockHeader)(&header), nil
}

func (db *blocksDB) DeleteBlocksByNumber(blockHeader []Blocks) error {
	for _, v := range blockHeader {
		result := db.gorm.Table("blocks").Where("number = ?", v.Number.Uint64()).Delete(&Blocks{})
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
	}
	return nil
}
