package pay

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	"gitlab.com/wolfplus/gamespace-lobby/define"
	// "gitlab.com/wolfplus/gamespace-lobby/db/ethPay"
)

var (
	client *ethclient.Client
	logger *log.Entry
)

type EthPay struct {
}

func NewEthPay() *EthPay {
	return &EthPay{}
}

func (e EthPay) Init() error {
	// go ListenPublicChain()

	logger = log.WithField("source", "pay")

	cli, err := ethclient.Dial("https://mainnet.infura.io/v3/8178e77d36ed4214b8806f68027c08aa")
	if err != nil {
		return err
	}
	client = cli

	return nil
}

func (e *EthPay) AfterInit() {
}

func (e EthPay) BeforeShutdown() {
}

func (e EthPay) Shutdown() error {
	return nil
}
func GetBalance(address string) float64 {
	account := common.HexToAddress(address)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return 0.0
	}
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
	value, _ := ethValue.Float64()
	return value
}

var ethWalletAccList = [5]string{
	"cb5afa40ca9901e0c998d0cfe038b485901dac2dbc576ae6f7df4ac1ecd78d95",
	"c46e1ec78c5f4691725a350da75d42fe99b64243f9e3e8df5615fe37a2e8f218",
	"006a29bcc064058f429d828b6527922e5a21563317ba61fbacb0a8505c5dfb70",
	"17394c6498dc0610a3d4003d6df11808bfa826d2b46584d14db1c2cecee8be41",
	"d657c33a35bc0ccdd4f3c14c3d0027560f700e670fc74066519d36e16ec62770",
}

var ethWalletAccIndex = -1

func NewAccount() (string, string, error) {

	ethWalletAccIndex += 1

	privateKey, err := crypto.HexToECDSA(ethWalletAccList[ethWalletAccIndex%5])
	// privateKey, err := crypto.GenerateKey()

	if err != nil {
		return "", "", err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyStr := hexutil.Encode(privateKeyBytes)[2:]

	fmt.Println("my creating new account and logging private key!! ", privateKey)

	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	logger.Infof("pri:%s, pub:%s, balance:%f", privateKeyStr, address, GetBalance(address))
	return privateKeyStr, address, nil
}
func (e *EthPay) createAccount() (string, error) {
	ks := keystore.NewKeyStore("./wallets", keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.NewAccount(define.EthWalletPrivateSecret)
	if err != nil {
		return "", err
	}
	return account.Address.Hex(), nil
}
func (e *EthPay) importAccount() (string, error) {
	file := "./wallets/UTC--2020-03-22T15-28-17.954198300Z--f13ed962ac125666ca17bb3a6e70180ac0c1de07"
	ks := keystore.NewKeyStore("./tmp", keystore.StandardScryptN, keystore.StandardScryptP)
	jsonBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	account, err := ks.Import(jsonBytes, define.EthWalletPrivateSecret, define.EthWalletPrivateSecret)
	if err != nil {
		return "", err
	}
	logger.Infof("addr:%s", account.Address.Hex())
	for i := 0; i < 10; i++ {
		acc, _ := ks.NewAccount(define.EthWalletPrivateSecret)
		logger.Infof("%d, addr:%s", i, acc.Address.Hex())

	}
	return account.Address.Hex(), nil
}
