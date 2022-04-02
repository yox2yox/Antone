package datapool

import (
	"context"
	"yox2yox/antone/bridge/accounting"
	"yox2yox/antone/bridge/pb"
	"yox2yox/antone/internal/log2"
)

type Endpoint struct {
	datapoolSv   *Service
	accountingSv *accounting.Service
}

func NewEndpoint(datapoolSv *Service, accountingSv *accounting.Service) *Endpoint {
	endpoint := &Endpoint{
		datapoolSv:   datapoolSv,
		accountingSv: accountingSv,
	}
	return endpoint
}

func (e *Endpoint) CreateDatapool(ctx context.Context, req *pb.CreateRequest) (*pb.DatapoolInfo, error) {
	log2.Debug.Printf("get access to CreateDatapool from USER[%s]", req.Userid)
	datapool, err := e.datapoolSv.CreateDatapool(req.Userid, int(req.Holdersnum))
	if err != nil {
		log2.Err.Printf("failed to create datapool %s", err)
		return nil, err
	}
	datapoolinfo := &pb.DatapoolInfo{
		Datapoolid: datapool.DatapoolId,
		Ownerid:    datapool.OwnrerId,
	}
	return datapoolinfo, nil
}

func (e *Endpoint) AddHolders(ctx context.Context, req *pb.AddHolderRequest) (*pb.AddHolderResult, error) {
	log2.Debug.Printf("get access to Addholders")
	data, _, err := e.datapoolSv.FetcheDatapoolFromRemote(req.Datapoolid)
	if err != nil {
		return nil, err
	}
	_, err = e.datapoolSv.AddHolders(req.Datapoolid, data, int(req.Holdersnum))
	if err != nil {
		return nil, err
	}
	return &pb.AddHolderResult{}, nil
}
