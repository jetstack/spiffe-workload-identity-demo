package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/jetstack/spiffe-demo/internal/pkg/config"
	"github.com/jetstack/spiffe-demo/internal/pkg/server/proto"
)

type Server struct {
	proto.UnimplementedSpiffeDemoServer
}

func (s *Server) HelloWorld(ctx context.Context, empty *emptypb.Empty) (*proto.HelloWorldResponse, error) {
	resp := &proto.HelloWorldResponse{}

	clientSVID, hasSVID := grpccredentials.PeerIDFromContext(ctx)
	if !hasSVID {
		return resp, errors.New("no SVID provided")
	}

	log.Println("processing message for", clientSVID.String())

	resp.Message = fmt.Sprintf("Hello %s from the server", clientSVID.String())

	return resp, nil
}

func (s *Server) Start(ctx context.Context) {
	server := grpc.NewServer(grpc.Creds(grpccredentials.MTLSServerCredentials(config.CurrentSource, config.CurrentSource, tlsconfig.AuthorizeAny())))
	proto.RegisterSpiffeDemoServer(server, s)
	listener, err := net.Listen("tcp", "[::]:9090")
	if err != nil {
		panic("fail")
	}

	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}
