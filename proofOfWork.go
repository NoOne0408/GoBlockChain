package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64 //最大的64位整数
)

const targetBits = 16 //对比位数，因为hash值为16进制，每一个数字代表4位，所以最终hash的前24位，即前6个数字都为0
type ProofOfWork struct {
	block  *Block   //区块
	target *big.Int //存储计算哈希对比的特定整数
}

//创建一个工作量证明的挖矿对象
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits)) //数据转换
	pow := &ProofOfWork{block, target}       //创建对象
	return pow
}

//准备数据进行挖矿计算
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,       //上一块哈希
			pow.block.HashTransactions(),  //当前数据
			IntToHex(pow.block.Timestamp), //时间16进制
			IntToHex(int64(targetBits)),   //位数16进制
			IntToHex(int64((nonce))),      //保存工作量的nonce
		}, []byte{},
	)
	return data
}

//挖矿执行
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	//fmt.Printf("当前挖矿计算的数据%s", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)     //准备数据
		hash = sha256.Sum256(data)         //计算出哈希
		fmt.Printf("\r%x", hash)           //打印显示哈希
		hashInt.SetBytes(hash[:])          //获取要对比的数据
		if hashInt.Cmp(pow.target) == -1 { //挖矿的校验:如果hashInt小于pow.target返回-1，即满足前若干位数值全为0
			break
		} else {
			nonce++
		}

	}
	fmt.Println()
	return nonce, hash[:] //nonce相当于答案，hash代表当前哈希
}

//校验挖矿是否真的成功
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce)   //准备数据
	hash := sha256.Sum256(data)                //计算出哈希
	hashInt.SetBytes(hash[:])                  //获取要对比的数据
	isValid := (hashInt.Cmp(pow.target) == -1) //校验:如果hashInt小于pow.target返回-1，即满足前若干位数值全为0
	return isValid
}
