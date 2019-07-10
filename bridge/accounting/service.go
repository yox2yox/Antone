package accounting

import (
	"errors"
	"math/rand"
	"time"
)

type Client struct {
	Id      string
	Balance int
}

type Worker struct {
	Addr       string
	Id         string
	Reputation int
}

type Service struct {
	Workers                     map[string]*Worker
	Clients                     map[string]*Client
	WorkersId                   []string
	Holders                     map[string][]string
	WithoutCommunicationForTest bool
}

var (
	ErrIDNotExist          = errors.New("this id doesn't exist")
	ErrIDAlreadyExists     = errors.New("this id already exsits")
	ErrWorkersAreNotEnough = errors.New("There are not enough workers")
)

func NewService(withoutCommunicationForTest bool) *Service {
	return &Service{
		Workers:                     map[string]*Worker{},
		Clients:                     map[string]*Client{},
		WorkersId:                   []string{},
		Holders:                     map[string][]string{},
		WithoutCommunicationForTest: withoutCommunicationForTest,
	}
}

func (s *Service) GetWorkersCount() int {
	return len(s.Workers)
}

//workerの中からnum台選択して返す
func (s *Service) SelectValidationWorkers(num int) ([]*Worker, error) {
	if len(s.Workers) < num {
		return nil, ErrWorkersAreNotEnough
	}

	rand.Seed(time.Now().UnixNano())
	picked := []*Worker{}
	for i := 0; i < num; i++ {
		pickedId := s.WorkersId[rand.Intn(len(s.Workers))]
		picked = append(picked, s.Workers[pickedId])
	}

	return picked, nil
}

//userIdのデータを保有しているHolderの中から一台選択して返す
func (s *Service) SelectDBHolder(userId string) *Worker {
	rand.Seed(time.Now().UnixNano())
	holderid := s.Holders[userId][rand.Intn(len(s.Holders))]
	return s.Workers[holderid]
}

//新規holderを登録する
func (s *Service) RegistarDBHolder(userId string, workerid string) error {
	rand.Seed(time.Now().UnixNano())
	_, exist := s.Workers[workerid]
	if exist == false {
		return ErrIDNotExist
	}
	for _, holder := range s.Holders[userId] {
		if holder == workerid {
			return ErrIDAlreadyExists
		}
	}

	//Remoteワーカに新規DB登録リクエスト送信
	//TODO:実装
	/*if !s.WithoutCommunicationForTest {
		conn, err := grpc.Dial(remote.Addr, grpc.WithInsecure())
		workerClient := workerpb.NewWorkerClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		vCodeWorker := &workerpb.ValidatableCode{Data: vCode.Data, Add: vCode.Add}
		validationResult, err := workerClient.OrderValidation(ctx, vCodeWorker)
	}*/

	s.Holders[userId] = append(s.Holders[userId], workerid)
	return nil
}

//Workerアカウントを新規作成
func (s *Service) CreateNewWorker(workerId string, Addr string) (*Worker, error) {
	_, exist := s.Workers[workerId]
	if exist {
		return nil, ErrIDAlreadyExists
	}
	s.Workers[workerId] = &Worker{
		Addr:       Addr,
		Id:         workerId,
		Reputation: 0,
	}
	s.WorkersId = append(s.WorkersId, workerId)
	return s.Workers[workerId], nil
}

//Clientの新規アカウントを作成
func (s *Service) CreateNewClient(userId string) (*Client, error) {
	_, exist := s.Clients[userId]
	if exist {
		return nil, ErrIDAlreadyExists
	}

	s.Clients[userId] = &Client{
		Id:      userId,
		Balance: 0,
	}
	return s.Clients[userId], nil
}
