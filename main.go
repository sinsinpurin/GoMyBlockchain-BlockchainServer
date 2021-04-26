package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/sinsinpurin/gomyblockchain"
	"github.com/sinsinpurin/gomyblockchain-blockchainserver/apitypes"
)

// Port ポート番号の設定 ex. :8085
var Port string

var blockChainCache gomyblockchain.BlockChainServer

func getBlockChain() *gomyblockchain.BlockChain {
	if blockChainCache.BlockChain.BlockChainAddress == "" {
		blockChainCache.Wallet = gomyblockchain.GenerateWallet()
		intPort, _ := strconv.Atoi(Port[1:])
		blockChainCache.BlockChain = *gomyblockchain.InitBlockChain(blockChainCache.Wallet.Address, intPort)
	}
	return &blockChainCache.BlockChain
}

func getBlockChainAddress() *gomyblockchain.Wallet {
	if blockChainCache.BlockChain.BlockChainAddress == "" {
		blockChainCache.Wallet = gomyblockchain.GenerateWallet()
		intPort, _ := strconv.Atoi(Port[1:])
		blockChainCache.BlockChain = *gomyblockchain.InitBlockChain(blockChainCache.Wallet.Address, intPort)
	}
	return &blockChainCache.Wallet
}

func main() {
	flag.StringVar(&Port, "p", ":8085", "Set Port Number")
	flag.Parse()

	bc := getBlockChain()
	go bc.SyncNeighbours()
	bcWallet := getBlockChainAddress()

	e := echo.New()
	e.GET("/chain", func(c echo.Context) error {
		return c.JSONPretty(http.StatusOK, bc.Chain, "	")
	})

	// wallet chacheされているwalletをJSONで返します
	e.POST("/wallet", func(c echo.Context) error {
		// wallet := blockchain.GenerateWallet()
		wallet := blockChainCache.Wallet
		return c.JSON(http.StatusOK, wallet)
	})

	e.GET("/transactions", func(c echo.Context) error {
		response := &apitypes.GetTransactionsResponseType{
			Transactions: bc.TransactionPool,
			Length:       len(bc.TransactionPool),
		}
		return c.JSONPretty(http.StatusOK, response, "	")
	})
	e.POST("/transactions", func(c echo.Context) error {
		param := new(apitypes.PostTransactionsRequestType)
		if err := c.Bind(param); err != nil {
			fmt.Println(err)
			return err
		}
		if param.RecipientAddress == "" || param.SenderAddress == "" || param.SenderPublicKey == "" || param.Value == 0 || param.Signature == "" {
			fmt.Println("Invalid variable")
			return c.JSON(http.StatusBadRequest, nil)
		}
		bSPK, _ := hex.DecodeString(param.SenderPublicKey)
		bSig, _ := hex.DecodeString(param.Signature)
		isAddTransaction := bc.AddTransactionSync(
			gomyblockchain.CreateTransaction(param.SenderAddress, param.RecipientAddress, param.Value),
			bSPK,
			bSig,
		)
		if isAddTransaction == false {
			fmt.Println("add transaction faild")
			return c.JSON(http.StatusExpectationFailed, nil)
		}
		return c.JSON(http.StatusOK, nil)
	})

	e.PUT("/transactions", func(c echo.Context) error {
		param := new(apitypes.PostTransactionsRequestType)
		if err := c.Bind(param); err != nil {
			fmt.Println(err)
			return err
		}

		if param.RecipientAddress == "" || param.SenderAddress == "" || param.SenderPublicKey == "" || param.Value == 0 || param.Signature == "" {
			fmt.Println("Invalid variable")
			return c.JSON(http.StatusBadRequest, nil)
		}
		bSPK, _ := hex.DecodeString(param.SenderPublicKey)
		bSig, _ := hex.DecodeString(param.Signature)
		isUpdate := bc.AddTransaction(
			gomyblockchain.CreateTransaction(param.SenderAddress, param.RecipientAddress, param.Value),
			bSPK,
			bSig,
		)
		if isUpdate == false {
			fmt.Println("add transaction faild")
			return c.JSON(http.StatusExpectationFailed, nil)
		}
		return c.JSON(http.StatusOK, "sync transaction")
	})

	e.DELETE("/transactions", func(c echo.Context) error {
		bc.TransactionPool = nil
		return c.JSON(http.StatusOK, nil)
	})

	e.GET("/mine", func(c echo.Context) error {
		result := bc.Mining(bcWallet.Address)
		if result {
			return c.JSON(http.StatusOK, "Mine Success")
		}
		return c.JSON(http.StatusBadRequest, nil)
	})

	miningStatusCh := make(chan bool)

	e.GET("/mine/start", func(c echo.Context) error {
		go bc.BackgroundMining(miningStatusCh, bcWallet.Address)
		return c.JSON(http.StatusOK, "Mining Start")
	})

	e.GET("/mine/stop", func(c echo.Context) error {
		miningStatusCh <- false
		return c.JSON(http.StatusOK, "Mining Stop")
	})

	e.PUT("/consensus", func(c echo.Context) error {
		result := bc.ResolveConflicts()
		return c.JSON(http.StatusOK, result)
	})

	e.GET("/amount", func(c echo.Context) error {
		query := new(apitypes.GetAmountRequestType)
		query.Address = c.QueryParam("Address")
		resp := apitypes.GetAmountResponseType{
			Amount: bc.CalculateTotalAmount(query.Address),
		}
		return c.JSON(http.StatusOK, resp)
	})

	e.Logger.Fatal(e.Start(Port))
}
