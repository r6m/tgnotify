package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"google.golang.org/grpc/codes"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/rezam90/tgnotify/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type server struct {
	client *http.Client
}

var (
	addr    = flag.String("addr", "localhost", "--addr localhost grpc server address")
	port    = flag.String("port", "50051", "--port 50051 grpc server port")
	baseUrl = "https://api.telegram.org/bot%s/"
)

func newServer() *server {
	return &server{
		client: http.DefaultClient,
	}
}

func (s *server) SendMessage(ctx context.Context, in *proto.SendMessageRequest) (*empty.Empty, error) {
	payload, err := json.Marshal(map[string]string{
		"chat_id": in.ChatId,
		"text":    in.Text,
	})
	if err != nil {
		logrus.Error("can't marshal ", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", in.BotToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		logrus.Error("can't init request ", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	req.Header.Add("Content-Type", "application/json")

	_, err = s.client.Do(req)
	if err != nil {
		logrus.Error("can't perform request", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", *addr, *port))
	if err != nil {
		log.Fatalln(err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterTgNotifyServer(grpcServer, newServer())

	go func() {
		log.Println("starting server on", fmt.Sprintf("%s:%s", *addr, *port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Println(err)
		}
	}()

	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("stopping server")
	grpcServer.GracefulStop()
}
