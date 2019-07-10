package accounting

import (
	"testing"
)

//WorkersCountが正しく機能するか
func TestAccountingService_WorkersCount(t *testing.T) {
	accounting := NewService(true)
	count := accounting.GetWorkersCount()
	if count != 1 {
		t.Fatalf("failed counting workers\nexpect:1 result:%d", count)
	}
}

func TestAccountingService_SelectValidationWorkers(t *testing.T) {
	accounting := NewService(true)
	workers, err := accounting.SelectValidationWorkers(1)
	if err != nil {
		t.Fatalf("failed get validation workers%#v", err)
	}
	if len(workers) != 1 {
		t.Fatalf("the number of workers is not expected\nexpected:1 result:%d", len(workers))
	}
	for _, worker := range workers {
		if worker.Id == "" || worker.Addr == "" {
			t.Fatalf("gotten worker data is broken")
		}
	}
}

func TestAccountingService_SelectDBHolders(t *testing.T) {
	//TODO: Test Get DB Holder
}

func TestAccountingService_CreateNewWorkerSuccess(t *testing.T) {
	accounting := NewService(true)
	worker, err := accounting.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed create worker%#v", err)
	}
	if worker == nil {
		t.Fatalf("created worker is nil")
	}
}

func TestAccountingService_CreateDupulicatedWorkerFail(t *testing.T) {
	accounting := NewService(true)
	worker, err := accounting.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed create worker%#v", err)
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
		t.Fatalf("failed create worker%#v", err)
	}
	if client == nil {
		t.Fatalf("created worker is nil")
	}
}

func TestAccountingService_CreateDupulicatedClientFail(t *testing.T) {
	accounting := NewService(true)
	client, err := accounting.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed create worker%#v", err)
	}
	if client == nil {
		t.Fatalf("created worker is nil")
	}
	client, err = accounting.CreateNewWorker("worker:hoge", "hoge")
	if err != ErrIDAlreadyExists {
		t.Fatalf("couldn't get dupulication error %#v", err)
	}
}
