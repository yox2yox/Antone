package datapool_test

import (
	"testing"
	"os"
	"yox2yox/antone/worker/datapool"
	"yox2yox/antone/internal/log2"
)

var (
	testUserId = "client0"
)

func TestMain(m *testing.M) {
	// パッケージ内のテストの実行
	code := m.Run()
	// 終了処理
	log2.Close()
	// テストの終了コードで exit
	os.Exit(code)
}

func Test_InitDataPool(t *testing.T) {
	datapool := datapool.NewService()
	if datapool == nil {
		t.Fatalf("want datapool is not nil, but nil")
	}
}

func Test_CreateNewDataPool(t *testing.T) {
	dp := datapool.NewService()
	if dp == nil {
		t.Fatalf("want datapool is not nil, but nil")
	}
	err := dp.CreateDataPool(testUserId, 0)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	err = dp.CreateDataPool(testUserId, 0)
	if err != datapool.ErrDataPoolAlreadyExist {
		t.Fatalf("want has error 'ErrDataPoolAlreadyExist', but has error %#v", err)
	}
}

func Test_GetDataPool(t *testing.T) {

	dp := datapool.NewService()
	if dp == nil {
		t.Fatalf("want datapool is not nil, but nil")
	}
	_, err := dp.GetDataPool(testUserId)
	if err != datapool.ErrDataPoolNotExist {
		t.Fatalf("want has error 'ErrDataPoolNotExist', but has error %#v", err)
	}

	err = dp.CreateDataPool(testUserId, 0)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	_, err = dp.GetDataPool(testUserId)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}

}

func Test_SetDataPool(t *testing.T) {
	dp := datapool.NewService()
	if dp == nil {
		t.Fatalf("want datapool is not nil, but nil")
	}

	err := dp.SetDataPool(testUserId, 100)
	if err != datapool.ErrDataPoolNotExist {
		t.Fatalf("want has error 'ErrDataPoolNotExist', but has error %#v", err)
	}

	err = dp.CreateDataPool(testUserId, 0)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	err = dp.SetDataPool(testUserId, 100)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	pool, err := dp.GetDataPool(testUserId)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	if pool != 100 {
		t.Fatalf("want datapool is 100, but %#v", pool)
	}

}

func Test_ExistDataPool(t *testing.T) {
	dp := datapool.NewService()
	if dp == nil {
		t.Fatalf("want datapool is not nil, but nil")
	}

	exist := dp.ExistDataPool(testUserId)
	if exist == true {
		t.Fatalf("want datapool does not exist, but exist")
	}

	err := dp.CreateDataPool(testUserId, 0)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}

	exist = dp.ExistDataPool(testUserId)
	if exist != true {
		t.Fatalf("want datapool exist, but doesn't exist")
	}

}
