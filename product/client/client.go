package client

import (
	"context"
	"log"

	"google.golang.org/grpc"

	"github.com/thomas/EcommerceAPI/product/models"
	"github.com/thomas/EcommerceAPI/product/proto/pb"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.ProductServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := pb.NewProductServiceClient(conn)
	return &Client{conn, client}, nil
}

func (client *Client) Close() {
	client.conn.Close()
}

func (client *Client) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	res, err := client.service.GetProduct(ctx, &pb.ProductByIdRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	return &models.Product{
		ID:          res.Product.Id,
		Name:        res.Product.Name,
		Description: res.Product.Description,
		Price:       res.Product.Price,
		AccountID:   int(res.Product.GetAccountId()),
		Category:    "", // Category is not available in the current implementation
	}, nil
}

func (client *Client) GetProducts(ctx context.Context, skip, take uint64, ids []string, query string) ([]models.Product, error) {
	res, err := client.service.GetProducts(ctx, &pb.GetProductsRequest{
		Skip:  skip,
		Take:  take,
		Ids:   ids,
		Query: query,
	})
	if err != nil {
		return nil, err
	}
	var products []models.Product
	for _, p := range res.Products {
		// Get the product from Elasticsearch to get the category field
		product, err := client.GetProduct(ctx, p.Id)
		if err != nil {
			// If there's an error, just use the data from the gRPC response
			products = append(products, models.Product{
				ID:          p.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
				AccountID:   int(p.AccountId),
				Category:    "", // Category is not available in the current implementation
			})
		} else {
			// Use the data from Elasticsearch
			products = append(products, *product)
		}
	}
	return products, nil
}

func (client *Client) PostProduct(ctx context.Context, name, description string, price float64, accountId int64, category string) (*models.Product, error) {
	// Note: The current implementation doesn't support category
	res, err := client.service.PostProduct(ctx, &pb.CreateProductRequest{
		Name:        name,
		Description: description,
		Price:       price,
		AccountId:   accountId,
	})
	if err != nil {
		log.Println("Error creating product", err)
		return nil, err
	}
	return &models.Product{
		ID:          res.Product.Id,
		Name:        res.Product.Name,
		Description: res.Product.Description,
		Price:       res.Product.Price,
		AccountID:   int(res.Product.GetAccountId()),
		Category:    category,
	}, nil
}

func (client *Client) UpdateProduct(ctx context.Context, id, name, description string, price float64, accountId int64, category string) (*models.Product, error) {
	// Note: The current implementation doesn't support category
	res, err := client.service.UpdateProduct(ctx, &pb.UpdateProductRequest{
		Id:          id,
		Name:        name,
		Description: description,
		Price:       price,
		AccountId:   accountId,
	})
	if err != nil {
		return nil, err
	}
	return &models.Product{
		ID:          res.Product.Id,
		Name:        res.Product.Name,
		Description: res.Product.Description,
		Price:       res.Product.Price,
		AccountID:   int(res.Product.GetAccountId()),
		Category:    category,
	}, nil
}

func (client *Client) DeleteProduct(ctx context.Context, productId string, accountId int64) error {
	_, err := client.service.DeleteProduct(ctx, &pb.DeleteProductRequest{ProductId: productId, AccountId: accountId})
	return err
}

func (client *Client) SearchProducts(ctx context.Context, query string, skip, take uint64, priceRange *models.PriceRange, category string, sortOrder string) ([]models.Product, error) {
	// For now, we'll use the existing GetProducts method and filter the results in the GraphQL resolver
	// In a real implementation, you would update the product service to support these filters
	res, err := client.service.GetProducts(ctx, &pb.GetProductsRequest{
		Skip:  skip,
		Take:  take,
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	// Convert the results to models.Product
	var products []models.Product
	for _, p := range res.Products {
		products = append(products, models.Product{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			AccountID:   int(p.AccountId),
			// Category is not available in the current implementation
			Category:    category,
		})
	}

	// Apply client-side filtering for price range
	if priceRange != nil {
		log.Printf("Applying price range filter: Min=%v, Max=%v", priceRange.Min, priceRange.Max)
		log.Printf("Before filtering: %d products", len(products))
		var filteredProducts []models.Product
		for _, product := range products {
			if (priceRange.Min <= 0 || product.Price >= priceRange.Min) &&
			   (priceRange.Max <= 0 || product.Price <= priceRange.Max) {
				filteredProducts = append(filteredProducts, product)
				log.Printf("Product %s with price %.2f passes filter", product.Name, product.Price)
			} else {
				log.Printf("Product %s with price %.2f filtered out", product.Name, product.Price)
			}
		}
		log.Printf("After filtering: %d products", len(filteredProducts))
		products = filteredProducts
	}

	// Apply client-side filtering for category
	if category != "" {
		log.Printf("Applying category filter: %s", category)
		log.Printf("Before filtering: %d products", len(products))
		var filteredProducts []models.Product
		for _, product := range products {
			log.Printf("Product %s has category '%s'", product.Name, product.Category)
			if product.Category == category {
				filteredProducts = append(filteredProducts, product)
				log.Printf("Product %s passes category filter", product.Name)
			} else {
				log.Printf("Product %s filtered out by category", product.Name)
			}
		}
		log.Printf("After filtering: %d products", len(filteredProducts))
		products = filteredProducts
	}

	// Implement sorting
	if sortOrder != "" {
		log.Printf("Applying sort order: %s", sortOrder)
		log.Printf("Before sorting: %v", products)

		switch sortOrder {
		case "PRICE_ASC":
			log.Printf("Sorting by price ascending")
			// Sort by price ascending
			for i := 0; i < len(products)-1; i++ {
				for j := i + 1; j < len(products); j++ {
					if products[i].Price > products[j].Price {
						products[i], products[j] = products[j], products[i]
					}
				}
			}
		case "PRICE_DESC":
			log.Printf("Sorting by price descending")
			// Sort by price descending
			for i := 0; i < len(products)-1; i++ {
				for j := i + 1; j < len(products); j++ {
					if products[i].Price < products[j].Price {
						products[i], products[j] = products[j], products[i]
					}
				}
			}
		case "NEWEST":
			log.Printf("Sorting by newest (ID as proxy)")
			// Sort by ID as a proxy for creation time (newest first)
			for i := 0; i < len(products)-1; i++ {
				for j := i + 1; j < len(products); j++ {
					if products[i].ID < products[j].ID {
						products[i], products[j] = products[j], products[i]
					}
				}
			}
		case "POPULARITY":
			// This would require additional data, for now we'll just log it
			log.Printf("Sorting by popularity not implemented")
		}

		log.Printf("After sorting: %v", products)
	}

	return products, nil
}
