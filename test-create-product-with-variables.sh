#!/bin/bash

# Create a product with a category using variables
curl -X POST -H "Content-Type: application/json" --data '{
  "query": "mutation CreateProduct($input: CreateProductInput!) { createProduct(product: $input) { id name description price category } }",
  "variables": {
    "input": {
      "name": "Gaming Headset Pro V2",
      "description": "High-quality gaming headset with surround sound and RGB lighting",
      "price": 149.99,
      "category": "electronics"
    }
  }
}' http://localhost:8080/graphql
