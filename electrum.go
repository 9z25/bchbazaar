package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchd/txscript"
	"github.com/gcash/bchutil"
)

const ElectrumWSS = "wss://bch.imaginary.cash:50004"

type ElectrumRequest struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type ElectrumResponse struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  interface{}     `json:"error"`
}

type UTXO struct {
	TxHash string `json:"tx_hash"`
	TxPos  uint32 `json:"tx_pos"`
	Value  int64  `json:"value"`
	Height int64  `json:"height"`
}

func addressToScriptHash(addr string) (string, error) {
	addr = strings.TrimPrefix(addr, "bitcoincash:")

	decoded, err := bchutil.DecodeAddress(addr, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}

	script, err := txscript.PayToAddrScript(decoded)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(script)

	// Electrum requires reversed sha256(scriptPubKey)
	for i, j := 0, len(hash)-1; i < j; i, j = i+1, j-1 {
		hash[i], hash[j] = hash[j], hash[i]
	}

	return hex.EncodeToString(hash[:]), nil
}

func electrumListUnspent(address string) ([]UTXO, error) {
	scripthash, err := addressToScriptHash(address)
	if err != nil {
		return nil, err
	}

	conn, _, err := websocket.DefaultDialer.Dial(ElectrumWSS, nil)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := ElectrumRequest{
		ID:     1,
		Method: "blockchain.scripthash.listunspent",
		Params: []interface{}{scripthash},
	}

	if err := conn.WriteJSON(req); err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	var resp ElectrumResponse
	if err := conn.ReadJSON(&resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("electrum error: %v", resp.Error)
	}

	var utxos []UTXO
	if err := json.Unmarshal(resp.Result, &utxos); err != nil {
		return nil, err
	}

	return utxos, nil
}

func checkBCHPayment(address string, amountBCH float64) (bool, string, error) {
	utxos, err := electrumListUnspent(address)
	if err != nil {
		return false, "", err
	}

	requiredSats := int64(math.Round(amountBCH * 100000000))

	for _, u := range utxos {
		if u.Value >= requiredSats {
			return true, u.TxHash, nil
		}
	}

	return false, "", nil
}