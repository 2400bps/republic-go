package arc

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/republicprotocol/republic-go/blockchain/bitcoin"
)

func refund(conn bitcoin.Conn, contract, contractTxBytes []byte) error {

	var contractTx wire.MsgTx
	err := contractTx.Deserialize(bytes.NewReader(contractTxBytes))
	if err != nil {
		return fmt.Errorf("failed to decode contract transaction: %v", err)
	}

	pushes, err := txscript.ExtractAtomicSwapDataPushes(0, contract)
	if err != nil {
		return err
	}
	if pushes == nil {
		return errors.New("contract is not an atomic swap script recognized by this tool")
	}

	refundTx, err := buildRefund(conn, contract, &contractTx)
	if err != nil {
		return err
	}

	txHash, err := conn.PromptPublishTx(refundTx, "refund")
	if err != nil {
		return err
	}

	conn.WaitForConfirmations(txHash, 1)

	return nil
}
