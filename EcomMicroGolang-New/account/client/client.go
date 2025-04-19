package client

import (
	"context"

	"github.com/thomas/EcommerceAPI/account/internal"
	"github.com/thomas/EcommerceAPI/account/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.AccountServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	C := pb.NewAccountServiceClient(conn)
	return &Client{conn, C}, nil
}

func (client *Client) Close() {
	client.conn.Close()
}

func (client *Client) Register(ctx context.Context, name, email, password string) (string, error) {
	response, err := client.service.Register(ctx, &pb.RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return response.Token, nil
}

func (client *Client) Login(ctx context.Context, email, password string) (string, error) {
	response, err := client.service.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return response.Token, nil
}

func (client *Client) GetAccount(ctx context.Context, Id string, userID string) (*internal.Account, error) {
	// Add the user ID to the context metadata
	ctx = metadata.AppendToOutgoingContext(ctx, "caller-id", userID)

	r, err := client.service.GetAccount(
		ctx,
		&pb.GetAccountRequest{Id: Id},
	)
	if err != nil {
		return nil, err
	}
	return &internal.Account{
		ID:    uint(r.Account.GetId()),
		Name:  r.Account.GetName(),
		Email: r.Account.GetEmail(),
	}, nil
}

func (client *Client) GetAccounts(ctx context.Context, skip, take uint64, userID string) ([]internal.Account, error) {
	// Add the user ID to the context metadata
	ctx = metadata.AppendToOutgoingContext(ctx, "caller-id", userID)

	r, err := client.service.GetAccounts(
		ctx,
		&pb.GetAccountsRequest{Take: take, Skip: skip},
	)
	if err != nil {
		return nil, err
	}
	var accounts []internal.Account
	for _, a := range r.Accounts {
		accounts = append(accounts, internal.Account{
			ID:    uint(a.GetId()),
			Name:  a.GetName(),
			Email: a.GetEmail(),
		})
	}
	return accounts, nil
}
