package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/thomas/EcommerceAPI/pkg/auth"
	"github.com/thomas/EcommerceAPI/product/models"
)

type queryResolver struct {
	server *Server
}

func (resolver *queryResolver) Accounts(
	ctx context.Context,
	pagination *PaginationInput,
	id *string,
) ([]*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Check authentication - use abort=true to properly enforce authentication
	accountId, err := auth.GetUserIdInt(ctx, true)
	if err != nil {
		log.Println("Authentication failed for Accounts query:", err)
		return nil, errors.New("unauthorized: you must be logged in to access account information")
	}

	// Convert accountId to string for comparison
	accountIdStr := strconv.Itoa(accountId)

	// If ID is provided, get specific account
	if id != nil {
		// Only allow users to access their own account unless implementing admin functionality
		if *id != accountIdStr {
			return nil, errors.New("unauthorized: you can only access your own account")
		}

		res, err := resolver.server.accountClient.GetAccount(ctx, *id, accountIdStr)
		if err != nil {
			log.Println("Error getting account:", err)
			return nil, err
		}
		return []*Account{{
			ID:    strconv.Itoa(int(res.ID)),
			Name:  res.Name,
			Email: res.Email,
		}}, nil
	}

	// For listing accounts, we'll only return the current user's account
	// In a real app, you might implement admin functionality here
	res, err := resolver.server.accountClient.GetAccount(ctx, accountIdStr, accountIdStr)
	if err != nil {
		log.Println("Error getting account:", err)
		return nil, err
	}

	return []*Account{{
		ID:    strconv.Itoa(int(res.ID)),
		Name:  res.Name,
		Email: res.Email,
	}}, nil
}

func (resolver *queryResolver) Product(
	ctx context.Context,
	pagination *PaginationInput,
	query, id *string,
	viewedProductsIds []*string,
	byAccountId *bool,
	ownedByMe *bool,
) ([]*Product, error) {
	// These parameters are not in the interface yet
	var priceRange *PriceRangeInput
	var category *string
	var sortBy *SortOrder

	// Extract the parameters from the context
	if opCtx := graphql.GetOperationContext(ctx); opCtx != nil {
		// First try to get parameters from variables (for variable-based queries)
		if opCtx.Variables != nil {
			if val, ok := opCtx.Variables["priceRange"]; ok {
				if prMap, ok := val.(map[string]interface{}); ok {
					var min, max float64
					if minVal, ok := prMap["min"]; ok {
						switch v := minVal.(type) {
						case float64:
							min = v
						case float32:
							min = float64(v)
						case int:
							min = float64(v)
						case int64:
							min = float64(v)
						case json.Number:
							min, _ = v.Float64()
						}
					}
					if maxVal, ok := prMap["max"]; ok {
						switch v := maxVal.(type) {
						case float64:
							max = v
						case float32:
							max = float64(v)
						case int:
							max = float64(v)
						case int64:
							max = float64(v)
						case json.Number:
							max, _ = v.Float64()
						}
					}
					priceRange = &PriceRangeInput{
						Min: min,
						Max: max,
					}
				}
			}
			if val, ok := opCtx.Variables["category"]; ok {
				if catStr, ok := val.(string); ok {
					category = &catStr
				}
			}
			if val, ok := opCtx.Variables["sortBy"]; ok {
				if sortStr, ok := val.(string); ok {
					sortEnum := SortOrder(sortStr)
					sortBy = &sortEnum
				}
			}
		}

		// We'll skip the direct query parameter extraction for now
		// as it requires more complex type handling
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Get single product by ID - public operation
	if id != nil {
		res, err := resolver.server.productClient.GetProduct(ctx, *id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return []*Product{{
			ID:          res.ID,
			Name:        res.Name,
			Description: res.Description,
			Price:       res.Price,
			AccountID:   res.AccountID,
		}}, nil
	}

	skip, take := uint64(0), uint64(0)
	if pagination != nil {
		skip, take = pagination.bounds()
	}

	// Get recommendations based on viewed products - public operation
	if viewedProductsIds != nil {
		log.Printf("Getting recommendations based on viewed products: %v, skip: %d, take: %d", viewedProductsIds, skip, take)

		// Extract product IDs from the input
		viewedIds := make([]string, 0, len(viewedProductsIds))
		for _, id := range viewedProductsIds {
			if id != nil {
				viewedIds = append(viewedIds, *id)
			}
		}

		// Get all products
		log.Println("Fetching all products from database...")
		allProducts, err := resolver.server.productClient.GetProducts(ctx, 0, 100, nil, "")
		if err != nil {
			log.Println("Error fetching products:", err)
			return nil, err
		}
		log.Printf("Found %d products in database", len(allProducts))

		// Filter out the viewed products
		var filteredProducts []models.Product
		for _, product := range allProducts {
			viewed := false
			for _, viewedId := range viewedIds {
				if product.ID == viewedId {
					viewed = true
					break
				}
			}
			if !viewed {
				filteredProducts = append(filteredProducts, product)
			}
		}
		log.Printf("After filtering out viewed products, %d products remain", len(filteredProducts))

		// Apply pagination
		end := int(skip) + int(take)
		if end > len(filteredProducts) {
			end = len(filteredProducts)
		}
		start := int(skip)
		if start >= len(filteredProducts) {
			start = 0
			end = 0
		}

		var paginatedProducts []models.Product
		if start < end {
			paginatedProducts = filteredProducts[start:end]
		}
		log.Printf("After pagination (skip=%d, take=%d), returning %d products", skip, take, len(paginatedProducts))

		// Convert to GraphQL products
		var products []*Product
		for _, product := range paginatedProducts {
			products = append(products, &Product{
				ID:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				AccountID:   product.AccountID,
			})
		}

		log.Printf("Returning %d products with skip=%d, take=%d", len(products), skip, take)
		return products, nil
	}

	// Get products owned by the current user - protected operation
	if ownedByMe != nil && *ownedByMe {
		// Enforce authentication
		accountIdStr := auth.GetUserId(ctx, true)
		if accountIdStr == "" {
			log.Println("Authentication failed for ownedByMe query: empty account ID")
			return nil, errors.New("unauthorized: you must be logged in to view your products")
		}

		// Convert string account ID to int
		accountId, err := strconv.Atoi(accountIdStr)
		if err != nil {
			log.Printf("Error converting account ID '%s' to integer: %v", accountIdStr, err)
			return nil, errors.New("error processing user ID")
		}

		log.Printf("Filtering products for account ID: %d", accountId)

		// Use the direct query to get all products
		productList, err := resolver.server.productClient.GetProducts(ctx, 0, 100, nil, "")
		if err != nil {
			log.Println("Error getting products:", err)
			return nil, err
		}

		log.Printf("Found %d total products", len(productList))

		// Filter products by account ID
		var filteredProducts []models.Product
		for _, product := range productList {
			log.Printf("Product %s has account ID %d (comparing with %d)", product.ID, product.AccountID, accountId)
			if product.AccountID == accountId {
				filteredProducts = append(filteredProducts, product)
			}
		}

		log.Printf("After filtering, found %d products owned by user", len(filteredProducts))

		// Convert to GraphQL products
		var products []*Product
		for _, product := range filteredProducts {
			products = append(products, &Product{
				ID:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				AccountID:   product.AccountID,
			})
		}

		return products, nil
	}

	// Get recommendations for a specific user - protected operation
	if byAccountId != nil && *byAccountId {
		// Enforce authentication
		accountId := auth.GetUserId(ctx, true)
		if accountId == "" {
			return nil, errors.New("unauthorized: you must be logged in to get personalized recommendations")
		}

		skip = 0
		take = 100
		res, err := resolver.server.recommenderClient.GetRecommendationForUser(ctx, accountId, skip, take)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		productList := res.GetRecommendedProducts()
		var products []*Product
		for _, product := range productList {
			products = append(products,
				&Product{
					ID:          product.Id,
					Name:        product.Name,
					Description: product.Description,
					Price:       product.Price,
					// AccountID is not available in ProductReplica
					AccountID:   0,
				},
			)
		}
		return products, nil
	}

	// Search products - public operation
	q := ""
	if query != nil {
		q = *query
	}

	// Convert GraphQL price range to product service price range
	var productPriceRange *models.PriceRange
	if priceRange != nil {
		log.Printf("GraphQL received price range: Min=%v, Max=%v", priceRange.Min, priceRange.Max)
		productPriceRange = &models.PriceRange{
			Min: priceRange.Min,
			Max: priceRange.Max,
		}
		log.Printf("Converted to product price range: Min=%v, Max=%v", productPriceRange.Min, productPriceRange.Max)
	}

	// Category and sort order are handled directly in the resolver

	// Call product service with basic search
	productList, err := resolver.server.productClient.GetProducts(ctx, skip, take, nil, q)
	if err != nil {
		log.Println("Error searching products:", err)
		return nil, err
	}

	// Convert to GraphQL products
	var products []*Product
	for _, product := range productList {
		products = append(products,
			&Product{
				ID:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				AccountID:   product.AccountID,
				Category:    product.Category,
			},
		)
	}

	// Apply price range filtering directly in the resolver
	if priceRange != nil {
		log.Printf("Applying price range filter: Min=%v, Max=%v", priceRange.Min, priceRange.Max)
		var filteredProducts []*Product
		for _, product := range products {
			if (priceRange.Min <= 0 || product.Price >= priceRange.Min) &&
			   (priceRange.Max <= 0 || product.Price <= priceRange.Max) {
				filteredProducts = append(filteredProducts, product)
			}
		}
		products = filteredProducts
	}

	// Apply category filtering
	if category != nil && *category != "" {
		log.Printf("Applying category filter: %s", *category)

		// Try to directly query Elasticsearch for products with the specified category
		query := fmt.Sprintf(`{"query": {"term": {"category": "%s"}}}`, *category)

		// Make a direct HTTP request to Elasticsearch
		resp, err := http.Post("http://product_db:9200/catalog/product/_search", "application/json", strings.NewReader(query))
		if err != nil {
			log.Printf("Error querying Elasticsearch: %v", err)
		} else {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading Elasticsearch response: %v", err)
			} else {
				// Parse the response
				var result map[string]interface{}
				if err := json.Unmarshal(body, &result); err != nil {
					log.Printf("Error parsing Elasticsearch response: %v", err)
				} else {
					// Extract the hits
					if hits, ok := result["hits"].(map[string]interface{}); ok {
						if hitsArray, ok := hits["hits"].([]interface{}); ok {
							log.Printf("Found %d products with category '%s'", len(hitsArray), *category)

							// Create a new products array
							var filteredProducts []*Product

							// Add products from Elasticsearch
							for _, hit := range hitsArray {
								hitMap := hit.(map[string]interface{})
								id := hitMap["_id"].(string)
								source := hitMap["_source"].(map[string]interface{})

								name := source["name"].(string)
								description := source["description"].(string)
								price := source["price"].(float64)
								accountID := int(source["accountID"].(float64))
								category := source["category"].(string)

								filteredProducts = append(filteredProducts, &Product{
									ID:          id,
									Name:        name,
									Description: description,
									Price:       price,
									AccountID:   accountID,
									Category:    category,
								})
							}

							products = filteredProducts
							return products, nil
						}
					}
				}
			}
		}

		// Fall back to client-side filtering if Elasticsearch query fails
		var filteredProducts []*Product
		for _, product := range products {
			if product.Category == *category {
				filteredProducts = append(filteredProducts, product)
			}
		}
		products = filteredProducts
	}

	// Apply sorting
	if sortBy != nil {
		log.Printf("Applying sort order: %s", *sortBy)

		switch *sortBy {
		case SortOrderPriceAsc:
			// Use sort.Slice for more reliable sorting
			sort.Slice(products, func(i, j int) bool {
				return products[i].Price < products[j].Price
			})
		case SortOrderPriceDesc:
			sort.Slice(products, func(i, j int) bool {
				return products[i].Price > products[j].Price
			})
		case SortOrderNewest:
			sort.Slice(products, func(i, j int) bool {
				return products[i].ID > products[j].ID
			})
		case SortOrderPopularity:
			// This would require additional data, for now we'll just log it
			log.Printf("Sorting by popularity not implemented")
		}
	}

	return products, nil
}

func (pagination PaginationInput) bounds() (uint64, uint64) {
	skipValue := uint64(0)
	takeValue := uint64(100)
	if pagination.Skip != 0 {
		skipValue = uint64(pagination.Skip)
	}
	if pagination.Take != 100 {
		takeValue = uint64(pagination.Take)
	}
	return skipValue, takeValue
}
