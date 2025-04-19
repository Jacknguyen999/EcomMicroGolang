package graph

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/thomas/EcommerceAPI/order/models"
	"github.com/thomas/EcommerceAPI/pkg/auth"
)

var (
	ErrInvalidParameter = errors.New("invalid parameter")
)

type mutationResolver struct {
	server *Server
}

func (resolver *mutationResolver) Register(ctx context.Context, in RegisterInput) (*AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	token, err := resolver.server.accountClient.Register(ctx, in.Name, in.Email, in.Password)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ginContext, ok := ctx.Value("GinContextKey").(*gin.Context)
	if !ok {
		return nil, errors.New("could not retrieve gin context")
	}
	ginContext.SetCookie("token", token, 3600, "/", "localhost", false, true)

	return &AuthResponse{Token: token}, nil
}

func (resolver *mutationResolver) Login(ctx context.Context, in LoginInput) (*AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	token, err := resolver.server.accountClient.Login(ctx, in.Email, in.Password)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ginContext, ok := ctx.Value("GinContextKey").(*gin.Context)
	if !ok {
		return nil, errors.New("could not retrieve gin context")
	}
	ginContext.SetCookie("token", token, 3600, "/", "localhost", false, true)

	return &AuthResponse{Token: token}, nil
}

func (resolver *mutationResolver) CreateProduct(ctx context.Context, in CreateProductInput) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	log.Println("CreateProduct called with input:", in)

	// Enforce authentication - this will abort the request if not authenticated
	accountId, err := auth.GetUserIdInt(ctx, true)
	if err != nil {
		log.Println("Authentication failed for CreateProduct:", err)
		return nil, errors.New("unauthorized: you must be logged in to create a product")
	}

	log.Println("CreateProduct called with accountId:", accountId)
	postProduct, err := resolver.server.productClient.PostProduct(ctx, in.Name, in.Description, in.Price, int64(accountId))
	if err != nil {
		log.Println("Error creating product:", err)
		return nil, err
	}

	log.Println("Created product:", postProduct)
	log.Println("Product id: ", postProduct.ID)

	return &Product{
		ID:          postProduct.ID,
		Name:        postProduct.Name,
		Description: postProduct.Description,
		Price:       postProduct.Price,
		AccountID:   accountId,
	}, nil
}

func (resolver *mutationResolver) UpdateProduct(ctx context.Context, in UpdateProductInput) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Enforce authentication - this will abort the request if not authenticated
	accountId, err := auth.GetUserIdInt(ctx, true)
	if err != nil {
		log.Println("Authentication failed for UpdateProduct:", err)
		return nil, errors.New("unauthorized: you must be logged in to update a product")
	}

	updatedProduct, err := resolver.server.productClient.UpdateProduct(ctx, in.ID, in.Name, in.Description, in.Price, int64(accountId))
	if err != nil {
		log.Println("Error updating product:", err)
		return nil, err
	}

	return &Product{
		ID:          updatedProduct.ID,
		Name:        updatedProduct.Name,
		Description: updatedProduct.Description,
		Price:       updatedProduct.Price,
		AccountID:   accountId,
	}, nil
}

func (resolver *mutationResolver) DeleteProduct(ctx context.Context, id string) (*bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Enforce authentication - this will abort the request if not authenticated
	accountId, err := auth.GetUserIdInt(ctx, true)
	if err != nil {
		log.Println("Authentication failed for DeleteProduct:", err)
		return nil, errors.New("unauthorized: you must be logged in to delete a product")
	}

	err = resolver.server.productClient.DeleteProduct(ctx, id, int64(accountId))
	if err != nil {
		log.Println("Error deleting product:", err)
		return nil, err
	}

	success := true
	return &success, nil
}

func (resolver *mutationResolver) CreateOrder(ctx context.Context, in OrderInput) (*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Validate input
	if len(in.Products) == 0 {
		return nil, errors.New("order must contain at least one product")
	}

	var products []*models.OrderedProduct
	for _, product := range in.Products {
		if product.Quantity <= 0 {
			return nil, errors.New("product quantity must be greater than zero")
		}
		products = append(products, &models.OrderedProduct{
			ID:       product.ID,
			Quantity: uint32(product.Quantity),
		})
	}

	// Enforce authentication - this will abort the request if not authenticated
	accountId, err := auth.GetUserIdInt(ctx, true)
	if err != nil {
		log.Println("Authentication failed for CreateOrder:", err)
		return nil, errors.New("unauthorized: you must be logged in to create an order")
	}

	// Convert accountId to string for the order service
	accountIdStr := strconv.Itoa(accountId)

	// Create the order
	postOrder, err := resolver.server.orderClient.PostOrder(ctx, accountIdStr, products)
	if err != nil {
		log.Println("Error creating order:", err)
		return nil, err
	}

	// Format the response
	var orderedProducts []*OrderedProduct
	for _, orderedProduct := range postOrder.Products {
		orderedProducts = append(orderedProducts, &OrderedProduct{
			ID:          orderedProduct.ID,
			Name:        orderedProduct.Name,
			Description: orderedProduct.Description,
			Price:       orderedProduct.Price,
			Quantity:    int(orderedProduct.Quantity),
		})
	}

	return &Order{
		ID:         strconv.Itoa(int(postOrder.ID)),
		CreatedAt:  postOrder.CreatedAt,
		TotalPrice: postOrder.TotalPrice,
		Products:   orderedProducts,
	}, nil
}
