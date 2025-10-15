"""
Locustfile for load testing BrokerX
"""
from locust import HttpUser, task, between

class BrokerXUser(HttpUser):
    wait_time = between(1, 3)

    def on_start(self):
        """
        Called once when a simulated user starts.
        Logs in and stores the session cookie automatically in self.client.
        """
        login_data = {
            "email": "buyer@email.com",
            "password": "password"
        }
        response = self.client.post("auth/login", data=login_data)
        if response.status_code == 200:
            print("Logged in successfully")
        else:
            print(f"Login failed ({response.status_code}): {response.text}")
    
    @task(1) 
    def orders(self):
        form_data = {
            "symbol": "AAPL",
            "type": "market",
            "action": "buy",
            "quantity": 1,
            "timing": "day"
        }
        with self.client.post("/order/place", data=form_data, catch_response=True) as response:
            try:
                data = response.text
                if response.status_code == 201:
                    if  "order placed sucessfully!" in response:
                        response.success()
                else:
                    response.failure(f"Erreur : {response.status_code} - {data.get('error', 'Unknown error')}")
            except ValueError:
                response.failure(f"Invalid response: {response.text}")