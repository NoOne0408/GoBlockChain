package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

//定义区块
type Block struct {
	Timestamp int64 //时间戳
	//Data          []byte //交易数据
	Transactions  []*Transaction //交易的集合
	PrevBlockHash []byte         //上一块数据的hash
	Hash          []byte         //当前数据hash
	Nonce         int            //工作量证明
}

//创建一个区块
func NewBlock(transaction []*Transaction, preBlockHash []byte) *Block {
	//block是一个指针，取得对象初始化之后的地址
	block := &Block{Timestamp: time.Now().Unix(),
		Transactions:  transaction,
		PrevBlockHash: preBlockHash,
		Hash:          []byte{}}

	pow := NewProofOfWork(block) //挖矿附加这个区块
	nonce, hash := pow.Run()     //开始挖矿
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

//创建创世区块
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

//对象转化位二进制字节集
func (block *Block) Serialize() []byte {
	var result bytes.Buffer            //开辟内存，存放二进制字节集
	encoder := gob.NewEncoder(&result) //编码对象创建
	err := encoder.Encode(block)       //编码操作
	if err != nil {
		log.Panic(err) //处理错误
	}

	return result.Bytes() //返回字节
}

//读取文件，读到二进制字节集
func DeSerializeBlock(data []byte) *Block {
	var block Block                                  //用于存储字节转换的对象
	decoder := gob.NewDecoder(bytes.NewReader(data)) //解码
	err := decoder.Decode(&block)                    //尝试解码
	if err != nil {
		log.Panic(err) //错误处理
	}
	return &block
}

//对于交易实现哈希计算
func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]

}
