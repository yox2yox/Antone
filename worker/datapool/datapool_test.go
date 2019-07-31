package datapool_test

import (
	"testing"
	"yox2yox/antone/worker/datapool"
)

var (
	testUserId = "client0"
)

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
	err := dp.CreateNewDataPool(testUserId, 0)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	err = dp.CreateNewDataPool(testUserId, 0)
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

	err = dp.CreateNewDataPool(testUserId, 0)
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

	err = dp.CreateNewDataPool(testUserId, 0)
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

	err := dp.CreateNewDataPool(testUserId, 0)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}

	exist = dp.ExistDataPool(testUserId)
	if exist != true {
		t.Fatalf("want datapool exist, but doesn't exist")
	}

}
