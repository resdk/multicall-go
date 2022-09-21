package multicall

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestCaller_Polygon_Execute(t *testing.T) {
	client, err := ethclient.Dial("https://polygon-rpc.com/")
	if err != nil {
		log.Fatal(err)
	}

	caller, _ := New(client)

	erc20Abi, _ := GetErc20Abi()

	callData, _ := erc20Abi.Pack("symbol")

	calls := []*Call{
		&Call{
			UserData: "WBTC",
			Target:   common.HexToAddress("0x1bfd67037b42cf73acf2047067bd4f2c47d9bfd6"),
			CallData: callData,
		},
		&Call{
			UserData: "USDC",
			Target:   common.HexToAddress("0x2791bca1f2de4661ed88a30c99a7a9449aa84174"),
			CallData: callData,
		},
	}

	results, _ := caller.Execute(calls, 10)
	for _, v := range results {
		out, _ := erc20Abi.Unpack("symbol", v.ReturnData)
		assert.Equal(t, v.UserData.(string), out[0])
	}
}
