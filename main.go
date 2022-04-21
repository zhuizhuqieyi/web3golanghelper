package main

import (
	"fmt"
	"log"
	"math/big"
	"time"

	_ "github.com/nikola43/go-ethereum/common/hexutil"
	"github.com/nikola43/web3manager/web3manager"
)

/*

   bsc: {
     url: 'https://bsc-dataseed.binance.org',
     accounts: [`${mnemonic}`],
     chainId: 56,
   },
   bsctestnet: {
     url: 'https://data-seed-prebsc-2-s3.binance.org:8545',
     accounts: [`${mnemonic}`],
     chainId: 97,
     gasMultiplier: 2
   },
*/

func main() {

	/*

		0x49c3Ea488e4F57e91b0aB002A16107f2A5EAD07d
		cde686c74df7db569dc5978b38ec5f051ad93a9f9729c4717993fec9a75fe335
	*/

	rpcUrl := "https://speedy-nodes-nyc.moralis.io/84a2745d907034e6d388f8d6/bsc/testnet"
	wsRpcUrl := "wss://speedy-nodes-nyc.moralis.io/84a2745d907034e6d388f8d6/bsc/testnet/ws"
	px := "cde686c74df7db569dc5978b38ec5f051ad93a9f9729c4717993fec9a75fe335"

	var goWeb3WsManager *web3manager.GoWeb3Manager = web3manager.NewGoWeb3Manager(
		rpcUrl,
		wsRpcUrl,
		px)

	fmt.Println(goWeb3WsManager.ChainId())

	/*
		wallets := make([]web3manager.Wallet, 0)



			// load .env file
			err := godotenv.Load(".env")
			if err != nil {
				log.Fatalf("Error loading .env file")
			}


		wPath := "./wallets"
		files, err := ioutil.ReadDir(wPath)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			fileName := file.Name()
			fmt.Println("fileName", fileName)

			wallet := web3manager.Wallet{
				PublicKey:  "",
				PrivateKey: "",
			}

			// Open our jsonFile
			jsonFile, _ := os.Open(wPath + "/" + fileName)
			byteValue, _ := ioutil.ReadAll(jsonFile)
			json.Unmarshal(byteValue, &wallet)
			fmt.Println(wallet)
			wallets = append(wallets, wallet)
		}
	*/

	contractAddress1 := "0x9Ac64Cc6e4415144C455BD8E4837Fea55603e5c3"
	out := make(chan string)
	var addresses []string
	addresses = append(addresses, contractAddress1)

	err2 := goWeb3WsManager.ListenBridgesEventsV2(addresses, out)
	if err2 != nil {
		log.Fatal(err2)
	}

	/*
		// connect with rpc
		rawurl := "https://bsc-dataseed.binance.org/"
		px := "479167c5f87fec4adf11306e562546622390455955a679f776cf6aa3ce0400d0"

		Web3ManagerInstance

		web3manager.Web3ManagerInstance = web3manager.NewWsWeb3Client(
			rawurl,
			px)
	*/

	//ethBasedClient := ethbasedclient.New(rawurl, wallets[0].PrivateKey)

	// contract addresses
	//pancakeContractAddress := common.HexToAddress("0x10ED43C718714eb63d5aA57B78B54704E256024E") // pancake router address
	//wBnbContractAddress := "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"                         // wbnb token adddress

	//tokenContractAddress := common.HexToAddress("0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56") // eth token adddress

	// create pancakeRouter pancakeRouterInstance
	//pancakeRouterInstance, instanceErr := PancakeRouter.NewPancake(pancakeContractAddress, ethBasedClient.Client)
	//errorsutil.HandleError(instanceErr)
	//fmt.Println("pancakeRouterInstance contract is loaded")
	//fmt.Println("pancakeRouterInstance", pancakeRouterInstance)

	// calculate gas and gas limit
	//gasLimit := uint64(2100000) // in units

	//gasPrice, gasPriceErr := gas.SuggestGasPrice(gas.GasPriorityAverage)
	//errorsutil.HandleError(gasPriceErr)

	// calculate fee and final value
	//gasFee := ethutils.CalcGasCost(gasLimit, gasPrice)
	//ethValue := ethutils.EtherToWei(big.NewFloat(0.01))
	//finalValue := big.NewInt(0).Sub(ethValue, gasFee)

	// set transaction data
	//ethBasedClient.ConfigureTransactor(finalValue, gasPrice, gasLimit)
	amountOutMin := big.NewInt(1)
	deadline := big.NewInt(time.Now().Unix() + 10000)
	//path := ethutils.GeneratePath(wBnbContractAddress, tokenContractAddress.Hex())

	fmt.Println("amountOutMin", amountOutMin)
	fmt.Println("deadline", deadline)
	//fmt.Println("path", path)

	/*
		// send transaction
		swapTx, SwapExactETHForTokensErr := pancakeRouterInstance.SwapExactETHForTokensSupportingFeeOnTransferTokens(
			ethBasedClient.Transactor,
			amountOutMin,
			path,
			ethBasedClient.Address,
			deadline)
		if SwapExactETHForTokensErr != nil {
			fmt.Println("SwapExactETHForTokensErr")
			fmt.Println(SwapExactETHForTokensErr)
		}

		txHash := swapTx.Hash().Hex()
		fmt.Println(txHash)
		genericutils.OpenBrowser("https://bscscan.com/tx/" + txHash)

		tx, err := ethutils.CancelTransaction(ethBasedClient.Client, swapTx, ethBasedClient.PrivateKey)
		errorsutil.HandleError(err)

		txHash = tx.Hash().Hex()
		fmt.Println(txHash)
		genericutils.OpenBrowser("https://bscscan.com/tx/" + txHash)
		os.Exit(0)
	*/
}

/*
	fmt.Println("ethValue")
	fmt.Println(ethValue)
	fmt.Println("finalValue")
	fmt.Println(finalValue)
	fmt.Println("gasLimit")
	fmt.Println(gasLimit)
	fmt.Println("gasPrice")
	fmt.Println(gasPrice)
	fmt.Println("gasFee")
	fmt.Println(gasFee)
	fmt.Println("nonce")
	fmt.Println(ethBasedClient.Transactor.Nonce)
	fmt.Println("amountOutMin")
	fmt.Println(amountOutMin)
	fmt.Println("path")
	fmt.Println(path)
	fmt.Println("deadline")
	fmt.Println(deadline)
	fmt.Println("transactor")
*/
