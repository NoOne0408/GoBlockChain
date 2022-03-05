package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

//定义区块
type Block struct {
	Timestamp     int64  //时间戳
	Data          []byte //交易数据
	PrevBlockHash []byte //上一块数据的hash
	Hash          []byte //当前数据hash
	Nonce         int    //工作量证明
}

////设置结构体对象哈希
//func (block *Block) SetHash() {
//	//处理当前时间，转化为10进制字符串，再转为字节
//	timestamp := []byte(strconv.FormatInt(block.Timestamp, 10))
//	//叠加要哈希的数据
//	headers := bytes.Join([][]byte{block.PrevBlockHash, block.Data, timestamp}, []byte{})
//	hash := sha256.Sum256(headers)
//	//设置hash
//	block.Hash = hash[:]
//}

//创建一个区块
func NewBlock(data string, preBlockHash []byte) *Block {
	//block是一个指针，取得对象初始化之后的地址
	block := &Block{Timestamp: time.Now().Unix(), Data: []byte(data), PrevBlockHash: preBlockHash, Hash: []byte{}}
	pow := NewProofOfWork(block) //挖矿附加这个区块
	nonce, hash := pow.Run()     //开始挖矿
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

//创建创世区块
func NewGenesisBlock() *Block {
	return NewBlock("xxx的区块链", []byte{})
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
