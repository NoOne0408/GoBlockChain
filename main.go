package main

func main() {
	block := NewBlockChain()
	defer block.db.Close()
	cli := CLI{block}
	cli.Run()

	//bc := NewBlockchain()
	//bc.AddBlock("我要上岸！")
	//bc.AddBlock("初试358")
	//bc.AddBlock("英语70")
	//bc.AddBlock("数学96")
	//bc.AddBlock("专业课121")
	//
	//for _, block := range bc.blocks {
	//	fmt.Printf("上一块哈希%x", block.PrevBlockHash)
	//	fmt.Printf("数据：%s", block.Data)
	//	fmt.Printf("当前哈希%x", block.Hash)
	//	pow := NewProofOfWork(block) //校验工作量
	//	fmt.Printf("\npow:%s\n", strconv.FormatBool(pow.Validate()))
	//	fmt.Println()
	//}

}
