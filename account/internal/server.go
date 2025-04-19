package internal

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/thomas/EcommerceAPI/account/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	pb.UnimplementedAccountServiceServer
	service Service
}

func ListenGRPC(service Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	serv := grpc.NewServer()

	pb.RegisterAccountServiceServer(serv, &grpcServer{
		UnimplementedAccountServiceServer: pb.UnimplementedAccountServiceServer{},
		service:                           service})
	reflection.Register(serv)
	return serv.Serve(lis)
}

func (server *grpcServer) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.AuthResponse, error) {
	token, err := server.service.Register(ctx, request.Name, request.Email, request.Password)
	if err != nil {
		return nil, err
	}
	return &pb.AuthResponse{
		Token: token,
	}, nil
}

func (server *grpcServer) Login(ctx context.Context, request *pb.LoginRequest) (*pb.AuthResponse, error) {
	token, err := server.service.Login(ctx, request.Email, request.Password)
	if err != nil {
		return nil, err
	}
	return &pb.AuthResponse{
		Token: token,
	}, nil
}

func (server *grpcServer) GetAccount(ctx context.Context, r *pb.GetAccountRequest) (*pb.AccountResponse, error) {
	// Check for authentication metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	// Get the caller ID from metadata
	callerIDs := md.Get("caller-id")
	if len(callerIDs) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing caller ID")
	}
	callerID := callerIDs[0]

	// Only allow users to access their own account
	if callerID != r.Id {
		return nil, status.Errorf(codes.PermissionDenied, "cannot access another user's account")
	}

	a, err := server.service.GetAccount(ctx, r.Id)
	if err != nil {
		return nil, err
	}
	return &pb.AccountResponse{Account: &pb.Account{
		Id:    uint64(a.ID),
		Name:  a.Name,
		Email: a.Email,
	}}, nil
}

func (server *grpcServer) GetAccounts(ctx context.Context, r *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	// Check for authentication metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	// Get the caller ID from metadata
	callerIDs := md.Get("caller-id")
	if len(callerIDs) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing caller ID")
	}
	callerID := callerIDs[0]

	// For admin operations, we might want to check if the caller has admin privileges
	// For now, we'll just allow the caller to get their own account

	// Get the accounts
	res, err := server.service.GetAccounts(ctx, r.Skip, r.Take)
	if err != nil {
		return nil, err
	}

	// Filter accounts to only return the caller's account
	// In a real system, you might have admin users who can see all accounts
	var accounts []*pb.Account
	for _, p := range res {
		// Convert account ID to string for comparison
		accountID := strconv.Itoa(int(p.ID))

		// Only include the account if it belongs to the caller
		if accountID == callerID {
			accounts = append(accounts, &pb.Account{
				Id:    uint64(int(p.ID)),
				Name:  p.Name,
				Email: p.Email,
			})
		}
	}

	return &pb.GetAccountsResponse{Accounts: accounts}, nil
}
