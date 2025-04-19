#!/usr/bin/env python3
import requests
import json
import time
import sys
import random
from datetime import datetime

# GraphQL endpoint
GRAPHQL_URL = "http://localhost:8080/graphql"

def print_header(text):
    print(f"\n{'=' * 10} {text} {'=' * 10}")

def print_subheader(text):
    print(f"\n{'-' * 5} {text} {'-' * 5}")

def print_success(text):
    print(f"✅ {text}")

def print_error(text):
    print(f"❌ {text}")

def print_info(text):
    print(f"ℹ️ {text}")

def execute_query(query, variables=None, token=None):
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    
    payload = {"query": query}
    if variables:
        payload["variables"] = variables
    
    try:
        response = requests.post(GRAPHQL_URL, json=payload, headers=headers)
        return response.json()
    except Exception as e:
        print_error(f"Failed to execute query: {str(e)}")
        return {"errors": [{"message": str(e)}]}

def login(email, password):
    print_info(f"Logging in as {email}")
    
    query = """
    mutation Login($account: LoginInput!) {
        login(account: $account) {
            token
        }
    }
    """
    
    variables = {
        "account": {
            "email": email,
            "password": password
        }
    }
    
    result = execute_query(query, variables)
    
    if "errors" in result:
        print_error(f"Failed to login: {result['errors'][0]['message']}")
        return None
    
    if "data" in result and result["data"]["login"]:
        token = result["data"]["login"]["token"]
        print_success(f"Successfully logged in as {email}")
        return token
    
    print_error("Login response did not contain a token")
    return None

def register(name, email, password):
    print_info(f"Registering user: {name} - {email}")
    
    query = """
    mutation Register($account: RegisterInput!) {
        register(account: $account) {
            token
        }
    }
    """
    
    variables = {
        "account": {
            "name": name,
            "email": email,
            "password": password
        }
    }
    
    result = execute_query(query, variables)
    
    if "errors" in result:
        print_error(f"Failed to register: {result['errors'][0]['message']}")
        return None
    
    if "data" in result and result["data"]["register"]:
        token = result["data"]["register"]["token"]
        print_success(f"Successfully registered {name}")
        return token
    
    print_error("Register response did not contain a token")
    return None

def create_product(token, name, description, price):
    print_info(f"Creating product: {name} - ${price}")
    
    query = """
    mutation CreateProduct($product: CreateProductInput!) {
        createProduct(product: $product) {
            id
            name
            description
            price
        }
    }
    """
    
    variables = {
        "product": {
            "name": name,
            "description": description,
            "price": float(price)
        }
    }
    
    result = execute_query(query, variables, token)
    
    if "errors" in result:
        print_error(f"Failed to create product: {result['errors'][0]['message']}")
        return None
    
    if "data" in result and result["data"]["createProduct"]:
        product = result["data"]["createProduct"]
        print_success(f"Created product: {product['id']} - {product['name']}")
        return product
    
    print_error("Create product response did not contain product data")
    return None

def get_products(token=None):
    print_info("Getting all products")
    
    query = """
    query GetProducts {
        product(pagination: { skip: 0, take: 100 }) {
            id
            name
            description
            price
        }
    }
    """
    
    result = execute_query(query, token=token)
    
    if "errors" in result:
        print_error(f"Failed to get products: {result['errors'][0]['message']}")
        return []
    
    if "data" in result and result["data"]["product"]:
        products = result["data"]["product"]
        print_success(f"Found {len(products)} products")
        return products
    
    print_error("Get products response did not contain product data")
    return []

def get_recommendations_by_viewed(product_ids, skip=0, take=5):
    print_info(f"Getting recommendations based on viewed products: {', '.join(product_ids)}")
    print_info(f"Pagination: skip={skip}, take={take}")
    
    query = """
    query GetRecommendations($ids: [String], $skip: Int!, $take: Int!) {
        product(viewedProductsIds: $ids, pagination: { skip: $skip, take: $take }) {
            id
            name
            description
            price
        }
    }
    """
    
    variables = {
        "ids": product_ids,
        "skip": skip,
        "take": take
    }
    
    result = execute_query(query, variables)
    
    if "errors" in result:
        print_error(f"Failed to get recommendations: {result['errors'][0]['message']}")
        return []
    
    if "data" in result and result["data"]["product"]:
        recommendations = result["data"]["product"]
        print_success(f"Received {len(recommendations)} recommendations")
        
        for i, product in enumerate(recommendations):
            print(f"  {i+1}. {product['name']} - ${product['price']} - {product['description']}")
        
        return recommendations
    
    print_error("Get recommendations response did not contain product data")
    return []

def get_personalized_recommendations(token):
    print_info("Getting personalized recommendations")
    
    query = """
    query GetPersonalizedRecommendations {
        product(byAccountId: true) {
            id
            name
            description
            price
        }
    }
    """
    
    result = execute_query(query, token=token)
    
    if "errors" in result:
        print_error(f"Failed to get personalized recommendations: {result['errors'][0]['message']}")
        return []
    
    if "data" in result and result["data"]["product"]:
        recommendations = result["data"]["product"]
        print_success(f"Received {len(recommendations)} personalized recommendations")
        
        for i, product in enumerate(recommendations):
            print(f"  {i+1}. {product['name']} - ${product['price']} - {product['description']}")
        
        return recommendations
    
    print_error("Get personalized recommendations response did not contain product data")
    return []

def create_order(token, product_ids):
    print_info(f"Creating order for products: {', '.join(product_ids)}")
    
    query = """
    mutation CreateOrder($order: OrderInput!) {
        createOrder(order: $order) {
            id
            totalPrice
            products {
                id
                name
                price
                quantity
            }
        }
    }
    """
    
    order_products = []
    for product_id in product_ids:
        order_products.append({
            "id": product_id,
            "quantity": 1
        })
    
    variables = {
        "order": {
            "products": order_products
        }
    }
    
    result = execute_query(query, variables, token)
    
    if "errors" in result:
        print_error(f"Failed to create order: {result['errors'][0]['message']}")
        return None
    
    if "data" in result and result["data"]["createOrder"]:
        order = result["data"]["createOrder"]
        print_success(f"Created order: {order['id']} with total price: ${order['totalPrice']}")
        return order
    
    print_error("Create order response did not contain order data")
    return None

def run_tests():
    print_header("RECOMMENDER SYSTEM COMPREHENSIVE TEST SUITE")
    print(f"Started at: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    
    # Step 1: Register or login test users
    print_header("USER AUTHENTICATION")
    
    # Generate unique email to avoid conflicts
    timestamp = int(time.time())
    test_email = f"test{timestamp}@example.com"
    
    token = register("Test User", test_email, "password123")
    if not token:
        # Try to login with a default test account
        token = login("test@example.com", "password123")
    
    if not token:
        print_error("Failed to authenticate. Aborting tests.")
        return
    
    # Step 2: Get existing products
    print_header("PRODUCT INVENTORY")
    products = get_products()
    
    if len(products) < 5:
        print_info("Not enough products for comprehensive testing. Creating additional test products...")
        
        # Create some test products with unique names
        test_products = [
            (f"Gaming Laptop {timestamp}", "High-performance gaming laptop with RTX 3080", 1999.99),
            (f"Smartphone {timestamp}", "Latest smartphone with 5G capabilities", 899.99),
            (f"Wireless Headphones {timestamp}", "Noise-cancelling wireless headphones", 249.99),
            (f"Smart Watch {timestamp}", "Fitness tracking smart watch", 199.99),
            (f"Tablet {timestamp}", "10-inch tablet with high-resolution display", 499.99)
        ]
        
        for name, description, price in test_products:
            product = create_product(token, name, description, price)
            if product:
                products.append(product)
    
    if len(products) < 3:
        print_error("Not enough products available for testing. Aborting tests.")
        return
    
    # Display available products
    print_subheader("AVAILABLE PRODUCTS")
    for i, product in enumerate(products):
        print(f"{i+1}. {product['name']} - ${product['price']} - {product['description']}")
    
    # Step 3: Create some orders to generate interactions
    print_header("USER INTERACTIONS")
    
    # Create orders for a subset of products
    if len(products) >= 3:
        create_order(token, [products[0]["id"], products[1]["id"]])
        time.sleep(1)  # Wait between orders
        create_order(token, [products[2]["id"]])
    
    # Wait for the system to process the interactions
    print_info("Waiting for the system to process interactions...")
    time.sleep(3)
    
    # Step 4: Test recommendations based on viewed products
    print_header("RECOMMENDATIONS BASED ON VIEWED PRODUCTS")
    
    # Test case 1: Get recommendations based on viewing a single product
    print_subheader("TEST CASE 1: Single Product View")
    if products:
        get_recommendations_by_viewed([products[0]["id"]], 0, 3)
    
    # Test case 2: Get recommendations based on viewing multiple products
    print_subheader("TEST CASE 2: Multiple Product Views")
    if len(products) >= 2:
        get_recommendations_by_viewed([products[0]["id"], products[1]["id"]], 0, 5)
    
    # Test case 3: Test pagination - skip first result
    print_subheader("TEST CASE 3: Pagination - Skip First Result")
    if products:
        get_recommendations_by_viewed([products[0]["id"]], 1, 2)
    
    # Test case 4: Test pagination - larger skip value
    print_subheader("TEST CASE 4: Pagination - Larger Skip Value")
    if products:
        get_recommendations_by_viewed([products[0]["id"]], 3, 2)
    
    # Test case 5: Test with larger take value
    print_subheader("TEST CASE 5: Larger Take Value")
    if products:
        get_recommendations_by_viewed([products[0]["id"]], 0, 10)
    
    # Test case 6: Test with random product
    print_subheader("TEST CASE 6: Random Product")
    if len(products) >= 3:
        random_product = random.choice(products)
        get_recommendations_by_viewed([random_product["id"]], 0, 3)
    
    # Step 5: Test personalized recommendations
    print_header("PERSONALIZED RECOMMENDATIONS")
    get_personalized_recommendations(token)
    
    print_header("TEST SUMMARY")
    print("All tests completed successfully!")
    print(f"Finished at: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")

if __name__ == "__main__":
    # Check if URL was provided as command line argument
    if len(sys.argv) > 1:
        GRAPHQL_URL = sys.argv[1]
    
    run_tests()
