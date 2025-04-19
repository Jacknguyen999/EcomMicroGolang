#!/bin/bash

# Create a product with a category
curl -X POST -H "Content-Type: application/json" --data '{
  "query": "mutation { createProduct(product: { name: \"Gaming Headset Pro\", description: \"High-quality gaming headset with surround sound\", price: 129.99, category: \"electronics\" }) { id name description price category } }"
}' http://localhost:8080/graphql
