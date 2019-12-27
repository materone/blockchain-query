// chaincode.g
package cchelper

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	// putils "github.com/hyperledger/fabric/protos/utils"
)

// TransactionDetail获取了交易的具体信息
type TransactionDetail struct {
	TransactionId string
	CreateTime    string
	Args          []string
}

// GetChaincodeAction gets the ChaincodeAction given chaicnode action bytes
func GetChaincodeAction(caBytes []byte) (*peer.ChaincodeAction, error) {
	chaincodeAction := &peer.ChaincodeAction{}
	err := proto.Unmarshal(caBytes, chaincodeAction)
	return chaincodeAction, errors.Wrap(err, "error unmarshaling ChaincodeAction")
}

// GetChaincodeActionPayload Get ChaincodeActionPayload from bytes
func GetChaincodeActionPayload(capBytes []byte) (*peer.ChaincodeActionPayload, error) {
	cap := &peer.ChaincodeActionPayload{}
	err := proto.Unmarshal(capBytes, cap)
	return cap, errors.Wrap(err, "error unmarshaling ChaincodeActionPayload")
}

// GetChaincodeProposalPayload Get ChaincodeProposalPayload from bytes
func GetChaincodeProposalPayload(bytes []byte) (*peer.ChaincodeProposalPayload, error) {
	cpp := &peer.ChaincodeProposalPayload{}
	err := proto.Unmarshal(bytes, cpp)
	return cpp, errors.Wrap(err, "error unmarshaling ChaincodeProposalPayload")
}

// GetProposalResponsePayload gets the proposal response payload
func GetProposalResponsePayload(prpBytes []byte) (*peer.ProposalResponsePayload, error) {
	prp := &peer.ProposalResponsePayload{}
	err := proto.Unmarshal(prpBytes, prp)
	return prp, errors.Wrap(err, "error unmarshaling ProposalResponsePayload")
}

// GetPayload Get Payload from Envelope message
func GetPayload(e *common.Envelope) (*common.Payload, error) {
	payload := &common.Payload{}
	err := proto.Unmarshal(e.Payload, payload)
	return payload, errors.Wrap(err, "error unmarshaling Payload")
}

func GetEnvelopeFromBlock(data []byte) (*common.Envelope, error) {
	// Block always begins with an envelope
	var err error
	env := &common.Envelope{}
	if err = proto.Unmarshal(data, env); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling Envelope")
	}

	return env, nil
}

// GetTransaction Get Transaction from bytes
func GetTransaction(txBytes []byte) (*peer.Transaction, error) {
	tx := &peer.Transaction{}
	err := proto.Unmarshal(txBytes, tx)
	return tx, errors.Wrap(err, "error unmarshaling Transaction")

}

// 从SDK中Block.BlockDara.Data中提取交易具体信息
func GetTransactionInfoFromData(data []byte, needArgs bool) (*TransactionDetail, error) {
	// proto.Bool(true)
	// channelHeader := &common.ChannelHeader{}
	// propPayload := &peer.ChaincodeProposalPayload{}
	// tx, err := GetEnvelopeFromBlock(data)
	// errors.Wrap(err, "error extracting Envelope from block")
	// fmt.Println(channelHeader, propPayload, tx, err)
	fmt.Println("info in GetTransactionInfoFromData")

	env, err := GetEnvelopeFromBlock(data)
	if err != nil {
		return nil, errors.Wrap(err, "error extracting Envelope from block")
	}
	if env == nil {
		return nil, errors.New("nil envelope")
	}
	payload, err := GetPayload(env)
	if err != nil {
		return nil, errors.Wrap(err, "error extracting Payload from envelope")
	}
	channelHeaderBytes := payload.Header.ChannelHeader
	channelHeader := &common.ChannelHeader{}
	if err := proto.Unmarshal(channelHeaderBytes, channelHeader); err != nil {
		return nil, errors.Wrap(err, "error extracting ChannelHeader from payload")
	}
	var (
		args []string
	)
	if needArgs {
		tx, err := GetTransaction(payload.Data)
		if err != nil {
			return nil, errors.Wrap(err, "error unmarshalling transaction payload")
		}

		// chainCodeAction, err := GetChaincodeAction(data)

		chaincodeActionPayload, err := GetChaincodeActionPayload(tx.Actions[0].Payload)
		if err != nil {
			return nil, errors.Wrap(err, "error unmarshalling chaincode action payload")
		}
		propPayload := &peer.ChaincodeProposalPayload{}
		if err := proto.Unmarshal(chaincodeActionPayload.ChaincodeProposalPayload, propPayload); err != nil {
			return nil, errors.Wrap(err, "error extracting ChannelHeader from payload")
		}
		invokeSpec := &peer.ChaincodeInvocationSpec{}
		err = proto.Unmarshal(propPayload.Input, invokeSpec)
		if err != nil {
			return nil, errors.Wrap(err, "error extracting ChannelHeader from payload")
		}
		for _, v := range invokeSpec.ChaincodeSpec.Input.Args {
			args = append(args, string(v))
		}
		//Action rwset
		propResPayload := &peer.ProposalResponsePayload{}
		if err := proto.Unmarshal(chaincodeActionPayload.Action.ProposalResponsePayload, propResPayload); err != nil {
			return nil, errors.Wrap(err, "error extracting ProposalResponsePayload from payload")
		}
		ccAction := &peer.ChaincodeAction{}
		if err := proto.Unmarshal(propResPayload.Extension, ccAction); err != nil {
			return nil, errors.Wrap(err, "error extracting ChaincodeAction from payload")
		}
		txRWS := &rwset.TxReadWriteSet{}
		if err := proto.Unmarshal(ccAction.Results, txRWS); err != nil {
			return nil, errors.Wrap(err, "error extracting TxReadWriteSet from payload")
		}
		fmt.Println("Output nsrws")
		kvRWS := &kvrwset.KVRWSet{}
		for _, v := range txRWS.NsRwset {
			if err := proto.Unmarshal(v.Rwset, kvRWS); err != nil {
				return nil, errors.Wrap(err, "error extracting NsReadWriteSet from payload")
			}
			// fmt.Printf("---%s", kvRWS)
			for _, v1 := range kvRWS.Writes {
				fmt.Printf("Key:%s isDelete:%t Value:%s\n", v1.Key, v1.IsDelete, v1.Value)
			}
		}
		// ccAction.Results
		// fmt.Printf("%s", propResPayload.Extension)

		// fmt.Println(chainCodeAction.GetResults())
	}
	result := &TransactionDetail{
		TransactionId: channelHeader.TxId,
		Args:          args,
		CreateTime:    time.Unix(channelHeader.Timestamp.Seconds, 0).Format("2006-01-02 15:04:05"),
	}
	return result, nil
}
