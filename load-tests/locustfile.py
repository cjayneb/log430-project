"""
Locustfile for load testing BrokerX
"""
import random
from locust import HttpUser, events, task, between
import numpy as np
import requests

shared_sessions = {
    "buyer": None,
    "seller": None,
}

# ------------------------------------------------------
# GLOBAL SETUP — Runs once before any users are spawned
# ------------------------------------------------------
@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    """
    Logs in as buyer and seller once before users are spawned,
    and stores their session cookies globally.
    """
    base_url = "http://brokerx:8080"

    # Buyer login
    session = requests.Session()
    session.post(
        f"{base_url}/auth/login",
        data={"email": "buyer@email.com", "password": "password"}
    )
    shared_sessions["buyer"] = session.cookies

    # Seller login
    session = requests.Session()
    session.post(
        f"{base_url}/auth/login",
        data={"email": "seller@email.com", "password": "password"}
    )
    shared_sessions["seller"] = session.cookies


def get_buy_order():
    order_type = random.choice(["market", "limit"])
    unit_price = 0.0 if order_type == "market" else random.choice(np.arange(165.0, 185.0, 0.01))
    return {
        "symbol": "AAPL",
        "type": order_type,
        "action": "buy",
        "quantity": random.choice(range(1,10)),
        "timing": "day",
        "unit_price": unit_price
    }

def get_sell_order():
    order_type = random.choice(["market", "limit"])
    unit_price = 0.0 if order_type == "market" else random.choice(np.arange(165.0, 185.0, 0.01))
    return {
        "symbol": "AAPL",
        "type": order_type,
        "action": "sell",
        "quantity": random.choice(range(1,10)),
        "timing": "day",
        "unit_price": unit_price
    }

class BrokerXBuyerUser(HttpUser):
    wait_time = between(0.02, 0.04)

    def on_start(self):
        if shared_sessions["buyer"]:
            self.client.cookies = shared_sessions["buyer"]
        else:
            print("No buyer cookies available")
    
    @task(1) 
    def orders(self):
        form_data = get_buy_order()
        with self.client.post("/order/place", data=form_data, catch_response=True) as response:
            try:
                data = response.text
                if response.status_code == 201:
                    if  "order placed sucessfully!" in response:
                        response.success()
                else:
                    response.failure(f"Erreur : {response.status_code} - {data.split('h3')[1]}")
            except ValueError:
                response.failure(f"Invalid response: {response.text}")

class BrokerXSellerUser(HttpUser):
    wait_time = between(0.02, 0.04)

    def on_start(self):
        if shared_sessions["seller"]:
            self.client.cookies = shared_sessions["seller"]
        else:
            print("No seller cookies available")
    
    @task(1) 
    def orders(self):
        form_data = get_sell_order()
        with self.client.post("/order/place", data=form_data, catch_response=True) as response:
            try:
                data = response.text
                if response.status_code == 201:
                    if  "order placed sucessfully!" in response:
                        response.success()
                else:
                    response.failure(f"Erreur : {response.status_code} - {data.split('h3')[1]}")
            except ValueError:
                response.failure(f"Invalid response: {response.text}")