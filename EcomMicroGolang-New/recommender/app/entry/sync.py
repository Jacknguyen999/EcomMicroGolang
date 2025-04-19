from kafka import KafkaConsumer
import json
import requests
import time
import logging
import sys
from app.db.session import ReplicaSession
from app.db.models import Product, Interaction
from config.settings import PRODUCT_API, KAFKA_SERVER

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

def sync_products():
    while True:
        try:
            logger.info("Attempting to connect to Kafka for product events...")
            consumer = KafkaConsumer(
                "product_events",
                bootstrap_servers=KAFKA_SERVER,
                auto_offset_reset='earliest',
                enable_auto_commit=True,
                group_id='recommender-product-group',
                value_deserializer=lambda x: json.loads(x.decode('utf-8'))
            )
            logger.info("Successfully connected to Kafka for product events")

            for message in consumer:
                try:
                    event = message.value
                    logger.info(f"Received product event: {event['type']}")

                    with ReplicaSession() as session:
                        if event["type"] in ["product_created", "product_updated"]:
                            product_data = event["data"]
                            logger.info(f"Processing product event: {event['type']} for product ID: {product_data['product_id']}")

                            product = session.query(Product).filter_by(id=product_data["product_id"]).first()
                            if product:
                                product.name = product_data["name"]
                                product.description = product_data["description"]
                                product.price = product_data["price"]
                                product.account_id = product_data["account_id"]
                            else:
                                # Extract account ID with fallbacks for different field names
                                account_id = product_data.get("accountID",
                                                product_data.get("account_id",
                                                    product_data.get("accountId", 0)))

                                logger.info(f"Creating product with account_id: {account_id}, data: {product_data}")

                                product = Product(
                                    id=product_data["product_id"],
                                    name=product_data["name"],
                                    description=product_data["description"],
                                    price=product_data["price"],
                                    account_id=account_id
                                )
                                session.add(product)
                            session.commit()
                            logger.info(f"Successfully processed product event: {event['type']}")

                        elif event["type"] == "product_deleted":
                            product_id = event["data"]["product_id"]
                            logger.info(f"Processing product deletion for ID: {product_id}")

                            product = session.query(Product).filter_by(id=product_id).first()
                            if product:
                                session.delete(product)
                                session.commit()
                                logger.info(f"Successfully deleted product ID: {product_id}")
                            else:
                                logger.warning(f"Product ID {product_id} not found for deletion")
                except Exception as e:
                    logger.error(f"Error processing product event: {str(e)}")

        except Exception as e:
            logger.error(f"Kafka connection error for product events: {str(e)}")
            logger.info("Retrying connection in 5 seconds...")
            time.sleep(5)

def process_interactions():
    while True:
        try:
            logger.info("Attempting to connect to Kafka for interaction events...")
            consumer = KafkaConsumer(
                "interaction_events",
                bootstrap_servers=KAFKA_SERVER,
                auto_offset_reset='earliest',
                enable_auto_commit=True,
                group_id='recommender-interaction-group',
                value_deserializer=lambda x: json.loads(x.decode('utf-8'))
            )
            logger.info("Successfully connected to Kafka for interaction events")

            for message in consumer:
                try:
                    event = message.value
                    logger.info(f"Received interaction event: {event['type']}")

                    with ReplicaSession() as session:
                        try:
                            interaction = Interaction(
                                user_id=event["data"]["user_id"],
                                product_id=event["data"]["product_id"],
                                interaction_type=event["type"]
                            )
                            session.add(interaction)

                            # Check if product exists in our database
                            product_id = event["data"]["product_id"]
                            product = session.query(Product).filter_by(id=product_id).first()

                            if not product:
                                logger.info(f"Product {product_id} not found in database, fetching from API")
                                try:
                                    response = requests.get(f"{PRODUCT_API}/{product_id}", timeout=5)
                                    response.raise_for_status()
                                    product_data = response.json()
                                    product = Product(**product_data)
                                    session.add(product)
                                    logger.info(f"Successfully fetched and added product {product_id}")
                                except requests.RequestException as e:
                                    logger.error(f"Failed to fetch product {product_id}: {str(e)}")

                            session.commit()
                            logger.info(f"Successfully processed interaction event for product {product_id}")
                        except Exception as e:
                            logger.error(f"Error processing interaction data: {str(e)}")
                            session.rollback()
                except Exception as e:
                    logger.error(f"Error processing interaction event: {str(e)}")

        except Exception as e:
            logger.error(f"Kafka connection error for interaction events: {str(e)}")
            logger.info("Retrying connection in 5 seconds...")
            time.sleep(5)

import threading

def run_in_thread(target):
    thread = threading.Thread(target=target)
    thread.daemon = True
    thread.start()
    return thread

if __name__ == "__main__":
    logger.info("Starting recommender sync service")

    # Start both processes in separate threads
    product_thread = run_in_thread(sync_products)
    interaction_thread = run_in_thread(process_interactions)

    # Keep the main thread alive
    try:
        while True:
            time.sleep(60)
            logger.info("Recommender sync service is running")
    except KeyboardInterrupt:
        logger.info("Shutting down recommender sync service")
    except Exception as e:
        logger.error(f"Error in main thread: {str(e)}")