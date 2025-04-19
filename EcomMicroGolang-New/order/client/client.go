package client

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/thomas/EcommerceAPI/order/models"
	"github.com/thomas/EcommerceAPI/order/proto/pb"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := pb.NewOrderServiceClient(conn)
	return &Client{conn, c}, nil
}

func (client *Client) Close() {
	client.conn.Close()
}

func (client *Client) PostOrder(
	ctx context.Context,
	accountID string,
	products []*models.OrderedProduct,
) (*models.Order, error) {
	var protoProducts []*pb.OrderProduct
	for _, p := range products {
		protoProducts = append(protoProducts, &pb.OrderProduct{
			Id:       p.ID,
			Quantity: p.Quantity,
		})
	}

	r, err := client.service.PostOrder(
		ctx,
		&pb.PostOrderRequest{
			AccountId: accountID,
			Products:  protoProducts,
		},
	)
	if err != nil {
		return nil, err
	}
	// Create response order
	newOrder := r.Order
	newOrderCreatedAt := time.Time{}
	newOrderCreatedAt.UnmarshalBinary(newOrder.CreatedAt)

	// Extract product details from the response
	var orderProducts []*models.OrderedProduct
	for _, p := range newOrder.Products {
		orderProducts = append(orderProducts, &models.OrderedProduct{
			ID:          p.Id,
			Quantity:    p.Quantity,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		})
	}

	return &models.Order{
		ID:         uint(r.Order.GetId()),
		CreatedAt:  newOrderCreatedAt,
		TotalPrice: newOrder.TotalPrice,
		AccountID:  newOrder.AccountId,
		Products:   orderProducts,
	}, nil
}

func (client *Client) GetOrdersForAccount(ctx context.Context, accountID string, userID string) ([]models.Order, error) {
	// Add the user ID to the context metadata
	ctx = metadata.AppendToOutgoingContext(ctx, "caller-id", userID)

	r, err := client.service.GetOrdersForAccount(ctx, &pb.GetOrdersForAccountRequest{
		AccountId: accountID,
	})
	if err != nil {
		log.Println("Error getting orders for account:", err)
		return nil, err
	}

	// Create response orders
	var orders []models.Order
	for _, orderProto := range r.Orders {
		newOrder := models.Order{
			ID:         uint(orderProto.Id),
			TotalPrice: orderProto.TotalPrice,
			AccountID:  orderProto.AccountId,
		}
		newOrder.CreatedAt = time.Time{}
		newOrder.CreatedAt.UnmarshalBinary(orderProto.CreatedAt)

		var products []*models.OrderedProduct
		for _, p := range orderProto.Products {
			products = append(products, &models.OrderedProduct{
				ID:          p.Id,
				Quantity:    p.Quantity,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
		newOrder.Products = products

		orders = append(orders, newOrder)
	}
	return orders, nil
}
