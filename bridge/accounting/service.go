package accounting

import (
	"math/rand"
	"time"
)

type Worker struct {
	Addr       string
	Id         string
	Reputation int
}

type Service struct {
	Workers   map[string]Worker
	WorkersId []string
	Holders   map[string][]string
}

func NewService() *Service {
	return &Service{
		Workers: map[string]Worker{
			"worker0": Worker{
				Addr:       "localhost",
				Id:         "woker0",
				Reputation: 0,
			},
		},
		WorkersId: []string{
			"worker0",
		},
		Holders: map[string][]string{
			"client0": []string{
				"worker0",
			},
		},
	}
}

func (s *Service) GetWorkersCount() int {
	return len(s.Workers)
}

func (s *Service) GetValidationWorkers(num int) []Worker {
	rand.Seed(time.Now().UnixNano())

	picked := []Worker{}
	for i := 0; i < num; i++ {
		pickedId := s.WorkersId[rand.Intn(len(s.Workers))]
		picked = append(picked, s.Workers[pickedId])
	}

	return picked
}

func (s *Service) GetDBHolder(userId string) Worker {
	rand.Seed(time.Now().UnixNano())
	holderid := s.Holders[userId][rand.Intn(len(s.Holders))]
	return s.Workers[holderid]
}
