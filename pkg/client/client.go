package client

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sc-chat/test-chat/internal/constants"
	"github.com/sc-chat/test-chat/internal/debug"
	"github.com/sc-chat/test-chat/pkg/chat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const ms = 500

// Client struct
type Client struct {
	Addr    string
	Name    string
	Timeout time.Duration
	Logger  debug.Logger

	chatClient chat.ChatClient
	token      string
	shutdown   bool
}

// Run method
func (c *Client) Run(ctx context.Context) error {
	connCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(connCtx, c.Addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.WithMessage(err, "failed to connect to provided address")
	}
	defer conn.Close()

	c.Logger.Debug("%s is connected to %s", c.Name, c.Addr)

	c.chatClient = chat.NewChatClient(conn)

	if c.token, err = c.login(ctx); err != nil {
		return errors.WithMessage(err, "failed to login")
	}

	c.Logger.Debug("Logged in successfully as %s", c.Name)

	err = c.stream(ctx)

	c.Logger.Debug("Logging out")
	if err := c.logout(ctx); err != nil {
		c.Logger.Debug("Failed to log out: %v", err)
	}

	return errors.WithMessage(err, "Stream error")
}

// Login method
func (c *Client) login(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	res, err := c.chatClient.Login(ctx, &chat.LoginRequest{
		Name: c.Name,
	})

	if err != nil {
		return "", err
	}

	return res.Token, nil
}

// Logout method
func (c *Client) logout(ctx context.Context) error {
	if c.shutdown {
		// DebugLogf("unable to logout (server sent shutdown signal)")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := c.chatClient.Logout(ctx, &chat.LogoutRequest{Token: c.token})
	if s, ok := status.FromError(err); ok && s.Code() == codes.Unavailable {
		// DebugLogf("unable to logout (connection already closed)")
		return nil
	}

	return err
}

func (c *Client) stream(ctx context.Context) error {
	// attach token for outgoing stream
	md := metadata.New(map[string]string{constants.TokenHeader: c.token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := c.chatClient.Stream(ctx)
	if err != nil {
		return err
	}
	defer client.CloseSend()

	c.Logger.Debug("Connected to stream")

	// run send/receive methods
	go c.send(client)
	return c.receive(client)
}

func (c *Client) receive(sc chat.Chat_StreamClient) error {
	for {
		res, err := sc.Recv()

		if s, ok := status.FromError(err); ok && s.Code() == codes.Canceled {
			c.Logger.Debug("Stream canceled (usually indicates shutdown)")
			return nil
		} else if err == io.EOF {
			c.Logger.Debug("Stream closed by server")
			return nil
		} else if err != nil {
			return err
		}

		// handle event
		switch evt := res.Event.(type) {
		case *chat.ResponseStream_ClientLogin:
			log.Printf("Server: %s is online", evt.ClientLogin.Name)
		case *chat.ResponseStream_ClientLogout:
			log.Printf("Server: %s is offline", evt.ClientLogout.Name)
		case *chat.ResponseStream_ClientMessage:
			log.Printf("%s: %s", evt.ClientMessage.Name, evt.ClientMessage.Message)
		case *chat.ResponseStream_ServerShutdown:
			c.Logger.Debug("The server is shutting down %#v", evt)
			c.shutdown = true

			// kill the client
			syscall.Kill(os.Getpid(), syscall.SIGTERM)
		default:
			c.Logger.Debug("Unexpected event from the server: %T", evt)
			return nil
		}
	}
}

func (c *Client) send(client chat.Chat_StreamClient) {
	sc := bufio.NewScanner(os.Stdin)
	sc.Split(bufio.ScanLines)

	for {
		select {
		case <-client.Context().Done():
			c.Logger.Debug("Client send loop disconnected")
		default:
			if sc.Scan() {
				if err := client.Send(&chat.RequestStream{Message: sc.Text()}); err != nil {
					c.Logger.Debug("Failed to send message: %v", err)
					return
				}
			} else {
				c.Logger.Debug("Input scanner failure: %v", sc.Err())
				return
			}
		}
	}
}

// NewClient returns Client pointer
func NewClient(addr, name string, allowDebug bool) (*Client, error) {
	// basic server address validation
	if addr == "" {
		return nil, errors.New("Invalid address")
	}

	// basic name validation
	if name == "" {
		return nil, errors.New("Invalid name")
	}

	return &Client{
		Addr:    addr,
		Name:    name,
		Timeout: time.Duration(ms) * time.Millisecond,
		Logger:  debug.NewLogger(allowDebug),
	}, nil
}
