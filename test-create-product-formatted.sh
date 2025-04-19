#!/bin/bash

# Create a product with a category
echo "Creating product with category 'electronics'..."
RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" --data '{
  "query": "mutation { createProduct(product: { name: \"Gaming Keyboard RGB\", description: \"Mechanical gaming keyboard with RGB lighting\", price: 89.99, category: \"electronics\" }) { id name description price category } }"
}' http://localhost:8080/graphql)

# Display the response
echo "Response:"
echo "$RESPONSE"

# Extract the category field
CATEGORY=$(echo "$RESPONSE" | grep -o '"category":"[^"]*"' | cut -d'"' -f4)
echo "Category value: $CATEGORY"
