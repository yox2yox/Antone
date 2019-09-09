package test

import (
	"context"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
	"yox2yox/antone/bridge"
	bConfig "yox2yox/antone/bridge/config"
	bpb "yox2yox/antone/bridge/pb"
	"yox2yox/antone/internal/log2"
	wPeer "yox2yox/antone/worker"
	wConfig "yox2yox/antone/worker/config"

	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	// パッケージ内のテストの実行
	code := m.Run()
	// 終了処理
	log2.Close()
	// テストの終了コードで exit
	os.Exit(code)
}

func Test_CreateNetWork(t *testing.T) {

	//通常クライアントId
	clientId := "client0"
	var clientData int32 = 0
	//ホルダーを複数持つクライアントId
	clientIdMultiHolder := "clientmh"
	var mhclientData int32 = 0
	workerscount := 6
	workersAddr := "127.0.0.1"
	workersBasePort := 10001
	wg := sync.WaitGroup{}

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
	wg.Add(1)
	go func() {
		err = bpeer.Run(ctxBridge)
		if err != nil {
			t.Fatalf("want no error, but has error %#v", err)
		}
		wg.Done()
	}()

	//ワーカPeer起動
	for i := 0; i < workerscount; i++ {
		workerConfig, err := wConfig.ReadWorkerConfig()
		if err != nil {
			t.Fatalf("want no error, but has error %#v", err)
		}
		workerConfig.Server.Addr = workersAddr + ":" + strconv.Itoa(workersBasePort+i)
		wpeer, err := wPeer.New(workerConfig, false, false)
		if err != nil {
			t.Log(workerConfig.Server.Addr)
			t.Fatalf("want no error, but has error %#v", err)
		}
		wpeers = append(wpeers, wpeer)
		ctxworker, cancelworker := context.WithTimeout(context.Background(), 10*time.Second)
		wg.Add(1)
		go func() {
			err = wpeer.Run(ctxworker)
			if err != nil {
				t.Fatalf("want no error, but has error %#v", err)
			}
			wg.Done()
		}()
		defer cancelworker()
	}

	time.Sleep(1 * time.Second)
	if bpeer.Accounting.GetWorkersCount() != workerscount {
		t.Fatalf("want workers count is %d, but %d", workerscount, bpeer.Accounting.GetWorkersCount())
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
	_, err = bpeer.Accounting.CreateNewClient(clientId)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	createdDp, err := bpeer.Datapool.CreateDatapool(clientId, 1)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}

	//データが登録されているかチェック
	countholder := 0
	for _, wpeer := range wpeers {
		data, err := wpeer.DataPool.GetDataPool(createdDp.DatapoolId)
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

	//validatableコード取得
	vCodeRequest := &bpb.ValidatableCodeRequest{Datapoolid: createdDp.DatapoolId, Add: 1}
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
		data, err := wpeer.DataPool.GetDataPool(createdDp.DatapoolId)
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
	vCodeRequest = &bpb.ValidatableCodeRequest{Datapoolid: createdDp.DatapoolId, Add: 1}
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
	vCodeRequest = &bpb.ValidatableCodeRequest{Datapoolid: createdDp.DatapoolId, Add: 2}
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
	vCodeRequest = &bpb.ValidatableCodeRequest{Datapoolid: createdDp.DatapoolId, Add: 3}
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
		data, err := wpeer.DataPool.GetDataPool(createdDp.DatapoolId)
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
	//ホルダー作成(4台)
	mhDatapool, err := bpeer.Datapool.CreateDatapool(clientIdMultiHolder, 4)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}

	//マルチホルダーのデータプールが作成されているかチェック
	countholder = 0
	for _, wpeer := range wpeers {
		data, err := wpeer.DataPool.GetDataPool(mhDatapool.DatapoolId)
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
	vCodeRequest = &bpb.ValidatableCodeRequest{Datapoolid: mhDatapool.DatapoolId, Add: 1}
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
		data, err := wpeer.DataPool.GetDataPool(mhDatapool.DatapoolId)
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
	vCodeRequest = &bpb.ValidatableCodeRequest{Datapoolid: createdDp.DatapoolId, Add: 1}
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
	vCodeRequest = &bpb.ValidatableCodeRequest{Datapoolid: mhDatapool.DatapoolId, Add: 2}
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
	vCodeRequest = &bpb.ValidatableCodeRequest{Datapoolid: createdDp.DatapoolId, Add: 3}
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
	vCodeRequest = &bpb.ValidatableCodeRequest{Datapoolid: mhDatapool.DatapoolId, Add: 4}
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
		data, err := wpeer.DataPool.GetDataPool(createdDp.DatapoolId)
		if err == nil {
			countholder += 1
			if data != clientData {
				t.Fatalf("want user's data==%d,but %#v", clientData, data)
			}
		}
		data, err = wpeer.DataPool.GetDataPool(mhDatapool.DatapoolId)
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

	bpeer.Close()
	cancelBridge()
	for _, wpeer := range wpeers {
		wpeer.Close()
	}
	wg.Wait()

}
