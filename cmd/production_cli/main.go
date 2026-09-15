package main

import (
	"fmt"
	"go-tcg/internal/productioncli"
	"os"
)

// main 以目前工作目錄作為卡牌資料根目錄啟動 production CLI。
// 輸入為命令列參數與標準 I/O；輸出為程序結束碼，副作用為執行單局與可能建立私人 replay 檔案。
func main() {
	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "取得工作目錄失敗：%v\n", err)
		os.Exit(1)
	}
	if err := productioncli.Run(
		os.Args[1:],
		os.Stdin,
		os.Stdout,
		workingDirectory,
	); err != nil {
		fmt.Fprintf(os.Stderr, "production CLI 結束：%v\n", err)
		os.Exit(1)
	}
}
