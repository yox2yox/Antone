package accounting

import (
	"math/rand"
	"time"
)

type Worker struct {
	addr       string
	id         string
	reputation int
}

type Service struct {
	workers   map[string]Worker
	workersId []string
	holders   map[string][]string
}

func NewService() *Service {
	return &Service{
		workers: map[string]Worker{
			"worker0": Worker{
				addr:       "localhost",
				id:         "woker0",
				reputation: 0,
			},
		},
		workersId: []string{
			"worker0",
		},
		holders: map[string][]string{
			"client0": []string{
				"worker0",
			},
		},
	}
}

func (s *Service) GetValidationWorkers(num int) []Worker {
	rand.Seed(time.Now().UnixNano())

	picked := []Worker{}
	for i := 0; i < num; i++ {
		pickedId := s.workersId[rand.Intn(len(s.workers))]
		picked = append(picked, s.workers[pickedId])
	}

	return picked
}

func (s *Service) GetDBHolder(userId string) Worker {
	rand.Seed(time.Now().UnixNano())
	holderid := s.holders[userId][rand.Intn(len(s.holders))]
	return s.workers[holderid]
}
