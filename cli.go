package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

//命令行接口
type CLI struct {
	blockchain *BlockChain
}

//创建区块链
func (cli *CLI) createBlockChain(address string) {
	bc := CreateBlockChain(address)
	bc.db.Close()
	fmt.Println("创建成功", address)
}

//查看账户
func (cli *CLI) getBalance(address string) {
	bc := NewBlockChain(address) //根据地址创建
	defer bc.db.Close()          //延迟关闭数据库
	balance := 0
	UTXOs := bc.FindUTXO(address) //查找交易金额
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("查询金额为%s:%d\n", address, balance)
}

//打印用法
func (cli *CLI) printUsage() {
	fmt.Println("用法如下")
	fmt.Println("getBalance -address ' ' 输入地址查询金额")
	fmt.Println("CreateBlockChain -address 根据地址创建区块链")
	fmt.Println("send -from From -to To -amount Amount 转账")
	fmt.Println("showBlockChain 显示区块链")

}

//校验参数合法性
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage() //显示用法
		os.Exit(1)
	}

}

func (cli *CLI) showBlockChain() {
	bc := NewBlockChain("")
	defer bc.db.Close()
	bci := bc.Iterator() //迭代
	for {
		block := bci.next()
		fmt.Printf("上一块哈希:%x\n", block.PrevBlockHash)
		fmt.Printf("当前哈希:%x\n", block.Hash)
		pow := NewProofOfWork(block) //工作量证明
		fmt.Printf("pow:%s \n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.db.Close()
	tx := NewUTXOTransaction(from, to, amount, bc) //转账
	bc.MineBlock([]*Transaction{tx})               //挖矿确认交易
	fmt.Println("交易成功")

}

//启动命令行
func (cli *CLI) Run() {
	cli.validateArgs() //校验
	//处理命令行参数
	getBalancecmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	CreateBlockChaincmd := flag.NewFlagSet("CreateBlockChain", flag.ExitOnError)
	sendcmd := flag.NewFlagSet("send", flag.ExitOnError)
	showchaincmd := flag.NewFlagSet("showBlockChain", flag.ExitOnError)

	getBalanceaddress := getBalancecmd.String("address", "", "查询地址")
	CreateBlockChainaddress := CreateBlockChaincmd.String("address", "", "查询地址")
	sendfrom := sendcmd.String("from", "", "谁给的")
	sendto := sendcmd.String("to", "", "给谁的")
	sendamount := sendcmd.Int("amount", 0, "金额")

	switch os.Args[1] {
	case "getBalance":
		err := getBalancecmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic("addBlock", err) //处理错误
		}
	case "CreateBlockChain":
		err := CreateBlockChaincmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}
	case "send":
		err := sendcmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic("addBlock", err) //处理错误
		}
	case "showBlockChain":
		err := showchaincmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic("showBlockChain", err) //处理错误
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalancecmd.Parsed() {
		if *getBalanceaddress == "" {
			getBalancecmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceaddress)
	}
	if CreateBlockChaincmd.Parsed() { //创建区块链
		if *CreateBlockChainaddress == "" {
			CreateBlockChaincmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*CreateBlockChainaddress)
	}
	if sendcmd.Parsed() {
		if *sendfrom == "" || *sendto == "" || *sendamount <= 0 {
			sendcmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendfrom, *sendto, *sendamount)
	}
	if showchaincmd.Parsed() {
		cli.showBlockChain() //显示区块链
	}
}
