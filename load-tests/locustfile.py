"""
Locustfile for load testing BrokerX
"""
import random
from locust import HttpUser, events, task, between
import numpy as np
import requests

shared_tokens = {
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
        "timing": random.choice(["day"]), # add IOC orders when they are properly handled
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
    login_url = "http://nginx/api/user/auth/login"

    def login_user(email, password):
        try:
            response = requests.post(
                login_url,
                json={"email": email, "password": password},
                timeout=5,
            )
            if response.status_code != 200:
                raise Exception(f"Login failed for {email}: {response.status_code} {response.text}")
            data = response.json()
            return data.get("token")
        except Exception as e:
            print(f"Error logging in {email}: {e}")
            return None

    shared_tokens["buyer"] = login_user("buyer@email.com", "password")
    shared_tokens["seller"] = login_user("seller@email.com", "password")

class BrokerXBuyerUser(HttpUser):
    wait_time = between(1, 2)

    def on_start(self):
        if not shared_tokens["buyer"]:
            raise RuntimeError("No buyer token available; login failed")
        self.jwt = shared_tokens["buyer"]
        self.headers = {
            "Authorization": f"Bearer {self.jwt}",
            "Content-Type": "application/json"
        }

    @task(1)
    def place_buy_order(self):
        order = make_random_order("buy")
        try:
            with self.client.post(
                "/api/order/place",
                json=order,
                headers=self.headers,
                catch_response=True
            ) as response:
                if response.status_code == 201:
                    response.success()
                else:
                    response.failure(f"Buy order failed : {response.text}")
        except requests.exceptions.ReadTimeout:
            events.request.fire(
                request_type="POST",
                name="/api/order/place",
                exception=requests.ReadTimeout("Request timed out"),
            )

class BrokerXSellerUser(HttpUser):
    wait_time = between(1, 2)

    def on_start(self):
        if not shared_tokens["seller"]:
            raise RuntimeError("No seller token available; login failed")
        self.jwt = shared_tokens["seller"]
        self.headers = {
            "Authorization": f"Bearer {self.jwt}",
            "Content-Type": "application/json"
        }

    @task(1)
    def place_sell_order(self):
        order = make_random_order("sell")
        with self.client.post(
            "/api/order/place",
            json=order,
            headers=self.headers,
            catch_response=True
        ) as response:
            if response.status_code == 201:
                response.success()
            else:
                response.failure(f"Sell order failed : {response.text}")