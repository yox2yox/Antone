package stolerance

import (
	"yox2yox/antone/internal/log2"
)

var e = 2.718281828459045

func CalcWorkerCred(f float64, reputation int) float64 {
	if reputation > 0 {
		return 1.0 - (f/(1-f))*(1/(float64(reputation)*e))
	} else {
		return 1.0 - f
	}
}

func CalcSecondaryWorkerCred(f float64, reputation int) float64 {
	if reputation > 0 {
		return 1.0 - (f/(1-f))*((float64(reputation)+1)/(float64(reputation)*e))
	} else {
		return (1.0 - f) * (1.0 - f)
	}
}

func CalcStepVotingGruopCred(targetidx int, groups [][]float64) float64 {
	if len(groups) <= 0 {
		return 0
	}
	return 1.0 - Lambda(groups[targetidx], true, []int{})
}

func CalcRGroupCred(targetidx int, groups [][]float64) float64 {
	var goodprob = []float64{}
	var badprob = []float64{}

	if len(groups) <= 0 {
		return 0
	}

	for _, group := range groups {
		good := Lambda(group, false, []int{})
		bad := Lambda(group, true, []int{})
		goodprob = append(goodprob, good)
		badprob = append(badprob, bad)
	}

	allptn := 0.0
	for index, prob := range goodprob {
		allptn += prob * Lambda(badprob, false, []int{index})
	}
	allptn += Lambda(badprob, false, []int{})

	cred := (goodprob[targetidx] * Lambda(badprob, false, []int{targetidx})) / allptn

	return cred
}

func CalcBestGroup(groups [][]float64, threshold float64) int {
	if threshold <= 0 || threshold >= 1 {
		log2.Err.Printf("threshold is not propbability.(%.30f)", threshold)
		return -1
	}
	maxgroup := 0
	maxcred := 0.0
	log2.Debug.Printf("start to calc groups' credibility%#v", groups)
	for index, _ := range groups {
		cred := CalcRGroupCred(index, groups)
		if cred > maxcred {
			maxcred = cred
			maxgroup = index
		}
	}

	log2.Debug.Printf("max credibility is %f", maxcred)
	return maxgroup
}

func CalcNeedWorkerCountAndBestGroup(avgcred float64, groups [][]float64, threshold float64) (int, int) {

	if threshold <= 0 || threshold >= 1 {
		log2.Err.Printf("threshold is not propbability.(%.30f)", threshold)
		return -1, -1
	}
	if avgcred <= 0 || avgcred >= 1 {
		log2.Err.Printf("average credibility is not propbability.(%.30f)", avgcred)
		return -1, -1
	}

	maxgroup := 0
	maxcred := 0.0
	log2.Debug.Printf("start to calc groups' credibility%#v", groups)
	for index, _ := range groups {
		cred := CalcRGroupCred(index, groups)
		if cred > maxcred {
			maxcred = cred
			maxgroup = index
		}
	}

	log2.Debug.Printf("max credibility is %f", maxcred)

	estimate := maxcred
	needcount := 0
	log2.Debug.Printf("start to calc need worker count")
	for estimate < threshold {
		needcount += 1
		groups[maxgroup] = append(groups[maxgroup], avgcred)
		estimate = CalcRGroupCred(maxgroup, groups)
	}

	log2.Debug.Printf("complete to calc need worker count [%d]", needcount)
	return needcount, maxgroup

}

func Sigma(target []float64, preverse bool, except []int) float64 {
	result := 0.0
	for index, x := range target {
		isExcepted := false
		for _, i := range except {
			if index == i {
				isExcepted = true
				break
			}
		}
		if !isExcepted {
			if preverse {
				x = 1.0 - x
			}
			result += x
		}
	}
	return result
}

func Lambda(target []float64, preverse bool, except []int) float64 {
	if len(target) == 0 {
		return 0
	}
	result := 1.0
	for index, x := range target {
		isExcepted := false
		for _, i := range except {
			if index == i {
				isExcepted = true
				break
			}
		}
		if !isExcepted {
			if preverse {
				x = 1 - x
			}
			result *= x
		}
	}
	return result
}
