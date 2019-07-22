package accounting_test

import (
	"testing"
	"yox2yox/antone/bridge/accounting"
)

var (
	testWorkersId = []string{
		"worker0",
		"worker1",
	}
	testWorkerAddr = "addr"
	testUserId     = "client0"
)

//WorkersCountが正しく機能するか
func TestAccountingService_WorkersCount(t *testing.T) {
	accounting := accounting.NewService(true)
	_, err := accounting.CreateNewWorker(testWorkersId[0], testWorkerAddr)
	if err != nil {
		t.Fatalf("failed to create worker %#v", err)
	}
	count := accounting.GetWorkersCount()
	if count != 1 {
		t.Fatalf("failed to count workers\nexpect:1 result:%d", count)
	}
}

func TestAccountingService_SelectValidationWorkers(t *testing.T) {
	accounting := accounting.NewService(true)

	_, err := accounting.CreateNewWorker(testWorkersId[0], testWorkerAddr)
	if err != nil {
		t.Fatalf("failed to create worker %#v", err)
	}

	workers, err := accounting.SelectValidationWorkers(1)
	if err != nil {
		t.Fatalf("failed to get validation workers%#v", err)
	}
	if len(workers) != 1 {
		t.Fatalf("the number of workers is not expected\nexpected:1 result:%d", len(workers))
	}
	for _, worker := range workers {
		if worker.Id == "" || worker.Addr == "" {
			t.Fatalf("gotten workerdata is broken")
		}
	}

	_, err = accounting.CreateNewWorker(testWorkersId[1], testWorkerAddr)
	if err != nil {
		t.Fatalf("failed to create worker %#v", err)
	}

	workers, err = accounting.SelectValidationWorkers(2)
	if err != nil {
		t.Fatalf("failed to get validation workers%#v", err)
	}
	if len(workers) != 2 {
		t.Fatalf("the number of workers is not expected\nexpected:2 result:%d", len(workers))
	}
	for _, worker := range workers {
		if worker.Id == "" || worker.Addr == "" {
			t.Fatalf("gotten workerdata is broken")
		}
	}

}

func TestAccountingService_SelectValidationWorkersFail(t *testing.T) {
	accountingS := accounting.NewService(true)

	_, err := accountingS.SelectValidationWorkers(1)
	if err != accounting.ErrWorkersAreNotEnough {
		t.Fatalf("failed to get NotEnough error%#v", err)
	}

	_, err = accountingS.CreateNewWorker(testWorkersId[1], testWorkerAddr)
	if err != nil {
		t.Fatalf("failed to create worker %#v", err)
	}

	_, err = accountingS.SelectValidationWorkers(2)
	if err != accounting.ErrWorkersAreNotEnough {
		t.Fatalf("failed to get NotEnough error%#v", err)
	}

}

func Test_RegistarNewDataPoolHolders(t *testing.T) {

	testHoldersNum := 1

	accountingS := accounting.NewService(true)
	_, err := accountingS.CreateDatapoolAndSelectHolders(testUserId, testHoldersNum)
	if err != accounting.ErrWorkersAreNotEnough {
		t.Fatalf("want has error %#v, but %#v", accounting.ErrWorkersAreNotEnough, err)
	}

	_, err = accountingS.CreateNewWorker(testWorkersId[0], testWorkerAddr)
	_, err = accountingS.CreateDatapoolAndSelectHolders(testUserId, testHoldersNum)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	_, err = accountingS.CreateDatapoolAndSelectHolders(testUserId, testHoldersNum)
	if err != accounting.ErrDataPoolAlreadyExists {
		t.Fatalf("want has error %#v, but %#v", accounting.ErrDataPoolAlreadyExists, err)
	}

}

func TestAccountingService_SelectDataPoolHolders(t *testing.T) {
	accountingS := accounting.NewService(true)

	holder, err := accountingS.SelectDataPoolHolder(testUserId)
	if err != accounting.ErrDataPoolHolderNotExist {
		t.Fatalf("want has error %#v, but %#v", accounting.ErrDataPoolHolderNotExist, err)
	}

	_, err = accountingS.CreateNewWorker(testWorkersId[0], testWorkerAddr)
	_, err = accountingS.CreateDatapoolAndSelectHolders(testUserId, 1)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}

	holder, err = accountingS.SelectDataPoolHolder(testUserId)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	if holder == nil {
		t.Fatalf("want holder not nil, but nil")
	}

}

func TestAccountingService_CreateNewWorkerSuccess(t *testing.T) {
	accounting := accounting.NewService(true)
	worker, err := accounting.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed to create worker%#v", err)
	}
	if worker == nil {
		t.Fatalf("created worker is nil")
	}
}

func TestAccountingService_CreateDupulicatedWorkerFail(t *testing.T) {
	accountingS := accounting.NewService(true)
	worker, err := accountingS.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed to create worker%#v", err)
	}
	if worker == nil {
		t.Fatalf("created worker is nil")
	}
	worker, err = accountingS.CreateNewWorker("worker:hoge", "hoge")
	if err != accounting.ErrIDAlreadyExists {
		t.Fatalf("couldn't get dupulication error %#v", err)
	}
}

func TestAccountingService_CreateNewClientSuccess(t *testing.T) {
	accounting := accounting.NewService(true)
	client, err := accounting.CreateNewClient("worker:hoge")
	if err != nil {
		t.Fatalf("failed to create worker%#v", err)
	}
	if client == nil {
		t.Fatalf("created worker is nil")
	}
}

func TestAccountingService_CreateDupulicatedClientFail(t *testing.T) {
	accountingS := accounting.NewService(true)
	client, err := accountingS.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed to create worker%#v", err)
	}
	if client == nil {
		t.Fatalf("created worker is nil")
	}
	client, err = accountingS.CreateNewWorker("worker:hoge", "hoge")
	if err != accounting.ErrIDAlreadyExists {
		t.Fatalf("couldn't get dupulication error %#v", err)
	}
}
