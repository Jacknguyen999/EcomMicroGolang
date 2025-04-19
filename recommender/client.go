package recommender

import (
	"context"
	"github.com/thomas/EcommerceAPI/recommender/generated/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.RecommenderServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := pb.NewRecommenderServiceClient(conn)
	return &Client{conn, client}, nil
}

func (client *Client) Close() {
	client.conn.Close()
}

func (client *Client) GetRecommendationForUser(ctx context.Context, userId string, skip, take uint64) (*pb.RecommendationResponse, error) {
	return client.service.GetRecommendations(
		ctx,
		&pb.RecommendationRequestForUserId{
			UserId: userId,
			Skip:   skip,
			Take:   take,
		},
	)
}

func (client *Client) GetRecommendationBasedOnViewed(ctx context.Context, ids []string, skip, take uint64) (*pb.RecommendationResponse, error) {
	return client.service.GetRecommendationsBasedOnViewed(
		ctx,
		&pb.RecommendationRequestOnViews{
			Ids:  ids,
			Skip: skip,
			Take: take,
		},
	)
}
