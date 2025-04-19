#!/bin/bash

# Filter products by category
echo "Filtering products by category 'electronics'..."
RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" --data '{
  "query": "query { product(pagination: {skip: 0, take: 10}, category: \"electronics\") { id name price category } }"
}' http://localhost:8080/graphql)

# Display the response
echo "Response:"
echo "$RESPONSE"

# Count the number of products
COUNT=$(echo "$RESPONSE" | grep -o '"id"' | wc -l)
echo "Number of products found: $COUNT"
