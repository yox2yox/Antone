package stolerance

import (
	"os"
	"testing"
	"yox2yox/antone/internal/log2"
)

var (
	groups1 = [][]float64{
		[]float64{
			0.3,
			0.6,
			0.2,
			0.1,
		},
		[]float64{
			0.9,
			0.9,
			0.4,
			0.1,
		},
		[]float64{
			0.2,
			0.8,
			0.5,
			0.1,
			0.4,
		},
	}
)

func TestMain(m *testing.M) {
	// パッケージ内のテストの実行
	code := m.Run()
	// 終了処理
	log2.Close()
	// テストの終了コードで exit
	os.Exit(code)
}

func TestWorkerCred(t *testing.T) {
	f := 0.3
	k := 1
	cred := CalcWorkerCred(f, k)
	if cred < 0 || cred > 1 {
		t.Fatalf("want cred=%f, but %f", 0.842337, cred)
	}
}

func TestSigma(t *testing.T) {
	datalist := []float64{
		0.1,
		0.4,
		0.2,
		0.3,
	}
	sum := sigma(datalist, false, []int{})
	if sum != 1 {
		t.Fatalf("want sum=%f, but %f", 1.0, sum)
	}
	sum = sigma(datalist, false, []int{1, 3})
	if sum >= 0.4 || sum < 0.3 {
		t.Fatalf("want sum=%f, but %.20f", 0.3, sum)
	}
	sum = sigma(datalist, true, []int{})
	if sum != 3 {
		t.Fatalf("want sum=%f, but %f", 0.0, sum)
	}

}

func TestLambda(t *testing.T) {
	datalist := []float64{
		0.1,
		0.4,
		0.2,
		0.3,
	}
	lam := lambda(datalist, false, []int{})
	if lam < 0.0024 {
		t.Fatalf("want sum=%f, but %.20f", 1.0, lam)
	}
	lam = lambda(datalist, false, []int{1, 3})
	if lam < 0.02 || lam >= 0.03 {
		t.Fatalf("want sum=%f, but %.20f", 0.3, lam)
	}
	lam = lambda(datalist, true, []int{})
	if lam < 0.3 || lam >= 0.4 {
		t.Fatalf("want sum=%f, but %f", 0.0, lam)
	}

}

func TestGroupCred(t *testing.T) {
	groups := groups1
	cred := CalcRGroupCred(0, groups)
	if cred < 0.002517 || cred >= 0.002519 {
		t.Fatalf("want cred=%f, but %f", 0.002518, cred)
	}
}

func TestNeedWorkerCount(t *testing.T) {
	groups := groups1
	avgcred := 0.7
	threshold := 0.9999999999999999

	count, group := CalcNeedWorkerCountAndBestGroup(avgcred, groups, threshold)

	if count != 42 {
		t.Fatalf("want count=%d, but %d", 42, count)
	}

	if group != 1 {
		t.Fatalf("want best group=%d, but %d", 1, group)
	}

}
