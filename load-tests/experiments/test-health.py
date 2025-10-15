"""
Locustfile for load testing BrokerX
"""
import random
from locust import HttpUser, task, between

class FlaskAPIUser(HttpUser):
    wait_time = between(1, 3)
    
    @task(1) 
    def orders(self):
        with self.client.get("/health", catch_response=True) as response:
            try:
                data = response.json()
                if response.status_code == 200:
                    if data["message"] == "OK":
                        response.success()
                    else:
                        response.failure("Wrong response from health endpoint!")
                else:
                    response.failure(f"Erreur : {response.status_code} - {data.get('error', 'Unknown error')}")
            except ValueError:
                response.failure(f"Invalid response: {response.text}")