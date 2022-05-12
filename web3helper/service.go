package web3helper

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/fatih/color"
	"github.com/hokaccha/go-prettyjson"
	"github.com/hrharder/go-gas"
	"github.com/nikola43/web3golanghelper/genericutils"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/sha3"

	//web3utils "github.com/nikola43/goweb3manager/goweb3manager/util"
	pancakeRouter "github.com/nikola43/web3golanghelper/contracts/IPancakeRouter02"
	pancakePair "github.com/nikola43/web3golanghelper/contracts/IPancakePair"
	pancakeFactory "github.com/nikola43/web3golanghelper/contracts/IPancakeFactory"
)

type Reserve struct {
	Reserve0           *big.Int
	Reserve1           *big.Int
	BlockTimestampLast uint32
}

type LogLevel int

const (
	NoneLogLevel   LogLevel = 0
	LowLogLevel    LogLevel = 1
	MediumLogLevel LogLevel = 2
	HighLogLevel   LogLevel = 3
)

var defaultGasLimit = uint64(7000000)
var logLevel = HighLogLevel

type Wallet struct {
	PublicKey  string `json:"PublicKey"`
	PrivateKey string `json:"PrivateKey"`
}

type Web3GolangHelper struct {
	plainPrivateKey string
	httpClient      *ethclient.Client
	wsClient        *ethclient.Client
	FromAddress     *common.Address
}

func (w *Web3GolangHelper) AddHttpClient(httpClient *ethclient.Client) error {

	if w.httpClient != nil {
		return errors.New("web3 Http provider already instanced")
	}

	w.httpClient = httpClient
	return nil
}

func (w *Web3GolangHelper) AddWsClient(wsClient *ethclient.Client) error {

	if w.wsClient != nil {
		return errors.New("web3 websocket provider already instanced")
	}

	w.wsClient = wsClient
	return nil
}

func (w *Web3GolangHelper) SuggestGasPrice() *big.Int {

	gasPrice, err := w.selectClient().SuggestGasPrice(context.Background())

	if err != nil {
		fmt.Println(err)
		return big.NewInt(0)
	}

	return gasPrice
}

func NewWeb3GolangHelper(rpcUrl, wsUrl string, plainPrivateKey string) *Web3GolangHelper {

	goWeb3WsManager := NewWsWeb3Client(
		wsUrl,
		plainPrivateKey)

		
	goWeb3HttpManager := NewHttpWeb3Client(
		rpcUrl)
		

	goWeb3Manager := &Web3GolangHelper{
		plainPrivateKey: plainPrivateKey,
		httpClient:      goWeb3HttpManager,
		wsClient:        goWeb3WsManager,
		FromAddress:     GeneratePublicAddressFromPrivateKey(plainPrivateKey),
	}

	return goWeb3Manager

}

func NewHttpWeb3Client(rpcUrl string) *ethclient.Client {

	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatal(err)
	}

	_, getBlockErr := client.BlockNumber(context.Background())
	if getBlockErr != nil {
		log.Fatal(getBlockErr)
	}

	return client
}

func (w *Web3GolangHelper) CurrentBlockNumber() uint64 {

	blockNumber, getBlockErr := w.selectClient().BlockNumber(context.Background())
	if getBlockErr != nil {
		fmt.Println(getBlockErr)
		return 0
	}

	return blockNumber
}

func (w *Web3GolangHelper) HttpClient() *ethclient.Client {
	return w.httpClient
}

func (w *Web3GolangHelper) WebSocketClient() *ethclient.Client {
	return w.wsClient
}

func (w *Web3GolangHelper) SetPrivateKey(plainPrivateKey string) *Web3GolangHelper {
	w.plainPrivateKey = plainPrivateKey
	return w
}

func NewWsWeb3Client(rpcUrl string, plainPrivateKey interface{}) *ethclient.Client {

	_, err := url.ParseRequestURI(rpcUrl)
	if err != nil {
		log.Fatal(err)
	}

	wsClient, wsClientErr := ethclient.Dial(rpcUrl)
	if wsClientErr != nil {
		log.Fatal(wsClientErr)
	}

	_, getBlockErr := wsClient.BlockNumber(context.Background())
	if getBlockErr != nil {
		log.Fatal(getBlockErr)
	}

	return wsClient
}

func (w *Web3GolangHelper) Unsubscribe() {
	time.Sleep(10 * time.Second)
	fmt.Println("---unsubscribe-----")
	//w.ethSubscription.Unsubscribe()
}

func (w *Web3GolangHelper) GetEthBalance(address string) *big.Int {
	account := common.HexToAddress(address)
	balance, err := w.httpClient.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil
	}
	return balance
}

func (w *Web3GolangHelper) IsAddressContract(address string) bool {

	if !ValidateAddress(address) {
		return false
	}

	bytecode, err := w.httpClient.CodeAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return false
	}
	return len(bytecode) > 0
}

func (w *Web3GolangHelper) ChainId() *big.Int {
	chainID, err := w.httpClient.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return chainID
}

func (w *Web3GolangHelper) PendingNonce() *big.Int {
	nonce, err := w.selectClient().PendingNonceAt(context.Background(), *w.FromAddress)
	if err != nil {
		log.Fatal(err)
	}
	// calculate next nonce
	return big.NewInt(int64(nonce))
}
func (w *Web3GolangHelper) SignTx(tx *types.Transaction) (*types.Transaction, error) {

	privateKey, privateKeyErr := crypto.HexToECDSA(w.plainPrivateKey)
	if privateKeyErr != nil {
		return nil, privateKeyErr
	}

	signedTx, signTxErr := types.SignTx(tx, types.NewEIP155Signer(w.ChainId()), privateKey)
	if signTxErr != nil {
		return nil, signTxErr
	}

	return signedTx, nil
}

func (w *Web3GolangHelper) NewContract(contractAddress string) {

	/*
		address := common.HexToAddress(contractAddress)
		instance, err := store.NewStore(address, w.httpClient)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("contract is loaded")
		return instance
	*/
}

func (w *Web3GolangHelper) SubscribeContractBridgeBSCEvent(contractAddressString string) error {

	if w.wsClient == nil {
		return errors.New("Nil Web3 Websocket Client")
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(contractAddressString)},
	}

	logs := make(chan types.Log)
	sub, err := w.wsClient.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Init Sub")
	for {
		select {
		case err := <-sub.Err():
			fmt.Println("Error")
			fmt.Println(err)
			log.Fatal(err)
		case vLog := <-logs:
			fmt.Println("Data")
			fmt.Println(string(vLog.Data))
			//fmt.Println("vLog.Address: " + vLog.Address.Hex())
			fmt.Println("vLog.TxHash: " + vLog.TxHash.Hex())
			fmt.Println("vLog.BlockNumber: " + strconv.FormatUint(vLog.BlockNumber, 10))

			/*

					event := struct {
						Key   [32]byte
						Value [32]byte
					}{}


				contractAbi, err := abi.JSON(strings.NewReader(bridgeAvax.BridgeAvaxMetaData.ABI))
				if err != nil {
					log.Fatal(err)
				}

				//r, err := contractAbi.Unpack(&event, "ItemSet", vLog.Data)
				r, err := contractAbi.Unpack("Transfer", vLog.Data)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(r)

			*/

			//fmt.Println(string(event.Key[:]))   // foo
			//fmt.Println(string(event.Value[:])) // bar

			fmt.Println("")
			//fmt.Println(vLog) // pointer to event log
		}
	}
}

func (w *Web3GolangHelper) EstimateTxResult(to string, txData []byte) bool {
	estimatedGas := w.EstimateGas(to, txData)
	return estimatedGas > 0
}

func (w *Web3GolangHelper) EstimateGas(to string, txData []byte) uint64 {
	toAddress := common.HexToAddress(to)
	estimateGas, estimateGasErr := w.selectClient().EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &toAddress,
		Data: txData,
	})
	if estimateGasErr != nil {
		return 0
	}
	return estimateGas
}

func (w *Web3GolangHelper) BuildContractEventSubscription(contractAddress string, logs chan types.Log) ethereum.Subscription {

	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(contractAddress)},
	}

	sub, err := w.WebSocketClient().SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		fmt.Println(sub)
	}
	return sub
}

func (w *Web3GolangHelper) SendTokens(tokenAddressString, toAddressString string, value *big.Int) (string, *big.Int, error) {

	toAddress := common.HexToAddress(toAddressString)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(value.Bytes(), 32)

	txData := BuildTxData(methodID, paddedAddress, paddedAmount)

	estimateGas := w.EstimateGas(tokenAddressString, txData)
	txId, txNonce, err := w.SignAndSendTransaction(toAddressString, ToWei(value, 18), txData, w.PendingNonce(), nil, estimateGas)
	if err != nil {
		return "", big.NewInt(0), err
	}

	return txId, txNonce, nil
}

func (w *Web3GolangHelper) selectClient() *ethclient.Client {
	var selectedClient *ethclient.Client
	if w.wsClient != nil {
		selectedClient = w.wsClient
	} else {
		if w.httpClient != nil {
			selectedClient = w.httpClient
		} else {
			log.Fatal("SuggestGasPrice: Not conected")
		}
	}
	return selectedClient
}

func (w *Web3GolangHelper) SendEth(toAddressString string, value string) (string, *big.Int, error) {

	txId, nonce, err := w.SignAndSendTransaction(toAddressString, ToWei(value, 18), make([]byte, 0), w.PendingNonce(), nil, nil)
	if err != nil {
		return "", big.NewInt(0), err
	}

	return txId, nonce, nil
}

func (w *Web3GolangHelper) SignAndSendTransaction(toAddressString string, value *big.Int, data []byte, nonce *big.Int, customGasPrice interface{}, customGasLimit interface{}) (string, *big.Int, error) {

	usedGasPrice, _ := w.selectClient().SuggestGasPrice(context.Background())
	if logLevel == MediumLogLevel {
		fmt.Println(color.CyanString("usedGasPrice -> suggestGasPrice: "), color.YellowString(strconv.Itoa(int(usedGasPrice.Int64())))+"\n")
	}

	if customGasPrice != nil {
		usedGasPrice = customGasPrice.(*big.Int)

		if logLevel == MediumLogLevel {
			fmt.Println(color.CyanString("usedGasPrice -> customGasPrice: "), color.YellowString(strconv.Itoa(int(usedGasPrice.Int64())))+"\n")
		}
	}

	usedGasLimit := defaultGasLimit
	if logLevel == MediumLogLevel {
		fmt.Println(color.CyanString("usedGasLimit -> defaultGasLimit: "), color.YellowString(strconv.Itoa(int(usedGasLimit)))+"\n")
	}

	if customGasLimit != nil {
		usedGasLimit = customGasLimit.(uint64)

		if logLevel == MediumLogLevel {
			fmt.Println(color.CyanString("usedGasLimit -> customGasLimit: "), color.YellowString(strconv.Itoa(int(usedGasLimit)))+"\n")
		}
	} else {
		if len(data) > 0 {
			usedGasLimit = w.EstimateGas(toAddressString, data)
			if logLevel == MediumLogLevel {
				fmt.Println(color.CyanString("usedGasLimit -> w.EstimateGas: "), color.YellowString(strconv.Itoa(int(usedGasLimit)))+"\n")
			}
		} else {

		}
	}

	toAddress := common.HexToAddress(toAddressString)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce.Uint64(),
		GasPrice: usedGasPrice,
		Gas:      usedGasLimit,
		To:       &toAddress,
		Value:    value,
		Data:     data,
	})

	singedTx, signTxErr := w.SignTx(tx)
	if signTxErr != nil {
		return "", big.NewInt(0), signTxErr
	}

	sendTxErr := w.selectClient().SendTransaction(context.Background(), singedTx)
	if sendTxErr != nil {
		return "", big.NewInt(0), sendTxErr
	}

	if logLevel == HighLogLevel {

		b, e := singedTx.MarshalJSON()
		if e != nil {
			fmt.Println("SendTransaction")
			return "", big.NewInt(0), e
		}

		var result map[string]interface{}
		json.Unmarshal(b, &result)
		s, _ := prettyjson.Marshal(result)

		timestamp := time.Now().Unix()

		fmt.Println(color.GreenString("Raw Transaction Hash: "), color.YellowString(tx.Hash().Hex()))
		fmt.Println(color.CyanString("Transaction Hash: "), color.YellowString(singedTx.Hash().Hex()))
		fmt.Println(color.MagentaString("Timestamp: "), color.YellowString(strconv.Itoa(int(timestamp))))
		fmt.Println(string(s))

		//OpenBrowser("https://testnet.snowtrace.io/tx/" + singedTx.Hash().Hex())
	}

	return singedTx.Hash().Hex(), nonce, nil
}

func (w *Web3GolangHelper) CancelTx(to string, nonce *big.Int, multiplier int64) (string, error) {

	gasPrice, _ := w.selectClient().SuggestGasPrice(context.Background())

	txId, _, err := w.SignAndSendTransaction(
		to,
		ToWei(0, 0),
		make([]byte, 0),
		nonce,
		nil,
		big.NewInt(gasPrice.Int64()*multiplier))

	if err != nil {
		return "", err
	}

	return txId, nil
}

func (w *Web3GolangHelper) GenerateContractEventSubscription(contractAddress string) (chan types.Log, ethereum.Subscription, error) {

	logs := make(chan types.Log)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(contractAddress)},
	}

	sub, err := w.wsClient.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		return nil, nil, err
	}

	return logs, sub, nil
}


func (w *Web3GolangHelper) Buy(tokenAddress string) {
	// contract addresses
	pancakeContractAddress := common.HexToAddress("0x9Ac64Cc6e4415144C455BD8E4837Fea55603e5c3") // pancake router address
	wBnbContractAddress := "0xae13d989daC2f0dEbFf460aC112a837C89BAa7cd"                         // wbnb token adddress
	tokenContractAddress := common.HexToAddress(tokenAddress)                                   // eth token adddress

	
	// create pancakeRouter pancakeRouterInstance
	pancakeRouterInstance, instanceErr := pancakeRouter.NewPancake(pancakeContractAddress, w.HttpClient())
	if instanceErr != nil {
		fmt.Println(instanceErr)
	}
	fmt.Println("pancakeRouterInstance contract is loaded")

	// calculate gas and gas limit
	gasLimit := uint64(2100000) // in units
	gasPrice, gasPriceErr := gas.SuggestGasPrice(gas.GasPriorityAverage)
	if gasPriceErr != nil {
		fmt.Println(gasPriceErr)
	}

	fmt.Println(

		wBnbContractAddress,
		tokenContractAddress,
		pancakeRouterInstance,
		gasLimit,
		gasPrice,
	)

	// calculate fee and final value
	gasFee := CalcGasCost(gasLimit, gasPrice)
	ethValue := EtherToWei(big.NewFloat(0.1))
	finalValue := big.NewInt(0).Sub(ethValue, gasFee)

	// set transaction data
	transactor := w.BuildTransactor(finalValue, gasPrice, gasLimit)
	amountOutMin := big.NewInt(1.0)
	deadline := big.NewInt(time.Now().Unix() + 10000)
	path := GeneratePath(wBnbContractAddress, tokenContractAddress.Hex())

	swapTx, SwapExactETHForTokensErr := pancakeRouterInstance.SwapExactETHForTokensSupportingFeeOnTransferTokens(
		transactor,
		amountOutMin,
		path,
		*w.FromAddress,
		deadline)
	if SwapExactETHForTokensErr != nil {
		fmt.Println("SwapExactETHForTokensErr")
		fmt.Println(SwapExactETHForTokensErr)
	}

	fmt.Println(swapTx)

	txHash := swapTx.Hash().Hex()
	fmt.Println(txHash)
	genericutils.OpenBrowser("https://testnet.bscscan.com/tx/" + txHash)
}

/*
   function swapExactETHForTokensSupportingFeeOnTransferTokens(
       uint amountOutMin,
       address[] calldata path,
       address to,
       uint deadline
   ) external payable;
*/

func (w *Web3GolangHelper)  BuyV2(tokenAddress string, value *big.Int) {
	toAddress := common.HexToAddress("0x9Ac64Cc6e4415144C455BD8E4837Fea55603e5c3")
	wBnbContractAddress := "0xae13d989daC2f0dEbFf460aC112a837C89BAa7cd"

	transferFnSignature := []byte("swapExactETHForTokensSupportingFeeOnTransferTokens(uint,address[],address,uint)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]

	path := GeneratePath(wBnbContractAddress, tokenAddress)
	pathString := []string{path[0].Hex(), path[1].Hex()}

	deadline := big.NewInt(time.Now().Unix() + 10000)
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(pathString)
	bs := buf.Bytes()
	fmt.Printf("%q", bs)

	paddedAmountOutMin := common.LeftPadBytes(value.Bytes(), 32)
	paddedPathA := common.LeftPadBytes(path[0].Bytes(), 32)
	paddedPathB := common.LeftPadBytes(path[1].Bytes(), 32)
	paddedPath := common.LeftPadBytes(bs, 32)
	paddedTo := common.LeftPadBytes(toAddress.Bytes(), 32)
	paddedDeadline := common.LeftPadBytes(deadline.Bytes(), 32)

	fmt.Println("paddedAmountOutMin", paddedAmountOutMin)
	fmt.Println("paddedPathA", paddedPathA)
	fmt.Println("paddedPathB", paddedPathB)
	fmt.Println("paddedPath", paddedPath)
	fmt.Println("paddedTo", paddedTo)
	fmt.Println("paddedDeadline", paddedDeadline)
	fmt.Println("paddedAmountOutMin", paddedAmountOutMin)
	fmt.Println("paddedAmountOutMin", paddedAmountOutMin)

	txData := BuildTxData(methodID, paddedAmountOutMin, paddedPath, paddedTo, paddedDeadline)

	fmt.Println("txData", txData)

	estimateGas := w.EstimateGas(toAddress.Hex(), txData)

	fmt.Println("estimateGas", estimateGas)

	txId, txNonce, err := w.SignAndSendTransaction(toAddress.Hex(), ToWei(value, 18), txData, w.PendingNonce(), nil, estimateGas)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(txId)
	fmt.Println(txNonce)
}

func (w *Web3GolangHelper) ListenBridgesEventsV2(contractsAddresses []string, out chan<- []chan types.Log) error {

	var logs []chan types.Log
	var subs []ethereum.Subscription

	fmt.Println("")
	fmt.Println(color.YellowString("  --------------------- Contracts Subscriptions ---------------------"))
	for i := 0; i < len(contractsAddresses); i++ {

		contractLog, contractSub, err := w.GenerateContractEventSubscription(contractsAddresses[i])
		if err != nil {
			return err
		}

		logs = append(logs, contractLog)
		subs = append(subs, contractSub)

		go func(i int) {
			fmt.Println(color.MagentaString("    Init Subscription: "), color.YellowString(contractsAddresses[i]))

			for {
				select {
				case err := <-subs[i].Err():
					fmt.Println(err)
					out <- logs

				case vLog := <-logs[i]:
					//fmt.Println(vLog) // pointer to event log
					fmt.Println("Data logs")
					fmt.Println(string(vLog.Data))
					//fmt.Println("vLog.Address: " + vLog.Address.Hex())
					fmt.Println("vLog.TxHash: " + vLog.TxHash.Hex())
					fmt.Println("vLog.BlockNumber: " + strconv.FormatUint(vLog.BlockNumber, 10))
					fmt.Println("")
					//out <- vLog.TxHash.Hex()
					out <- logs
				}
			}
		}(i)
	}
	return nil
}

func (w *Web3GolangHelper) SwitchAccount(plainPrivateKey string) {
	// create privateKey from string key
	privateKey, privateKeyErr := crypto.HexToECDSA(plainPrivateKey)
	if privateKeyErr != nil {
		fmt.Println(privateKeyErr)
	}


	// generate public key and address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	// generate address from public key
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	w.FromAddress = &address
}

func (w *Web3GolangHelper) BuildTransactor(value *big.Int, gasPrice *big.Int, gasLimit uint64) *bind.TransactOpts  {
	privateKey, privateKeyErr := crypto.HexToECDSA(w.plainPrivateKey)
	if privateKeyErr != nil {
		fmt.Println(privateKeyErr)
	}

	transactor, transactOptsErr := bind.NewKeyedTransactorWithChainID(privateKey, w.ChainId())

	if transactOptsErr != nil {
		fmt.Println(transactOptsErr)
	}

	if value.String() != "-1" {
		transactor.Value = value
	}

	transactor.GasPrice = gasPrice
	transactor.GasLimit = gasLimit
	transactor.Nonce = w.PendingNonce()
	transactor.Context = context.Background()
	return transactor
}

func (w *Web3GolangHelper) Balance(account common.Address) *big.Int {
	// get current balance
	balance, balanceErr := w.httpClient.BalanceAt(context.Background(), account, nil)
	if balanceErr != nil {
		fmt.Println(balanceErr)
	}

	return balance
}


func GweiToEther(wei *big.Int) *big.Float {
	f := new(big.Float)
	f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	fWei.SetMode(big.ToNearestEven)
	return f.Quo(fWei.SetInt(wei), big.NewFloat(params.GWei))
}

func GweiToWei(wei *big.Int) *big.Int {
	eth := GweiToEther(wei)
	ethWei := EtherToWei(eth)
	return ethWei
}

// Wei ->
func WeiToGwei(wei *big.Int) *big.Int {
	f := new(big.Float)
	f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	fWei.SetMode(big.ToNearestEven)
	v := f.Quo(fWei.SetInt(wei), big.NewFloat(params.GWei))
	i, _ := new(big.Int).SetString(v.String(), 10)

	return i
}

func EtherToGwei(eth *big.Float) *big.Int {
	truncInt, _ := eth.Int(nil)
	truncInt = new(big.Int).Mul(truncInt, big.NewInt(params.GWei))
	fracStr := strings.Split(fmt.Sprintf("%.9f", eth), ".")[1]
	fracStr += strings.Repeat("0", 9-len(fracStr))
	fracInt, _ := new(big.Int).SetString(fracStr, 10)
	wei := new(big.Int).Add(truncInt, fracInt)
	return wei
}

// CalcGasCost calculate gas cost given gas limit (units) and gas price (wei)
func CalcGasCost(gasLimit uint64, gasPrice *big.Int) *big.Int {
	gasLimitBig := big.NewInt(int64(gasLimit))
	return gasLimitBig.Mul(gasLimitBig, gasPrice)
}

func GeneratePath(tokenAContractPlainAddress string, tokenBContractPlainAddress string) []common.Address {
	tokenAContractAddress := common.HexToAddress(tokenAContractPlainAddress)
	tokenBContractAddress := common.HexToAddress(tokenBContractPlainAddress)

	path := make([]common.Address, 0)
	path = append(path, tokenAContractAddress)
	path = append(path, tokenBContractAddress)

	return path
}

func CancelTransaction(client *ethclient.Client, transaction *types.Transaction, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	value := big.NewInt(0)

	// generate public key and address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	// generate address from public key
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	var data []byte

	fmt.Println(transaction.GasPrice())

	newGasPrice := big.NewInt(0).Add(transaction.GasPrice(), big.NewInt(0).Div(big.NewInt(0).Mul(transaction.GasPrice(), big.NewInt(10)), big.NewInt(100)))
	fmt.Println(newGasPrice)
	tx := types.NewTransaction(transaction.Nonce(), address, value, transaction.Gas(), newGasPrice, data)

	// get chain id
	chainID, chainIDErr := client.ChainID(context.Background())
	if chainIDErr != nil {
		log.Fatal(chainIDErr)
		return nil, chainIDErr
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return signedTx, nil
}


/*
func (w *Web3GolangHelper) getTokenPairs(token *models.EventsCatched) string {
	//lpPairs := make([]*models.LpPair, 0)

	lpPairAddress := w.getPair(token.TokenAddress)

	//append(lpPairs, )

	fmt.Println("lpPairAddress", lpPairAddress)
	return lpPairAddress
}
*/

func (w *Web3GolangHelper)  GetReserves(tokenAddress string) Reserve {

	pairInstance, instanceErr := pancakePair.NewPancake(common.HexToAddress("0xB7926C0430Afb07AA7DEfDE6DA862aE0Bde767bc"), w.HttpClient())
	if instanceErr != nil {
		fmt.Println(instanceErr)
	}

	reserves, getReservesErr := pairInstance.GetReserves(nil)
	if getReservesErr != nil {
		fmt.Println(getReservesErr)
	}

	return reserves
}

func (w *Web3GolangHelper) GetPair(tokenAddress string) string {

	factoryInstance, instanceErr := pancakeFactory.NewPancake(common.HexToAddress("0xB7926C0430Afb07AA7DEfDE6DA862aE0Bde767bc"), w.HttpClient())
	if instanceErr != nil {
		fmt.Println(instanceErr)
	}

	wBnbContractAddress := "0xae13d989daC2f0dEbFf460aC112a837C89BAa7cd"

	lpPairAddress, getPairErr := factoryInstance.GetPair(nil, common.HexToAddress(wBnbContractAddress), common.HexToAddress(tokenAddress))
	if getPairErr != nil {
		fmt.Println(getPairErr)
	}

	return lpPairAddress.Hex()

}


// IsValidAddress validate hex address
func IsValidAddress(iaddress interface{}) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	switch v := iaddress.(type) {
	case string:
		return re.MatchString(v)
	case common.Address:
		return re.MatchString(v.Hex())
	default:
		return false
	}
}

// IsZeroAddress validate if it's a 0 address
func IsZeroAddress(iaddress interface{}) bool {
	var address common.Address
	switch v := iaddress.(type) {
	case string:
		address = common.HexToAddress(v)
	case common.Address:
		address = v
	default:
		return false
	}

	zeroAddressBytes := common.FromHex("0x0000000000000000000000000000000000000000")
	addressBytes := address.Bytes()
	return reflect.DeepEqual(addressBytes, zeroAddressBytes)
}

// ToDecimal wei to decimals
func ToDecimal(ivalue interface{}, decimals int) decimal.Decimal {
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		value.SetString(v, 10)
	case *big.Int:
		value = v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result
}

// ToWei decimals to wei
func ToWei(iamount interface{}, decimals int) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}

func WeiToEther(wei *big.Int) *big.Float {
	f := new(big.Float)
	f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	fWei.SetMode(big.ToNearestEven)
	return f.Quo(fWei.SetInt(wei), big.NewFloat(params.Ether))
}

func EtherToWei(eth *big.Float) *big.Int {
	truncInt, _ := eth.Int(nil)
	truncInt = new(big.Int).Mul(truncInt, big.NewInt(params.Ether))
	fracStr := strings.Split(fmt.Sprintf("%.18f", eth), ".")[1]
	fracStr += strings.Repeat("0", 18-len(fracStr))
	fracInt, _ := new(big.Int).SetString(fracStr, 10)
	wei := new(big.Int).Add(truncInt, fracInt)
	return wei
}

func GeneratePublicAddressFromPrivateKey(plainPrivateKey string) *common.Address {
	privateKey, err := crypto.HexToECDSA(plainPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &fromAddress
}

func ValidateAddress(address string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	return re.MatchString(address)
}

// SigRSV signatures R S V returned as arrays
func SigRSV(isig interface{}) ([32]byte, [32]byte, uint8) {
	var sig []byte
	switch v := isig.(type) {
	case []byte:
		sig = v
	case string:
		sig, _ = hexutil.Decode(v)
	}

	sigstr := common.Bytes2Hex(sig)
	rS := sigstr[0:64]
	sS := sigstr[64:128]
	R := [32]byte{}
	S := [32]byte{}
	copy(R[:], common.FromHex(rS))
	copy(S[:], common.FromHex(sS))
	vStr := sigstr[128:130]
	vI, _ := strconv.Atoi(vStr)
	V := uint8(vI + 27)

	return R, S, V
}

func BuildTxData(data ...[]byte) []byte {
	var txData []byte

	for _, v := range data {
		txData = append(txData, v...)
	}

	return txData
}

func GenerateAddressFromPlainPrivateKey(pk string) (common.Address, *ecdsa.PrivateKey, error) {

	var address common.Address
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return address, privateKey, err
	}

	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return address, privateKey, errors.New("error casting public key to ECDSA")
	}

	return crypto.PubkeyToAddress(*publicKeyECDSA), privateKey, nil
}
