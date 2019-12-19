package worker

import (
	"context"
	"errors"
	"math/rand"
	"time"
	"yox2yox/antone/internal/log2"
	"yox2yox/antone/pkg/stolerance"
	"yox2yox/antone/worker/datapool"
	pb "yox2yox/antone/worker/pb"
)

type Endpoint struct {
	Datapool   *datapool.Service
	Db         map[string]int32
	Id         string
	AttackMode int
	BadMode    bool
	Reputation int
}

func NewEndpoint(datapool *datapool.Service, badmode bool, attackMode int) *Endpoint {
	return &Endpoint{
		Datapool: datapool,
		Db: map[string]int32{
			"client0": 0,
		},
		BadMode:    badmode,
		AttackMode: attackMode,
		Reputation: 0,
	}
}

var (
	ErrDataPoolAlreadyExist = errors.New("this user's datapool already exists")
	ErrDataPoolNotExist     = errors.New("this user's datapool does not exist")
)

func (e *Endpoint) GetValidatableCode(ctx context.Context, vCodeRequest *pb.ValidatableCodeRequest) (*pb.ValidatableCode, error) {
	datapoolid := vCodeRequest.Datapoolid
	data, ver, err := e.Datapool.GetDataPool(datapoolid)
	if err == datapool.ErrDataPoolNotExist {
		return nil, ErrDataPoolNotExist
	} else if err != nil {
		return nil, err
	}
	return &pb.ValidatableCode{Data: data, Add: vCodeRequest.Add, Version: ver}, nil
}

func (e *Endpoint) OrderValidation(ctx context.Context, validatableCode *pb.ValidatableCode) (*pb.ValidationResult, error) {
	log2.Debug.Printf("got bad reputations %#v", validatableCode.Badreputations)
	log2.Debug.Printf("start working for validation Attack[%d] FirstNodeFault[%#v]", e.AttackMode, validatableCode.FirstNodeIsfault)

	if e.AttackMode == 1 {
		// Attack 1
		creds := []float64{}
		for _, rep := range validatableCode.Badreputations {
			wcred := stolerance.CalcWorkerCred(validatableCode.FaultyFraction, int(rep))
			creds = append(creds, wcred)
		}
		results := [][]float64{creds}
		log2.Debug.Printf("got bad results %#v", results)
		gcred := stolerance.CalcRGroupCred(0, results)
		log2.Debug.Printf("group cred for bad workers is %f", gcred)
		if e.BadMode && gcred >= float64(validatableCode.Threshould) {
			return &pb.ValidationResult{Pool: -1, Reject: false}, nil
		}
	}

	if e.AttackMode == 2 {
		// Attack 2 : 不正ワーカ数のみを利用
		if e.BadMode && len(validatableCode.Badreputations) >= 2 {
			return &pb.ValidationResult{Pool: -1, Reject: false}, nil
		}
	}

	if e.AttackMode == 3 {
		// Attack 3 : Credibilityおよび不正ワーカ数を利用
		if e.BadMode && len(validatableCode.Badreputations) >= 2 {
			badThreshold := true
			for _, rep := range validatableCode.Badreputations {
				if stolerance.CalcWorkerCred(validatableCode.FaultyFraction, int(rep)) < float64(validatableCode.Threshould) {
					badThreshold = false
				}
			}
			if badThreshold {
				return &pb.ValidationResult{Pool: -1, Reject: false}, nil
			}
		}
	}

	if e.AttackMode == 4 {
		// Attack 4 : Credibilityおよび不正ワーカ数を利用For Step Voting
		if e.BadMode && validatableCode.CountUnstabotagable <= 0 {
			stabotageRate := 1.0
			if len(validatableCode.Badreputations) <= 1 {
				cred := stolerance.CalcWorkerCred(validatableCode.FaultyFraction, int(validatableCode.Badreputations[0]))
				log2.Debug.Printf("first bad node start calc satabotage %d %f", validatableCode.Badreputations[0], cred)
				rand.Seed(time.Now().UnixNano())
				if rand.Float64() <= stabotageRate {
					log2.Debug.Printf("first bad node try to sabotage")
					return &pb.ValidationResult{Pool: -1, Reject: false}, nil
				}
			} else if validatableCode.FirstNodeIsfault {
				log2.Debug.Printf("first node was fault")
				isFist := true
				groups := [][]float64{[]float64{}}
				for _, rep := range validatableCode.Badreputations {
					if isFist {
						groups[0] = append(groups[0], stolerance.CalcWorkerCred(validatableCode.FaultyFraction, int(rep)))
						isFist = false
					} else {
						groups[0] = append(groups[0], stolerance.CalcSecondaryWorkerCred(validatableCode.FaultyFraction, int(rep)))
					}
				}
				gred := stolerance.CalcStepVotingGruopCred(0, groups)
				log2.Debug.Printf("bad group is %#v", groups)
				log2.Debug.Printf("bad group cred is %f", gred)
				if gred >= float64(validatableCode.Threshould) {
					log2.Debug.Printf("secondary worker sabotaged")
					return &pb.ValidationResult{Pool: -1, Reject: false}, nil
				}
			}
		}
	}

	pool := validatableCode.Data + validatableCode.Add
	e.Reputation++
	return &pb.ValidationResult{Pool: pool, Reject: false}, nil
}
