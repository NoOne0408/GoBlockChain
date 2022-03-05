package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

const dbFile = "blockchain.db" //数据库名
const blockBucket = "blocks"   //名称

type BlockChain struct {
	tip []byte   //二进制数据
	db  *bolt.DB //数据库
}

type BlockChainIterator struct {
	currentHash []byte   //当前哈希
	db          *bolt.DB //数据库

}

//增加一个区块
func (block *BlockChain) AddBlock(data string) {
	var lastHash []byte //上一块哈希
	err := block.db.View(func(tx *bolt.Tx) error {
		block := tx.Bucket([]byte(blockBucket)) //取得数据
		lastHash = block.Get([]byte("1"))       //取得第一块
		return nil
	})
	if err != nil {
		log.Panic(err) //处理打开错误
	}
	newBlock := NewBlock(data, lastHash) //创建一个新的区块
	err = block.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err) //处理压入错误
		}
		err = bucket.Put([]byte("1"), newBlock.Hash)
		if err != nil {
			log.Panic(err) //处理压入错误
		}
		block.tip = newBlock.Hash

		return nil
	})

}

//迭代器
func (block *BlockChain) Iterator() *BlockChainIterator {
	bcit := &BlockChainIterator{block.tip, block.db}
	return bcit //根据区块链创建区块链迭代器
}

//获取下一个区块
func (it *BlockChainIterator) next() *Block {
	var block *Block
	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		encodedBlock := bucket.Get(it.currentHash) //抓去二进制数据
		block = DeSerializeBlock(encodedBlock)     //解码

		return nil
	})
	if err != nil {
		log.Panic(err) //处理压入错误
	}
	it.currentHash = block.PrevBlockHash //哈希赋值
	return block
}

//新建一个区块
func NewBlockChain() *BlockChain {
	var tip []byte                          //存储数据库的二进制数据
	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err) //处理打开错误
	}
	//处理数据更新
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //按照名称打开数据库表格
		if bucket == nil {
			fmt.Println("当前数据库没有区块链，创建新区块链！")

			genesis := NewGenesisBlock() //创建创世区块

			bucket, err := tx.CreateBucket([]byte(blockBucket)) //创建数据库表

			if err != nil {
				log.Panic(err) //处理创建错误
			}

			err = bucket.Put(genesis.Hash, genesis.Serialize()) //存入数据
			if err != nil {
				log.Panic(err) //处理存入错误
			}

			err = bucket.Put([]byte("1"), genesis.Hash) //存入数据
			if err != nil {
				log.Panic(err) //处理存入错误
			}
			tip = genesis.Hash //取得哈希
		} else {
			tip = bucket.Get([]byte("1"))
		}
		return nil
	},
	)
	if err != nil {
		log.Panic(err) //处理数据库更新错误
	}
	bc := BlockChain{tip, db}
	return &bc

}

//type BlockChain struct {
//	blocks []*Block //一个数组，存储block区块指针
//
//}
//
////增加一个区块
//func (blocks *BlockChain) AddBlock(data string) {
//	prevBlock := blocks.blocks[len(blocks.blocks)-1] //取出最后一个区块
//	newBlock := NewBlock(data, prevBlock.Hash)       //创建一个区块
//	blocks.blocks = append(blocks.blocks, newBlock)  //区块链尾部插入新区块
//
//}
//
////创建一个区块链
//func NewBlockchain() *BlockChain {
//	return &BlockChain{[]*Block{NewGenesisBlock()}}
//
//}
