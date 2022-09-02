package multicall

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"strings"
)

type Caller struct {
	Client          *ethclient.Client
	ContractAddress *common.Address
	Abi             *abi.ABI
	Signer          *bind.TransactOpts
}

type Call struct {
	Id       string         `json:"id"`
	Target   common.Address `json:"target"`
	CallData []byte         `json:"call_data"`
}

type Response struct {
	Success    bool   `json:"success"`
	ReturnData []byte `json:"returnData"`
}

func (call *Call) GetMultiCall() *Multicall2Call {
	return &Multicall2Call{Target: call.Target, CallData: call.CallData}
}

func GetErc20Abi() (*abi.ABI, error) {
	erc20Abi, err := abi.JSON(strings.NewReader(Erc20ABI))
	if err != nil {
		return nil, err
	}
	return &erc20Abi, nil
}

func randomSigner() *bind.TransactOpts {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	signer, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1))
	if err != nil {
		panic(err)
	}

	signer.NoSend = true
	signer.Context = context.Background()
	signer.GasPrice = big.NewInt(0)

	return signer
}

func New(client *ethclient.Client, contractAddress *common.Address, mcAbi *abi.ABI, signer *bind.TransactOpts) *Caller {
	if mcAbi == nil {
		tmpAbi, err := abi.JSON(strings.NewReader(MultiCall2ABI))
		if err != nil {
			panic(err)
		}
		mcAbi = &tmpAbi
	}
	if signer == nil {
		signer = randomSigner()
	}

	return &Caller{
		Signer:          signer,
		Client:          client,
		Abi:             mcAbi,
		ContractAddress: contractAddress,
	}
}

// NewEth https://github.com/makerdao/multicall
func NewEth(client *ethclient.Client) *Caller {
	contractAddress := common.HexToAddress("0x5ba1e12693dc8f9c48aad8770482f4739beed696")
	return New(client, &contractAddress, nil, nil)
}

// NewPolygon https://github.com/makerdao/multicall/pull/24
func NewPolygon(client *ethclient.Client) *Caller {
	contractAddress := common.HexToAddress("0x275617327c958bD06b5D6b871E7f491D76113dd8")
	return New(client, &contractAddress, nil, nil)
}

func execute(caller *Caller, todos []Multicall2Call) []*Response {
	responses := make([]*Response, len(todos))

	callData, err := caller.Abi.Pack("tryAggregate", false, todos)
	if err != nil {
		panic(err)
	}

	resp, err := caller.Client.CallContract(
		context.Background(),
		ethereum.CallMsg{To: caller.ContractAddress, Data: callData},
		nil,
	)
	if err != nil {
		panic(err)
	}

	// Unpack results
	unpackedResp, _ := caller.Abi.Unpack("tryAggregate", resp)

	a, err := json.Marshal(unpackedResp[0])
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(a, &responses)
	if err != nil {
		panic(err)
	}

	return responses
}

func (caller *Caller) Execute(calls []*Call, batchSize int) map[string]*Response {
	responses := make([]*Response, 0, len(calls))

	todos := make([]Multicall2Call, 0, batchSize)

	for _, call := range calls {
		todos = append(todos, *call.GetMultiCall())
		if len(todos) >= batchSize {
			responses = append(responses, execute(caller, todos)...)
			todos = make([]Multicall2Call, 0, batchSize)
		}
	}

	// deal with left todos
	if len(todos) > 0 {
		responses = append(responses, execute(caller, todos)...)
	}

	results := make(map[string]*Response, len(calls))
	for i, response := range responses {
		results[calls[i].Id] = response
	}

	return results
}
