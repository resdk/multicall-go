# multicall-go

Minimal golang ethereum multicall implementation, inspired by https://github.com/trayvox/go-eth-multicall, add some features:

* Support eth like chains, like polygon
* Support batch size, split big requests into small ones

# Example

```go
client, err := ethclient.Dial("https://polygon-rpc.com/")
if err != nil {
	log.Fatal(err)
}

caller := NewPolygon(client)

erc20Abi, _ := GetErc20Abi()

callData, _ := erc20Abi.Pack("symbol")

calls := []*Call{
	&Call{
		Id:       "WBTC",
		Target:   common.HexToAddress("0x1bfd67037b42cf73acf2047067bd4f2c47d9bfd6"),
		CallData: callData,
	},
	&Call{
		Id:       "USDC",
		Target:   common.HexToAddress("0x2791bca1f2de4661ed88a30c99a7a9449aa84174"),
		CallData: callData,
	},
}

results := caller.Execute(calls, 10)
for k, v := range results {
	out, _ := erc20Abi.Unpack("symbol", v.ReturnData)
	assert.Equal(t, k, out[0])
}
```
