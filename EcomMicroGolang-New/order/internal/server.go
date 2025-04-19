package internal

import (
	"context"
	"fmt"
	"log"
	"net"

	mapset "github.com/deckarep/golang-set/v2"
	account "github.com/thomas/EcommerceAPI/account/client"
	"github.com/thomas/EcommerceAPI/order/models"
	"github.com/thomas/EcommerceAPI/order/proto/pb"
	product "github.com/thomas/EcommerceAPI/product/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	pb.UnimplementedOrderServiceServer
	service       Service
	accountClient *account.Client
	productClient *product.Client
}

func ListenGRPC(service Service, accountURL string, productURL string, port int) error {
	accountClient, err := account.NewClient(accountURL)
	if err != nil {
		return err
	}

	productClient, err := product.NewClient(productURL)
	if err != nil {
		accountClient.Close()
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		accountClient.Close()
		productClient.Close()
		return err
	}

	serv := grpc.NewServer()
	pb.RegisterOrderServiceServer(serv, &grpcServer{
		pb.UnimplementedOrderServiceServer{},
		service,
		accountClient,
		productClient,
	})
	reflection.Register(serv)

	return serv.Serve(lis)
}

func (server *grpcServer) PostOrder(ctx context.Context, request *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	// Pass the account ID as both the ID to fetch and the caller ID for authentication
	_, err := server.accountClient.GetAccount(ctx, request.AccountId, request.AccountId)
	if err != nil {
		log.Println("Error getting account", err)
		return nil, err
	}

	// Create a map to aggregate quantities for duplicate product IDs
	productQuantities := make(map[string]uint32)
	var uniqueProductIDs []string

	// Aggregate quantities and collect unique product IDs
	for _, p := range request.Products {
		productQuantities[p.Id] += p.Quantity

		// Only add to uniqueProductIDs if we haven't seen this ID before
		found := false
		for _, id := range uniqueProductIDs {
			if id == p.Id {
				found = true
				break
			}
		}
		if !found {
			uniqueProductIDs = append(uniqueProductIDs, p.Id)
		}
	}

	// Get product details from the product service
	orderedProducts, err := server.productClient.GetProducts(ctx, 0, 0, uniqueProductIDs, "")
	if err != nil {
		log.Println("Error getting ordered products", err)
		return nil, err
	}

	var products []*models.OrderedProduct
	totalPrice := 0.0

	// Create ordered products with aggregated quantities
	for _, p := range orderedProducts {
		quantity, exists := productQuantities[p.ID]
		if exists && quantity > 0 {
			productObj := &models.OrderedProduct{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
				Quantity:    quantity,
			}

			products = append(products, productObj)
			totalPrice += productObj.Price * float64(productObj.Quantity)
		}
	}

	postOrder, err := server.service.PostOrder(ctx, request.AccountId, totalPrice, products)
	if err != nil {
		log.Println("Error posting postOrder", err)
		return nil, err
	}

	orderProto := &pb.Order{
		Id:         uint64(postOrder.ID),
		AccountId:  postOrder.AccountID,
		TotalPrice: postOrder.TotalPrice,
		Products:   []*pb.ProductInfo{},
	}
	orderProto.CreatedAt, _ = postOrder.CreatedAt.MarshalBinary()
	for _, p := range postOrder.Products {
		orderProto.Products = append(orderProto.Products, &pb.ProductInfo{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
		})
	}

	return &pb.PostOrderResponse{
		Order: orderProto,
	}, nil
}

func (server *grpcServer) GetOrdersForAccount(ctx context.Context, request *pb.GetOrdersForAccountRequest) (*pb.GetOrdersForAccountResponse, error) {
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

	// Only allow users to access their own orders
	if callerID != request.AccountId {
		return nil, status.Errorf(codes.PermissionDenied, "cannot access another user's orders")
	}

	accountOrders, err := server.service.GetOrdersForAccount(ctx, request.AccountId)
	if err != nil {
		log.Println("Error getting orders for account:", err)
		return nil, err
	}

	// Taking unique products. We use set to avoid repeating
	productIDsSet := mapset.NewSet[string]()
	for _, o := range accountOrders {
		for _, p := range o.Products {
			productIDsSet.Add(p.ID)
		}
	}

	productIDs := productIDsSet.ToSlice()

	products, err := server.productClient.GetProducts(ctx, 0, 0, productIDs, "")
	if err != nil {
		log.Println("Error getting account products: ", err)
		return nil, err
	}

	// Collecting orders

	var orders []*pb.Order
	for _, order := range accountOrders {
		// Encode order
		encodedOrder := &pb.Order{
			AccountId:  order.AccountID,
			Id:         uint64(order.ID),
			TotalPrice: order.TotalPrice,
			Products:   []*pb.ProductInfo{},
		}
		encodedOrder.CreatedAt, _ = order.CreatedAt.MarshalBinary()

		// Decorate orders with products
		for _, orderedProduct := range order.Products {
			// Populate product fields
			for _, prod := range products {
				if prod.ID == orderedProduct.ID {
					orderedProduct.Name = prod.Name
					orderedProduct.Description = prod.Description
					orderedProduct.Price = prod.Price
					break
				}
			}

			encodedOrder.Products = append(encodedOrder.Products, &pb.ProductInfo{
				Id:          orderedProduct.ID,
				Name:        orderedProduct.Name,
				Description: orderedProduct.Description,
				Price:       orderedProduct.Price,
				Quantity:    orderedProduct.Quantity,
			})
		}

		orders = append(orders, encodedOrder)
	}
	return &pb.GetOrdersForAccountResponse{Orders: orders}, nil
}
