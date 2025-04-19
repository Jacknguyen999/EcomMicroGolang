package client

import (
	"context"
	"log"
	"sort"
	"strings"

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
	// Use the repository's direct search capabilities
	log.Printf("Searching products with query=%s, priceRange=%v, category=%s, sortOrder=%s", query, priceRange, category, sortOrder)

	// Get all products first to ensure we have the most up-to-date data
	allProducts, err := client.GetProducts(ctx, 0, 1000, nil, "")
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d total products in the database", len(allProducts))

	// Apply filtering manually to ensure we include all products
	var filteredProducts []models.Product

	// Filter by query (name/description)
	if query != "" {
		for _, product := range allProducts {
			// Simple case-insensitive substring search
			if strings.Contains(strings.ToLower(product.Name), strings.ToLower(query)) ||
			   strings.Contains(strings.ToLower(product.Description), strings.ToLower(query)) {
				filteredProducts = append(filteredProducts, product)
			}
		}
	} else {
		// If no query, include all products
		filteredProducts = allProducts
	}

	// Apply price range filtering
	if priceRange != nil && (priceRange.Min > 0 || priceRange.Max > 0) {
		log.Printf("Applying price range filter: Min=%v, Max=%v", priceRange.Min, priceRange.Max)
		var priceFilteredProducts []models.Product

		for _, product := range filteredProducts {
			passesFilter := true

			// Apply minimum price filter if set
			if priceRange.Min > 0 && product.Price < priceRange.Min {
				passesFilter = false
			}

			// Apply maximum price filter if set
			if priceRange.Max > 0 && product.Price > priceRange.Max {
				passesFilter = false
			}

			if passesFilter {
				priceFilteredProducts = append(priceFilteredProducts, product)
				log.Printf("Product %s with price %.2f passes price filter", product.Name, product.Price)
			} else {
				log.Printf("Product %s with price %.2f filtered out by price", product.Name, product.Price)
			}
		}

		filteredProducts = priceFilteredProducts
		log.Printf("After price filtering: %d products", len(filteredProducts))
	}

	// Apply category filtering
	if category != "" {
		log.Printf("Applying category filter: %s", category)
		var categoryFilteredProducts []models.Product

		for _, product := range filteredProducts {
			if product.Category == category {
				categoryFilteredProducts = append(categoryFilteredProducts, product)
				log.Printf("Product %s passes category filter", product.Name)
			} else {
				log.Printf("Product %s filtered out by category", product.Name)
			}
		}

		filteredProducts = categoryFilteredProducts
		log.Printf("After category filtering: %d products", len(filteredProducts))
	}

	// Apply sorting
	if sortOrder != "" {
		log.Printf("Applying sort order: %s", sortOrder)

		switch sortOrder {
		case "PRICE_ASC":
			sort.Slice(filteredProducts, func(i, j int) bool {
				return filteredProducts[i].Price < filteredProducts[j].Price
			})
			log.Printf("Sorted products by price ascending")

		case "PRICE_DESC":
			sort.Slice(filteredProducts, func(i, j int) bool {
				return filteredProducts[i].Price > filteredProducts[j].Price
			})
			log.Printf("Sorted products by price descending")

		case "NEWEST":
			sort.Slice(filteredProducts, func(i, j int) bool {
				return filteredProducts[i].ID > filteredProducts[j].ID
			})
			log.Printf("Sorted products by newest first")

		case "POPULARITY":
			log.Printf("Sorting by popularity not implemented")
		}
	}

	// Apply pagination
	start := int(skip)
	end := int(skip + take)

	if start >= len(filteredProducts) {
		return []models.Product{}, nil
	}

	if end > len(filteredProducts) {
		end = len(filteredProducts)
	}

	log.Printf("Returning products %d to %d (total: %d)", start, end, len(filteredProducts))
	return filteredProducts[start:end], nil
}
