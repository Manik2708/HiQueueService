package server

import (
	"context"
	"errors"
	"net"

	cn "github.com/Manik2708/HiQueueService/pkg/consumer"
	pb "github.com/Manik2708/HiQueueService/pkg/grpc"
	"google.golang.org/grpc"
)


type CoreServer struct{
	port string
	cns *cn.HiConsumer
	pb.UnimplementedQueueServiceServer
}

func (c *CoreServer) RecieveMessages(rq *pb.RecieveMessageRequest, s grpc.ServerStreamingServer[pb.RecieveMessageResponse]) error{
	chnl := make(chan cn.ChannelResponse[string])
	irq := cn.NewRecievingRequest(rq.Id, chnl)
	c.cns.RecChan <- *irq
	for rsp := range chnl{
		if rsp.Err != nil{
			return rsp.Err
		}
		strRsp := &pb.RecieveMessageResponse{}
		strRsp.Content = *rsp.Msg
		s.Send(strRsp)
	}
	return nil
}

func (c *CoreServer) SendMessages(ctx context.Context, rq *pb.SendMessageRequest) (*pb.SendMesssageResponse, error){
	chnl := make(chan cn.ChannelResponse[bool])
	irq := cn.NewSendingRequest(rq.Id, rq.Content, chnl)
	c.cns.SendChan <- *irq
	for rsp := range chnl{
		if rsp.Err != nil{
			return nil,rsp.Err
		}else{
			response := &pb.SendMesssageResponse{}
			response.Success = *rsp.Msg
			return response, nil
		}
	}
	return nil, errors.New("operation unsuccesful due to unknown error")
}

func (c *CoreServer) New() error {
	lis, err := net.Listen("tcp", ":"+c.port)
	if err != nil{
		return err
	}
	s := grpc.NewServer()
	pb.RegisterQueueServiceServer(s, &CoreServer{})
	err = s.Serve(lis)
	if err != nil{
		return err
	}
	return nil
}

