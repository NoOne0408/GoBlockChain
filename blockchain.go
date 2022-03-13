package main

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const dbFile = "blockchain.db" //数据库名
const blockBucket = "blocks"   //名称
const genesisCoinbaseData = "xxx的区块链"

type BlockChain struct {
	tip []byte   //二进制数据
	db  *bolt.DB //数据库
}

type BlockChainIterator struct {
	currentHash []byte   //当前哈希
	db          *bolt.DB //数据库

}

//挖矿交易
func (blockchain *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte //最后的哈希
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		lastHash = bucket.Get([]byte("1"))
		return nil
	})
	if err != nil {
		log.Panic(err) //处理错误
	}

	newBlock := NewBlock(transactions, lastHash) //创建新的区块
	err = blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		err := bucket.Put(newBlock.Hash, newBlock.Serialize()) //存入数据库
		if err != nil {
			log.Panic(err) //处理错误
		}
		err = bucket.Put([]byte("1"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		blockchain.tip = newBlock.Hash //保存上一块的哈希
		return nil
	})
}

//获取此地址的没有使用输出的交易
func (blockchain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction        //交易事务
	spentTXOS := make(map[string][]int) //开辟内存
	bci := blockchain.Iterator()        //迭代器
	for {
		block := bci.next()                     //循环每一个区块
		for _, tx := range block.Transactions { //循环区块中的每个交易
			txID := hex.EncodeToString(tx.ID) //获取交易ID转为字符串
		Outputs:
			for outindex, out := range tx.Vout { //循环交易中的输出
				if spentTXOS[txID] != nil {
					for _, spentOut := range spentTXOS[txID] {
						if spentOut == outindex {
							continue Outputs //循环到不等
						}
					}
				}
				if out.CanBeUnlockedOutPutWith(address) { //如果此地址可以解锁这笔交易中的这个输出，就将其加入可花费列表
					unspentTXs = append(unspentTXs, *tx) //加入列表
				}
			}

			if tx.IsCoinBase() == false { //如果是普通交易而非铸币交易
				for _, in := range tx.Vin {
					if in.CanUnlockOutPutWith(address) { //判断是否可以解锁
						inTxID := hex.EncodeToString(in.Txid) //编码为字符串
						spentTXOS[inTxID] = append(spentTXOS[inTxID], in.Vout)
					}
				}
			}

		}
		if len(block.PrevBlockHash) == 0 { //最后一块跳出
			break
		}
	}
	return unspentTXs
}

//获取所有没有使用的交易输出
func (blockchain *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput                                               //数组用来存放 所有向本地址转的币
	unspentTransactions := blockchain.FindUnspentTransactions(address) //查找所有向本地址转币的交易
	for _, tx := range unspentTransactions {                           //循环所有交易
		for _, out := range tx.Vout { //循环所有交易输出
			if out.CanBeUnlockedOutPutWith(address) { //判断是否锁定
				UTXOs = append(UTXOs, out) //加入数据
			}
		}

	}
	return UTXOs

}

//查找进行转账的交易
//在转账前首先要遍历和此地址相关的所有的收账的交易
func (blockchain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)                  //输出
	unspentTxs := blockchain.FindUnspentTransactions(address) //根据地址查看所有交易
	accmulated := 0                                           //累计
Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID) //获取编号
		for outindex, out := range tx.Vout {
			if out.CanBeUnlockedOutPutWith(address) && accmulated < amount {
				accmulated += out.Value                                       //统计金额
				unspentOutputs[txID] = append(unspentOutputs[txID], outindex) //可以使用的交易输出叠加起来
				if accmulated >= amount {                                     //当累计金额足够时跳出循环
					break Work
				}
			}
		}
	}
	return accmulated, unspentOutputs

}

//新建一个区块
func NewBlockChain() *BlockChain {
	//无数据库先创建
	if dbExists() == false {
		fmt.Println("数据库不存在，请先创建一个")
		os.Exit(1) //退出
	}
	var tip []byte                          //存储数据库的二进制数据
	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err) //处理打开错误
	}
	//处理数据更新
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //按照名称打开数据库表格
		tip = bucket.Get([]byte("1"))            //获取区块指针
		return nil
	})
	if err != nil {
		log.Panic(err) //处理数据库更新错误
	}
	bc := BlockChain{tip, db} //创建一个区块链
	return &bc

}

//迭代器
func (blockchain *BlockChain) Iterator() *BlockChainIterator {
	bcit := &BlockChainIterator{blockchain.tip, blockchain.db}
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

//判断数据库是否存在
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

//创建一个区块链一个数据库
func CreateBlockChain(address string) *BlockChain {
	if dbExists() {
		fmt.Println("数据库存在！无需创建")
		os.Exit(1) //退出

	}
	var tip []byte                          //存储数据库的二进制数据
	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err) //处理打开错误
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinBaseTX(address, genesisCoinbaseData) //创建创世区块的事务交易
		genesis := NewGenesisBlock(cbtx)                    //根据创世区块的事务交易创建创世区块
		bucket, err := tx.CreateBucket([]byte(blockBucket))
		if err != nil {
			log.Panic(err) //处理更新错误
		}
		err = bucket.Put(genesis.Hash, genesis.Serialize()) //存储
		if err != nil {
			log.Panic(err) //处理压入错误
		}
		err = bucket.Put([]byte("1"), genesis.Hash) //创建区块指针
		if err != nil {
			log.Panic(err) //处理压入错误
		}
		tip = genesis.Hash
		return nil
	})

	bc := BlockChain{tip, db} //返回创建好的区块链
	return &bc
}
