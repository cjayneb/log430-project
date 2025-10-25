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

def make_random_order(action):
    order_type = random.choice(["market", "limit"])
    unit_price = 0.0 if order_type == "market" else random.choice(np.arange(165.0, 185.0, 0.01))
    return {
        "symbol": "AAPL",
        "type": order_type,
        "action": action,
        "quantity": random.choice(range(1,10)),
        "timing": random.choice(["day", "ioc"]),
        "unit_price": unit_price
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
    base_url = "http://nginx:80"

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

class BrokerXBuyerUser(HttpUser):
    wait_time = between(1, 2)

    def on_start(self):
        if shared_sessions["buyer"]:
            self.client.cookies = shared_sessions["buyer"]
        else:
            print("No buyer cookies available")
    
    @task(1) 
    def orders(self):
        with self.client.post("/order/place", data=make_random_order("buy"), catch_response=True) as response:
            try:
                data = response.text
                if response.status_code == 201:
                    if  "order placed" in response:
                        response.success()
                else:
                    response.failure(f"Erreur : {response.status_code} - {data.split('h3')[1]}")
            except Exception:
                response.failure(f"Invalid response: {response.text}")

class BrokerXSellerUser(HttpUser):
    wait_time = between(1, 2)

    def on_start(self):
        if shared_sessions["seller"]:
            self.client.cookies = shared_sessions["seller"]
        else:
            print("No seller cookies available")
    
    @task(1) 
    def orders(self):
        with self.client.post("/order/place", data=make_random_order("sell"), catch_response=True) as response:
            try:
                data = response.text
                if response.status_code == 201:
                    if  "order placed" in response:
                        response.success()
                else:
                    response.failure(f"Erreur : {response.status_code} - {data.split('h3')[1]}")
            except Exception:
                response.failure(f"Invalid response: {response.text}")