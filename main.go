package main

import (
	"blockchain-query/cchelper"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	// "time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"

	// "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	// "github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var wg sync.WaitGroup

func handle(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintln(w, "All Done: ", r.RemoteAddr)
	os.Exit(0)
}

func calltx(client *channel.Client, cnt *string) {
	var args [][]byte
	args = append(args, []byte("a"))
	args = append(args, []byte("b"))
	args = append(args, []byte(*cnt))

	fmt.Println("args:", args)
	request := channel.Request{
		ChaincodeID: "mycc",
		Fcn:         "invoke",
		Args:        args,
	}
	response, err := client.Execute(request, channel.WithTargetEndpoints("peer0.org1.example.com", "peer0.org2.example.com"))
	if err != nil {
		log.Fatal("query fail: ", err.Error())
	} else {
		fmt.Printf("response is ChaincodeStatus[%s] TransactionID[%s] TxValidationCode[%s]\n", response.ChaincodeStatus, response.TransactionID, response.TxValidationCode)
	}
	wg.Done()
}

func calltx1(client *channel.Client, cnt *string) {
	var args [][]byte
	args = append(args, []byte("myvar"))
	args = append(args, []byte(*cnt))
	args = append(args, []byte("+"))

	// fmt.Println("args:", args)
	request := channel.Request{
		ChaincodeID: "bigdatacc",
		Fcn:         "update",
		Args:        args,
	}
	response, err := client.Execute(request, channel.WithTargetEndpoints("peer0.org1.example.com", "peer0.org2.example.com"))
	if err != nil {
		log.Fatal("update call fail: ", err.Error())
	} else {
		fmt.Printf("response is ChaincodeStatus[%s] TransactionID[%s] TxValidationCode[%s]\n", response.ChaincodeStatus, response.TransactionID, response.TxValidationCode)
	}
	wg.Done()
}

func main() {
	configPath := flag.String("config", "./config.yaml", "App config path")
	flag.StringVar(configPath, "c", "./config.yaml", "App config path short")
	blockId := flag.Uint64("id", 4, "Block ID which will display")
	transCount := flag.String("n", "18", "transfer number n point from a to b")
	batchSize := flag.Int("s", 1000, "Batch times  of invoke chaincode !")
	// port := flag.String("p", "3000", "Liston Port")
	flag.Parse()
	fmt.Println("Begin init")

	//读取配置文件，创建SDK
	configProvider := config.FromFile(*configPath)
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		log.Fatalf("create sdk fail: %s\n", err.Error())
	}

	//读取配置文件(config.yaml)中的组织(org1.example.com)的用户(Admin)
	mspClient, err := mspclient.New(sdk.Context(),
		mspclient.WithOrg("org1.example.com"))
	if err != nil {
		log.Fatalf("create msp client fail: %s\n", err.Error())
	}

	adminIdentity, err := mspClient.GetSigningIdentity("Admin")
	if err != nil {
		log.Fatalf("get admin identify fail: %s\n", err.Error())
	} else {
		fmt.Println("AdminIdentify is found:")
		fmt.Println(adminIdentity)
	}

	fmt.Println("Init with config.")

	//调用合约
	channelProvider := sdk.ChannelContext("chufan",
		fabsdk.WithUser("Admin"),
		fabsdk.WithOrg("org1.example.com"))
	fmt.Println("create channel context.")
	channelClient, err := channel.New(channelProvider)
	if err != nil {
		log.Fatalf("create channel client fail: %s\n", err.Error())
	}
	fmt.Println("new channel finished.")
	var args [][]byte

	//query
	args = append(args, []byte("a"))

	request := channel.Request{
		ChaincodeID: "mycc",
		Fcn:         "query",
		Args:        args,
	}
	response, err := channelClient.Query(request, channel.WithTargetEndpoints("peer0.org1.example.com"))
	if err != nil {
		log.Fatal("query fail: ", err.Error())
	} else {
		fmt.Printf("response is %s\n", response.Payload)
	}

	//ledger info query
	ledgerClient, err := ledger.New(channelProvider)
	if err != nil {
		fmt.Println("Failed to create new resource management client: %s", err)
	}

	ledgerInfo, err := ledgerClient.QueryInfo()
	if err != nil {
		fmt.Println("QueryInfo return error: %s", err)
	}

	fmt.Println(ledgerInfo)

	//call invoke
	x := *batchSize / 1000
	y := *batchSize % 1000
	for i := 0; i < x; i++ {
		wg = sync.WaitGroup{}
		for j := 0; j < 1000; j++ {
			wg.Add(1)
			go calltx1(channelClient, transCount)
		}
		fmt.Println("before wait")
		wg.Wait()
		fmt.Println("out ", i)
		// time.Sleep(3 * time.Second)
	}

	wg = sync.WaitGroup{}
	for k := 0; k < y; k++ {
		wg.Add(1)
		go calltx1(channelClient, transCount)
	}
	fmt.Println("before wait end")
	wg.Wait()

	//query
	args = nil
	args = append(args, []byte("a"))

	request = channel.Request{
		ChaincodeID: "mycc",
		Fcn:         "query",
		Args:        args,
	}
	response, err = channelClient.Query(request, channel.WithTargetEndpoints("peer0.org1.example.com"))
	if err != nil {
		log.Fatal("query fail: ", err.Error())
	} else {
		fmt.Printf("response is %s\n", response.Payload)
	}

	//ledger client test

	blockinfo, err := ledgerClient.QueryBlock(*blockId)
	if err != nil {
		fmt.Printf("QueryBlock return error: %s", err)
	} else {
		fmt.Println(blockinfo.String())
		fmt.Println("Block Data:", blockinfo.Data)
		fmt.Println("Block Get Data:", blockinfo.GetData())
		fmt.Println(cchelper.GetTransactionInfoFromData(blockinfo.GetData().Data[0], true))
	}

	//close resource
	defer sdk.Close()
	// time.Sleep(2 * time.Second)

	//spawn http
	http.HandleFunc("/", handle)
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", "3000"), nil)
}
