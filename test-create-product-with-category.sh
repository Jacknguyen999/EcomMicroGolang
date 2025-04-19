#!/bin/bash

# Create a product with a category using a direct mutation
curl -X POST -H "Content-Type: application/json" --data '{
  "query": "mutation { createProduct(product: { name: \"Gaming Keyboard RGB\", description: \"Mechanical gaming keyboard with RGB lighting\", price: 89.99, category: \"electronics\" }) { id name description price category } }"
}' http://localhost:8080/graphql
