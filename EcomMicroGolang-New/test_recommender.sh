#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}===== Testing Recommender System =====${NC}"

# Function to add a product
add_product() {
    local name=$1
    local description=$2
    local price=$3
    
    echo -e "${YELLOW}Adding product: $name - $price${NC}"
    
    response=$(curl -s -X POST -H "Content-Type: application/json" --data "{
        \"query\": \"mutation { createProduct(product: { name: \\\"$name\\\", description: \\\"$description\\\", price: $price }) { id name description price } }\"
    }" http://localhost:8080/graphql)
    
    # Extract product ID from response
    product_id=$(echo $response | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    
    if [ -n "$product_id" ]; then
        echo -e "${GREEN}Successfully added product with ID: $product_id${NC}"
        echo $product_id
    else
        echo -e "${RED}Failed to add product. Response: $response${NC}"
        echo ""
    fi
}

# Function to register a user
register_user() {
    local name=$1
    local email=$2
    local password=$3
    
    echo -e "${YELLOW}Registering user: $name - $email${NC}"
    
    response=$(curl -s -X POST -H "Content-Type: application/json" --data "{
        \"query\": \"mutation { register(account: { name: \\\"$name\\\", email: \\\"$email\\\", password: \\\"$password\\\" }) { token } }\"
    }" http://localhost:8080/graphql)
    
    # Extract token from response
    token=$(echo $response | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$token" ]; then
        echo -e "${GREEN}Successfully registered user. Token: $token${NC}"
        echo $token
    else
        echo -e "${RED}Failed to register user. Response: $response${NC}"
        echo ""
    fi
}

# Function to create an order (simulates user interaction)
create_order() {
    local token=$1
    local product_id=$2
    local quantity=$3
    
    echo -e "${YELLOW}Creating order for product: $product_id, quantity: $quantity${NC}"
    
    response=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $token" --data "{
        \"query\": \"mutation { createOrder(order: { products: [{ id: \\\"$product_id\\\", quantity: $quantity }] }) { id totalPrice } }\"
    }" http://localhost:8080/graphql)
    
    # Extract order ID from response
    order_id=$(echo $response | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    
    if [ -n "$order_id" ]; then
        echo -e "${GREEN}Successfully created order with ID: $order_id${NC}"
        echo $order_id
    else
        echo -e "${RED}Failed to create order. Response: $response${NC}"
        echo ""
    fi
}

# Function to get recommendations based on viewed products
get_recommendations_by_viewed() {
    local product_ids=$1
    local skip=$2
    local take=$3
    
    echo -e "${YELLOW}Getting recommendations based on viewed products: $product_ids, skip: $skip, take: $take${NC}"
    
    response=$(curl -s -X POST -H "Content-Type: application/json" --data "{
        \"query\": \"query { product(viewedProductsIds: $product_ids, pagination: { skip: $skip, take: $take }) { id name description price } }\"
    }" http://localhost:8080/graphql)
    
    echo -e "${GREEN}Recommendations:${NC}"
    echo $response | jq -r '.data.product[] | "- \(.name): \(.price) - \(.description)"' 2>/dev/null || echo $response
    echo ""
}

# Function to get personalized recommendations
get_personalized_recommendations() {
    local token=$1
    
    echo -e "${YELLOW}Getting personalized recommendations${NC}"
    
    response=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $token" --data "{
        \"query\": \"query { product(byAccountId: true) { id name description price } }\"
    }" http://localhost:8080/graphql)
    
    echo -e "${GREEN}Personalized Recommendations:${NC}"
    echo $response | jq -r '.data.product[] | "- \(.name): \(.price) - \(.description)"' 2>/dev/null || echo $response
    echo ""
}

# Add test products
echo -e "${YELLOW}===== Adding Test Products =====${NC}"
product1=$(add_product "Gaming Laptop" "High-performance gaming laptop with RTX 3080" 1999.99)
product2=$(add_product "Smartphone" "Latest smartphone with 5G capabilities" 899.99)
product3=$(add_product "Wireless Headphones" "Noise-cancelling wireless headphones" 249.99)
product4=$(add_product "Smart Watch" "Fitness tracking smart watch" 199.99)
product5=$(add_product "Tablet" "10-inch tablet with high-resolution display" 499.99)
product6=$(add_product "Bluetooth Speaker" "Portable bluetooth speaker with 20h battery life" 129.99)
product7=$(add_product "Mechanical Keyboard" "RGB mechanical keyboard for gaming" 149.99)
product8=$(add_product "Wireless Mouse" "Ergonomic wireless mouse" 79.99)
product9=$(add_product "External SSD" "1TB external SSD with USB-C" 179.99)
product10=$(add_product "Monitor" "27-inch 4K monitor" 349.99)

# Register test users
echo -e "${YELLOW}===== Registering Test Users =====${NC}"
user1_token=$(register_user "John Doe" "john@example.com" "password123")
user2_token=$(register_user "Jane Smith" "jane@example.com" "password456")

# Create orders to simulate user interactions
echo -e "${YELLOW}===== Creating Test Orders =====${NC}"
if [ -n "$user1_token" ] && [ -n "$product1" ]; then
    create_order "$user1_token" "$product1" 1
fi

if [ -n "$user1_token" ] && [ -n "$product3" ]; then
    create_order "$user1_token" "$product3" 1
fi

if [ -n "$user2_token" ] && [ -n "$product2" ]; then
    create_order "$user2_token" "$product2" 1
fi

if [ -n "$user2_token" ] && [ -n "$product5" ]; then
    create_order "$user2_token" "$product5" 1
fi

# Test recommendations based on viewed products
echo -e "${YELLOW}===== Testing Recommendations Based on Viewed Products =====${NC}"

# Test case 1: Recommendations based on viewing product 1
if [ -n "$product1" ]; then
    get_recommendations_by_viewed "[\\\"$product1\\\"]" 0 3
fi

# Test case 2: Recommendations based on viewing products 1 and 2
if [ -n "$product1" ] && [ -n "$product2" ]; then
    get_recommendations_by_viewed "[\\\"$product1\\\", \\\"$product2\\\"]" 0 5
fi

# Test case 3: Recommendations with pagination
if [ -n "$product3" ]; then
    get_recommendations_by_viewed "[\\\"$product3\\\"]" 2 3
fi

# Test personalized recommendations
echo -e "${YELLOW}===== Testing Personalized Recommendations =====${NC}"
if [ -n "$user1_token" ]; then
    get_personalized_recommendations "$user1_token"
fi

if [ -n "$user2_token" ]; then
    get_personalized_recommendations "$user2_token"
fi

echo -e "${GREEN}===== Testing Complete =====${NC}"
