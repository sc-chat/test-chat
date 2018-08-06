package server

import (
	"context"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	"github.com/sc-chat/test-chat/internal/constants"
	"github.com/sc-chat/test-chat/internal/debug"
	"github.com/sc-chat/test-chat/internal/randint"
	"github.com/sc-chat/test-chat/internal/sha256"
	"github.com/sc-chat/test-chat/pkg/chat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const tokenHeader = "x-token"

// NewServer returns Server pointer
func NewServer(addr string, allowDebug bool) (*Server, error) {
	// basic server address validation
	if addr == "" {
		return nil, errors.New("Invalid address")
	}

	return &Server{
		Addr:      addr,
		Clients:   NewClientState(),
		Logger:    debug.NewLogger(allowDebug),
		Broadcast: make(chan chat.ResponseStream, 1000),
	}, nil
}

// Server struct
type Server struct {
	Addr      string
	Clients   ClientProcessor
	Logger    debug.Logger
	Broadcast chan chat.ResponseStream
}

// Run method
func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	srv := grpc.NewServer()
	chat.RegisterChatServer(srv, s)

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return errors.WithMessage(err, "Failed to start on provided address")
	}

	s.Logger.Debug("Server listening on %s", s.Addr)

	go s.broadcast(ctx)

	go func() {
		sErr := srv.Serve(l)
		if sErr != nil {
			log.Println("Serve error", sErr)
		}
		cancel()
	}()

	<-ctx.Done()

	s.Broadcast <- chat.ResponseStream{
		Timestamp: ptypes.TimestampNow(),
		Event: &chat.ResponseStream_ServerShutdown{
			ServerShutdown: &chat.ResponseStream_Shutdown{},
		},
	}

	s.Logger.Debug("Shutting down")

	srv.GracefulStop()
	close(s.Broadcast)
	return nil
}

// Login method
func (s *Server) Login(ctx context.Context, req *chat.LoginRequest) (*chat.LoginResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	// generate unique token
	token := sha256.NewHash(req.Name + "_" + strconv.FormatInt(time.Now().Unix(), 10) + "_" + randint.NewRandomIntString(8))

	// add client
	ok := s.Clients.Add(req.Name, token)

	s.Logger.Debug("%s (%s) has logged in", req.Name, token)

	if ok {
		s.Broadcast <- chat.ResponseStream{
			Timestamp: ptypes.TimestampNow(),
			Event: &chat.ResponseStream_ClientLogin{
				ClientLogin: &chat.ResponseStream_Login{
					Name: req.Name,
				},
			},
		}
	}

	return &chat.LoginResponse{Token: token}, nil
}

// Logout method
func (s *Server) Logout(ctx context.Context, req *chat.LogoutRequest) (*chat.LogoutResponse, error) {
	name, ok := s.Clients.Remove(req.Token)
	if !ok && name == "" {
		return nil, status.Error(codes.NotFound, "Token not found")
	}

	s.Logger.Debug("%s (%s) has logged out", name, req.Token)

	if ok {
		s.Broadcast <- chat.ResponseStream{
			Timestamp: ptypes.TimestampNow(),
			Event: &chat.ResponseStream_ClientLogout{
				ClientLogout: &chat.ResponseStream_Logout{
					Name: name,
				},
			},
		}
	}

	return new(chat.LogoutResponse), nil
}

// Stream method
func (s *Server) Stream(srv chat.Chat_StreamServer) error {
	token, ok := s.getToken(srv.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "Missing token header")
	}

	name, ok := s.Clients.GetNameByToken(token)
	if !ok {
		return status.Error(codes.Unauthenticated, "Invalid token")
	}

	go s.sendEventsToClient(srv, token)

	for {
		req, err := srv.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		s.Logger.Debug("%s (%s) has sent a message: %s", name, token, req.Message)

		s.Broadcast <- chat.ResponseStream{
			Timestamp: ptypes.TimestampNow(),
			Event: &chat.ResponseStream_ClientMessage{
				ClientMessage: &chat.ResponseStream_Message{
					Name:    name,
					Message: req.Message,
				},
			},
		}
	}

	<-srv.Context().Done()
	return srv.Context().Err()
}

func (s *Server) sendEventsToClient(srv chat.Chat_StreamServer, token string) {
	stream := s.Clients.AddStream(token)
	defer s.Clients.CloseStream(token)

	for {
		select {
		case <-srv.Context().Done():
			// client is closed
			return

		// read new event
		case res := <-stream:
			if r, ok := status.FromError(srv.Send(&res)); ok {
				switch r.Code() {
				case codes.OK:
					// nothing to do
				case codes.Unavailable, codes.Canceled, codes.DeadlineExceeded:
					s.Logger.Debug("Client (%s) terminated connection", token)
					return

				default:
					s.Logger.Debug("Failed to send to client (%s): %v", token, r.Err())
					return
				}
			}
		}
	}
}

// brodcast method spreads event to all connected clients
func (s *Server) broadcast(ctx context.Context) {
	for res := range s.Broadcast {
		s.Clients.Broadcast(res)
	}
}

// getToken method returns token from stream meta data
func (s *Server) getToken(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md[constants.TokenHeader]) == 0 {
		return "", false
	}

	return md[constants.TokenHeader][0], true
}
