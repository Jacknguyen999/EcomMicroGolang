#!/usr/bin/env python3
import requests
import json
import time
import random
import argparse
from colorama import Fore, Style, init

# Initialize colorama
init()

# GraphQL endpoint
GRAPHQL_URL = "http://localhost:8080/graphql"

class RecommenderTester:
    def __init__(self):
        self.products = []
        self.users = []
        self.interactions = []
    
    def print_header(self, text):
        print(f"\n{Fore.YELLOW}{'=' * 10} {text} {'=' * 10}{Style.RESET_ALL}")
    
    def print_success(self, text):
        print(f"{Fore.GREEN}{text}{Style.RESET_ALL}")
    
    def print_error(self, text):
        print(f"{Fore.RED}{text}{Style.RESET_ALL}")
    
    def print_info(self, text):
        print(f"{Fore.CYAN}{text}{Style.RESET_ALL}")
    
    def execute_query(self, query, variables=None, token=None):
        headers = {"Content-Type": "application/json"}
        if token:
            headers["Authorization"] = f"Bearer {token}"
        
        payload = {"query": query}
        if variables:
            payload["variables"] = variables
        
        response = requests.post(GRAPHQL_URL, json=payload, headers=headers)
        
        try:
            return response.json()
        except:
            self.print_error(f"Failed to parse response: {response.text}")
            return {"errors": [{"message": "Failed to parse response"}]}
    
    def add_product(self, name, description, price):
        self.print_info(f"Adding product: {name} - ${price}")
        
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
        
        result = self.execute_query(query, variables)
        
        if "errors" in result:
            self.print_error(f"Failed to add product: {result['errors'][0]['message']}")
            return None
        
        product = result["data"]["createProduct"]
        self.products.append(product)
        self.print_success(f"Successfully added product: {product['id']} - {product['name']}")
        return product
    
    def register_user(self, name, email, password):
        self.print_info(f"Registering user: {name} - {email}")
        
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
        
        result = self.execute_query(query, variables)
        
        if "errors" in result:
            self.print_error(f"Failed to register user: {result['errors'][0]['message']}")
            return None
        
        token = result["data"]["register"]["token"]
        user = {"name": name, "email": email, "token": token}
        self.users.append(user)
        self.print_success(f"Successfully registered user: {name} with token")
        return user
    
    def create_order(self, user, products):
        self.print_info(f"Creating order for user: {user['name']}")
        
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
        for product, quantity in products:
            order_products.append({
                "id": product["id"],
                "quantity": quantity
            })
        
        variables = {
            "order": {
                "products": order_products
            }
        }
        
        result = self.execute_query(query, variables, user["token"])
        
        if "errors" in result:
            self.print_error(f"Failed to create order: {result['errors'][0]['message']}")
            return None
        
        order = result["data"]["createOrder"]
        self.print_success(f"Successfully created order: {order['id']} with total price: ${order['totalPrice']}")
        
        # Record interactions
        for product, quantity in products:
            self.interactions.append({
                "user": user,
                "product": product,
                "type": "purchase",
                "quantity": quantity
            })
        
        return order
    
    def get_recommendations_by_viewed(self, product_ids, skip=0, take=5):
        product_id_strings = [f'"{pid}"' for pid in product_ids]
        self.print_info(f"Getting recommendations based on viewed products: [{', '.join(product_id_strings)}], skip: {skip}, take: {take}")
        
        query = """
        query GetRecommendations($productIds: [String], $skip: Int!, $take: Int!) {
            product(viewedProductsIds: $productIds, pagination: { skip: $skip, take: $take }) {
                id
                name
                description
                price
            }
        }
        """
        
        variables = {
            "productIds": product_ids,
            "skip": skip,
            "take": take
        }
        
        result = self.execute_query(query, variables)
        
        if "errors" in result:
            self.print_error(f"Failed to get recommendations: {result['errors'][0]['message']}")
            return None
        
        recommendations = result["data"]["product"]
        self.print_success(f"Received {len(recommendations)} recommendations:")
        
        for i, product in enumerate(recommendations):
            print(f"{i+1}. {Fore.CYAN}{product['name']}{Style.RESET_ALL} - ${product['price']} - {product['description']}")
        
        return recommendations
    
    def get_personalized_recommendations(self, user):
        self.print_info(f"Getting personalized recommendations for user: {user['name']}")
        
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
        
        result = self.execute_query(query, token=user["token"])
        
        if "errors" in result:
            self.print_error(f"Failed to get personalized recommendations: {result['errors'][0]['message']}")
            return None
        
        recommendations = result["data"]["product"]
        self.print_success(f"Received {len(recommendations)} personalized recommendations:")
        
        for i, product in enumerate(recommendations):
            print(f"{i+1}. {Fore.CYAN}{product['name']}{Style.RESET_ALL} - ${product['price']} - {product['description']}")
        
        return recommendations
    
    def run_tests(self):
        self.print_header("SETTING UP TEST DATA")
        
        # Add test products
        products = [
            self.add_product("Gaming Laptop", "High-performance gaming laptop with RTX 3080", 1999.99),
            self.add_product("Smartphone", "Latest smartphone with 5G capabilities", 899.99),
            self.add_product("Wireless Headphones", "Noise-cancelling wireless headphones", 249.99),
            self.add_product("Smart Watch", "Fitness tracking smart watch", 199.99),
            self.add_product("Tablet", "10-inch tablet with high-resolution display", 499.99),
            self.add_product("Bluetooth Speaker", "Portable bluetooth speaker with 20h battery life", 129.99),
            self.add_product("Mechanical Keyboard", "RGB mechanical keyboard for gaming", 149.99),
            self.add_product("Wireless Mouse", "Ergonomic wireless mouse", 79.99),
            self.add_product("External SSD", "1TB external SSD with USB-C", 179.99),
            self.add_product("Monitor", "27-inch 4K monitor", 349.99)
        ]
        
        # Filter out None values (failed product creations)
        products = [p for p in products if p]
        
        if not products:
            self.print_error("No products were created successfully. Aborting tests.")
            return
        
        # Register test users
        user1 = self.register_user("John Doe", f"john{random.randint(1000, 9999)}@example.com", "password123")
        user2 = self.register_user("Jane Smith", f"jane{random.randint(1000, 9999)}@example.com", "password456")
        
        if not user1 or not user2:
            self.print_error("Failed to register test users. Aborting tests.")
            return
        
        # Create orders to simulate user interactions
        self.print_header("CREATING USER INTERACTIONS")
        
        # User 1 buys gaming products
        self.create_order(user1, [(products[0], 1)])  # Gaming Laptop
        self.create_order(user1, [(products[6], 1)])  # Mechanical Keyboard
        self.create_order(user1, [(products[7], 1)])  # Wireless Mouse
        
        # User 2 buys mobile products
        self.create_order(user2, [(products[1], 1)])  # Smartphone
        self.create_order(user2, [(products[4], 1)])  # Tablet
        self.create_order(user2, [(products[3], 1)])  # Smart Watch
        
        # Wait a moment for the system to process the interactions
        self.print_info("Waiting for the system to process interactions...")
        time.sleep(2)
        
        # Test recommendations based on viewed products
        self.print_header("TESTING RECOMMENDATIONS BASED ON VIEWED PRODUCTS")
        
        # Test case 1: Recommendations after viewing a gaming laptop
        self.get_recommendations_by_viewed([products[0]["id"]], 0, 3)
        
        # Test case 2: Recommendations after viewing smartphone and tablet
        self.get_recommendations_by_viewed([products[1]["id"], products[4]["id"]], 0, 5)
        
        # Test case 3: Recommendations with pagination
        self.get_recommendations_by_viewed([products[2]["id"]], 2, 3)
        
        # Test personalized recommendations
        self.print_header("TESTING PERSONALIZED RECOMMENDATIONS")
        
        # Test case 4: Personalized recommendations for user 1 (gaming enthusiast)
        self.get_personalized_recommendations(user1)
        
        # Test case 5: Personalized recommendations for user 2 (mobile enthusiast)
        self.get_personalized_recommendations(user2)
        
        self.print_header("TESTING COMPLETE")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Test the recommender system")
    parser.add_argument("--url", help="GraphQL endpoint URL", default="http://localhost:8080/graphql")
    args = parser.parse_args()
    
    GRAPHQL_URL = args.url
    
    tester = RecommenderTester()
    tester.run_tests()
