package accounting

import (
	"testing"
)

var (
	testWorkersId = []string{
		"worker0",
		"worker1",
	}
	testWorkerAddr = "addr"
)

//WorkersCountが正しく機能するか
func TestAccountingService_WorkersCount(t *testing.T) {
	accounting := NewService(true)
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
	accounting := NewService(true)

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
	accountingS := NewService(true)

	_, err := accountingS.SelectValidationWorkers(1)
	if err != ErrWorkersAreNotEnough {
		t.Fatalf("failed to get NotEnough error%#v", err)
	}

	_, err = accountingS.CreateNewWorker(testWorkersId[1], testWorkerAddr)
	if err != nil {
		t.Fatalf("failed to create worker %#v", err)
	}

	_, err = accountingS.SelectValidationWorkers(2)
	if err != ErrWorkersAreNotEnough {
		t.Fatalf("failed to get NotEnough error%#v", err)
	}

}

func TestAccountingService_SelectDBHolders(t *testing.T) {
	//TODO: Test Get DB Holder
}

func TestAccountingService_CreateNewWorkerSuccess(t *testing.T) {
	accounting := NewService(true)
	worker, err := accounting.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed to create worker%#v", err)
	}
	if worker == nil {
		t.Fatalf("created worker is nil")
	}
}

func TestAccountingService_CreateDupulicatedWorkerFail(t *testing.T) {
	accounting := NewService(true)
	worker, err := accounting.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed to create worker%#v", err)
	}
	if worker == nil {
		t.Fatalf("created worker is nil")
	}
	worker, err = accounting.CreateNewWorker("worker:hoge", "hoge")
	if err != ErrIDAlreadyExists {
		t.Fatalf("couldn't get dupulication error %#v", err)
	}
}

func TestAccountingService_CreateNewClientSuccess(t *testing.T) {
	accounting := NewService(true)
	client, err := accounting.CreateNewClient("worker:hoge")
	if err != nil {
		t.Fatalf("failed to create worker%#v", err)
	}
	if client == nil {
		t.Fatalf("created worker is nil")
	}
}

func TestAccountingService_CreateDupulicatedClientFail(t *testing.T) {
	accounting := NewService(true)
	client, err := accounting.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed to create worker%#v", err)
	}
	if client == nil {
		t.Fatalf("created worker is nil")
	}
	client, err = accounting.CreateNewWorker("worker:hoge", "hoge")
	if err != ErrIDAlreadyExists {
		t.Fatalf("couldn't get dupulication error %#v", err)
	}
}
