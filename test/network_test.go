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

	//通常クライアントId
	clientId := "client0"
	var clientData int32 = 0
	//ホルダーを複数持つクライアントId
	clientIdMultiHolder := "clientmh"
	var mhclientData int32 = 0

	workersId := []string{
		"worker0",
		"worker1",
		"worker2",
		"worker3",
		"worker4",
		"worker5",
	}
	workersAddr := "127.0.0.1"
	workersBasePort := 10001

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
	defer bpeer.Close()
	defer cancelBridge()

	//ワーカPeer起動
	for i, _ := range workersId {
		workerConfig, err := wConfig.ReadWorkerConfig()
		if err != nil {
			t.Fatalf("want no error, but has error %#v", err)
		}
		workerConfig.Server.Addr = workersAddr + ":" + strconv.Itoa(workersBasePort+i)
		wpeer, err := wPeer.New(workerConfig, false)
		if err != nil {
			t.Log(workerConfig.Server.Addr)
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
		defer wpeer.Close()
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
		_, err = bpeer.Accounting.CreateNewWorker(workerId, wpeers[i].WorkerConfig.Server.Addr)
		if err != nil {
			t.Fatalf("want no error,but error %#v", err)
		}
	}

	_, err = bpeer.Accounting.CreateNewClient(clientId)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}

	//データが登録されているかチェック
	countholder := 0
	for _, wpeer := range wpeers {
		data, err := wpeer.DataPool.GetDataPool(clientId)
		if err == nil {
			countholder += 1
			if data != 0 {
				t.Fatalf("want user's data==0,but %#v", data)
			}
		}
	}
	if countholder < 1 {
		t.Fatalf("want holders count is 1 ,but %#v", countholder)
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
	clientData += 1

	time.Sleep(3 * time.Second)

	//データが更新されているかチェック
	countholder = 0
	for _, wpeer := range wpeers {
		data, err := wpeer.DataPool.GetDataPool(clientId)
		if err == nil {
			countholder += 1
			if data != clientData {
				t.Fatalf("want user's data==%d,but %#v", clientData, data)
			}
		}
	}
	if countholder < 1 {
		t.Fatalf("want holders count is 1 ,but %#v", countholder)
	}

	//加算連続
	//Data += 1
	vCodeRequest = &bpb.ValidatableCodeRequest{Userid: clientId, Add: 1}
	vCode, err = clientOrder.RequestValidatableCode(ctxOrder, vCodeRequest)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if vCode == nil {
		t.Fatalf("want validatable code is not nil,but nil")
	}
	if vCode.Data != clientData {
		t.Fatalf("want vcode.data == %d,but %#v", clientData, vCode.Data)
	}
	clientData += 1

	//Data += 2
	vCodeRequest = &bpb.ValidatableCodeRequest{Userid: clientId, Add: 2}
	vCode, err = clientOrder.RequestValidatableCode(ctxOrder, vCodeRequest)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if vCode == nil {
		t.Fatalf("want validatable code is not nil,but nil")
	}
	if vCode.Data != clientData {
		t.Fatalf("want vcode.data == %d,but %#v", clientData, vCode.Data)
	}
	clientData += 2

	//Data += 3
	vCodeRequest = &bpb.ValidatableCodeRequest{Userid: clientId, Add: 3}
	vCode, err = clientOrder.RequestValidatableCode(ctxOrder, vCodeRequest)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if vCode == nil {
		t.Fatalf("want validatable code is not nil,but nil")
	}
	if vCode.Data != clientData {
		t.Fatalf("want vcode.data == %d,but %#v", clientData, vCode.Data)
	}
	clientData += 3

	//バリデーションが完了するまで待機
	time.Sleep(3 * time.Second)

	//データが更新されているかチェック
	countholder = 0
	for _, wpeer := range wpeers {
		data, err := wpeer.DataPool.GetDataPool(clientId)
		if err == nil {
			countholder += 1
			if data != clientData {
				t.Fatalf("want user's data==%d,but %#v", clientData, data)
			}
		}
	}
	if countholder < 1 {
		t.Fatalf("want holders count is 1 ,but %#v", countholder)
	}

	//マルチホルダークライアント作成
	_, err = bpeer.Accounting.CreateNewClient(clientIdMultiHolder)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	//ホルダー作成(3台)
	_, err = bpeer.Accounting.CreateDatapoolAndSelectHolders(clientIdMultiHolder, 0, 3)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}

	//マルチホルダーのデータプールが作成されているかチェック
	countholder = 0
	for _, wpeer := range wpeers {
		data, err := wpeer.DataPool.GetDataPool(clientIdMultiHolder)
		if err == nil {
			countholder += 1
			if data != 0 {
				t.Fatalf("want user's data==0, but %#v", data)
			}
		}
	}
	if countholder != 4 {
		t.Fatalf("want holders count is 4, but %#v", countholder)
	}

	//マルチホルダークライアントのvalidatableコード取得
	//mhData += 1
	vCodeRequest = &bpb.ValidatableCodeRequest{Userid: clientIdMultiHolder, Add: 1}
	vCode, err = clientOrder.RequestValidatableCode(ctxOrder, vCodeRequest)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if vCode == nil {
		t.Fatalf("want validatable code is not nil,but nil")
	}
	if vCode.Data != 0 {
		t.Fatalf("want vcode.data == 0,but %#v", vCode.Data)
	}
	mhclientData += 1

	time.Sleep(3 * time.Second)

	//データが更新されているかチェック
	countholder = 0
	for _, wpeer := range wpeers {
		data, err := wpeer.DataPool.GetDataPool(clientIdMultiHolder)
		if err == nil {
			countholder += 1
			if data != mhclientData {
				t.Fatalf("want user's data==%d,but %#v", mhclientData, data)
			}
		}
	}
	if countholder != 4 {
		t.Fatalf("want holders count is 4,but %#v", countholder)
	}

	//加算連続(複数クライアント)
	//Data += 1
	vCodeRequest = &bpb.ValidatableCodeRequest{Userid: clientId, Add: 1}
	vCode, err = clientOrder.RequestValidatableCode(ctxOrder, vCodeRequest)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if vCode == nil {
		t.Fatalf("want validatable code is not nil,but nil")
	}
	if vCode.Data != clientData {
		t.Fatalf("want vcode.data == %d,but %#v", clientData, vCode.Data)
	}
	clientData += 1

	//mhData += 2
	vCodeRequest = &bpb.ValidatableCodeRequest{Userid: clientIdMultiHolder, Add: 2}
	vCode, err = clientOrder.RequestValidatableCode(ctxOrder, vCodeRequest)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if vCode == nil {
		t.Fatalf("want validatable code is not nil,but nil")
	}
	if vCode.Data != mhclientData {
		t.Fatalf("want vcode.data == %d,but %#v", mhclientData, vCode.Data)
	}
	mhclientData += 2

	//Data+=3
	vCodeRequest = &bpb.ValidatableCodeRequest{Userid: clientId, Add: 3}
	vCode, err = clientOrder.RequestValidatableCode(ctxOrder, vCodeRequest)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if vCode == nil {
		t.Fatalf("want validatable code is not nil,but nil")
	}
	if vCode.Data != clientData {
		t.Fatalf("want vcode.data == %d,but %#v", clientData, vCode.Data)
	}
	clientData += 3

	//mhData+=4
	vCodeRequest = &bpb.ValidatableCodeRequest{Userid: clientIdMultiHolder, Add: 4}
	vCode, err = clientOrder.RequestValidatableCode(ctxOrder, vCodeRequest)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if vCode == nil {
		t.Fatalf("want validatable code is not nil,but nil")
	}
	if vCode.Data != mhclientData {
		t.Fatalf("want vcode.data == %d,but %#v", mhclientData, vCode.Data)
	}
	mhclientData += 4

	//バリデーションが完了するまで待機
	time.Sleep(3 * time.Second)

	//データが更新されているかチェック
	countholder = 0
	for _, wpeer := range wpeers {
		data, err := wpeer.DataPool.GetDataPool(clientId)
		if err == nil {
			countholder += 1
			if data != clientData {
				t.Fatalf("want user's data==%d,but %#v", clientData, data)
			}
		}
		data, err = wpeer.DataPool.GetDataPool(clientIdMultiHolder)
		if err == nil {
			countholder += 1
			if data != mhclientData {
				t.Fatalf("want user's data==%d,but %#v", mhclientData, data)
			}
		}
	}
	if countholder < 1 {
		t.Fatalf("want holders count is 1 ,but %#v", countholder)
	}

}
