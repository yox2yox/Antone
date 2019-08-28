package test

import (
	"context"
	"strconv"
	"testing"
	"time"
	"yox2yox/antone/bridge"
	bConfig "yox2yox/antone/bridge/config"
	bpb "yox2yox/antone/bridge/pb"
	wPeer "yox2yox/antone/worker"
	wConfig "yox2yox/antone/worker/config"

	"google.golang.org/grpc"
)

func Test_CreateNetWork(t *testing.T) {

	clientId := "client0"
	workersId := []string{
		"worker0",
		"worker1",
		"worker2",
		"worker3",
		"worker4",
		"worker5",
	}
	workersAddr := "localhost"
	workersBasePort := 10000

	var wpeers []*wPeer.Peer

	//ブリッジPeer起動
	bridgeConfig, err := bConfig.ReadBridgeConfig()
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	bpeer, err := bridge.New(bridgeConfig, false)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	ctxBridge, cancelBridge := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		err = bpeer.Run(ctxBridge)
		if err != nil {
			t.Fatalf("want no error, but has error %#v", err)
		}
	}()
	defer cancelBridge()

	//ワーカPeer起動
	workerConfig, err := wConfig.ReadWorkerConfig()
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	for i, _ := range workersId {
		workerConfig.Server.Addr = workersAddr + ":" + strconv.Itoa(workersBasePort+i)
		wpeer, err := wPeer.New(workerConfig.Server, false)
		if err != nil {
			t.Fatalf("want no error, but has error %#v", err)
		}
		wpeers = append(wpeers, wpeer)
		ctxworker, cancelworker := context.WithTimeout(context.Background(), 10*time.Second)
		go func() {
			err = wpeer.Run(ctxworker)
			if err != nil {
				t.Fatalf("want no error, but has error %#v", err)
			}
		}()
		defer cancelworker()
	}

	//ブリッジクライアント作成
	connBridge, err := grpc.Dial(bpeer.Config.Server.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer connBridge.Close()
	clientOrder := bpb.NewOrdersClient(connBridge)
	ctxOrder, orderCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer orderCancel()

	//ワーカクライアント作成(未使用)
	// connWorker, err := grpc.Dial(workerConfig.Server.Addr, grpc.WithInsecure())
	// if err != nil {
	// 	t.Fatalf("want no error,but error %#v", err)
	// }
	// defer connWorker.Close()

	//clientWorker := wpb.NewWorkerClient(connWorker)
	//ctxWorker, workerCancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer workerCancel()

	//各種アカウント登録
	for i, workerId := range workersId {
		_, err = bpeer.Accounting.CreateNewWorker(workerId, wpeers[i].ServerConfig.Addr)
		if err != nil {
			t.Fatalf("want no error,but error %#v", err)
		}
	}

	_, err = bpeer.Accounting.CreateNewClient(clientId)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	for _, wpeer := range wpeers {
		data, err := wpeer.DataPool.GetDataPool(clientId)
		if err != nil {
			t.Fatalf("want no error,but error %#v", err)
		}
		if data != 0 {
			t.Fatalf("want user's data==0,but %#v", data)
		}
	}

	t.Log("Start to get validatable code")

	//validatableコード取得
	vCodeRequest := &bpb.ValidatableCodeRequest{Userid: clientId, Add: 1}
	vCode, err := clientOrder.RequestValidatableCode(ctxOrder, vCodeRequest)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if vCode == nil {
		t.Fatalf("want validatable code is not nil,but nil")
	}
	if vCode.Data != 0 {
		t.Fatalf("want vcode.data == 0,but %#v", vCode.Data)
	}

	time.Sleep(3 * time.Second)

}
