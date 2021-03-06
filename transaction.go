package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 10 //矿工挖矿给予的奖励

//输入
type TXInput struct {
	Txid      []byte //存储交易id
	Vout      int    //保存交易中的一个output索引
	ScriptSig string //保存一个任意定义的钱包地址
}

//检查地址是否启动事务
func (input *TXInput) CanUnlockOutPutWith(unlockingData string) bool {
	return input.ScriptSig == unlockingData
}

//输出
type TXOutput struct {
	Value        int    //保存币的数量
	ScriptPubkey string //输出脚本，使用P2PK方式
}

//检查是否可以解锁输出
func (output *TXOutput) CanBeUnlockedOutPutWith(unlockingData string) bool {
	return output.ScriptPubkey == unlockingData
}

//交易类
type Transaction struct {
	ID   []byte
	Vin  []TXInput  //输入列表，币的来源
	Vout []TXOutput //输出列表，币的去向
}

//检查交易事务是否为coinbase,挖矿得来的奖励币
func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1

}

//设置交易ID
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer        //开辟内存
	var hash [32]byte               //哈希数组
	enc := gob.NewEncoder(&encoded) //解码对象
	err := enc.Encode(tx)           //解码
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes()) //计算哈希
	tx.ID = hash[:]                       //设置哈希
}

//挖矿交易
func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("挖矿奖励给%s", to)
	}

	txin := TXInput{[]byte{}, -1, data}                        //-1代表铸币交易，没有来源
	txout := TXOutput{subsidy, to}                             //向发起铸币交易的地址发送币
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}} //生成交易
	return &tx
}

//转账交易
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput                                       //输入
	var outputs []TXOutput                                     //输出
	acc, validOutputs := bc.FindSpendableOutputs(from, amount) //获取可以付款的币额以及可以使用的交易的输出
	if acc < amount {
		log.Panic("交易金额不足")
	}
	for txid, outs := range validOutputs { //循环遍历输出
		txID, err := hex.DecodeString(txid) //解码
		if err != nil {
			log.Panic(err) //处理错误
		}
		for _, out := range outs { //生成此交易的输入列表
			input := TXInput{txID, out, from}
			inputs = append(inputs, input) //输出的交易
		}
	}
	//生成此交易的输出
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		//记录以后的金额
		outputs = append(outputs, TXOutput{acc - amount, from})
	}
	tx := Transaction{nil, inputs, outputs} //交易
	tx.SetID()                              //设置id
	return &tx                              //返回交易
}
