"""
Locustfile for load testing BrokerX
"""
import random
from locust import HttpUser, task, between

class FlaskAPIUser(HttpUser):
    # Temps d'attente entre les requêtes (1 à 3 secondes)
    wait_time = between(1, 3)
    
    # Proportion d'exécution 1:1:1, ce qui signifie : 1/3, 1/3, 1/3 (33 % des appels à l'écriture et 66 % des appels à la lecture)
    @task(1) 
    def orders(self):
        """Test POST /orders endpoint (write)"""
        mock_order = {
            "user_id": random.choice([1, 2, 3]),
            "items": [{"product_id": random.choice([1, 2, 3, 4]), "quantity": random.choice([1, 2])}] 
        }   

        # Ajouter aléatoirement plusiers articles (30 % des fois)
        if random.randint(1, 10) <= 3:
            mock_order["items"].append({"product_id": random.choice([1, 2, 3, 4]), "quantity": random.choice([1, 2])})

        with self.client.post("/orders", 
                            json=mock_order, 
                            headers={"Content-Type": "application/json"},
                            catch_response=True) as response:
            try:
                data = response.json()
                if response.status_code == 201:
                    if "order_id" in data:
                        response.success()
                    else:
                        response.failure("Aucun ID renvoyé pour la commande créée")
                else:
                    response.failure(f"Erreur : {response.status_code} - {data.get('error', 'Unknown error')}")
            except ValueError:
                response.failure(f"Invalid JSON response: {response.text}")

    @task(1) 
    def highest_spenders(self):
        """Test GET /orders/reports/highest-spenders endpoint (read)"""
        with self.client.get("/orders/reports/highest-spenders", catch_response=True) as response:
            try:
                data = response.json()
                if response.status_code == 200: 
                    if str(type(data)) == "<class 'list'>":
                        response.success()
                    else:
                        response.failure("Le resultat n'est pas une liste : " + str(data))
                else:
                    response.failure(f"Erreur : {response.status_code} - {data.get('error', 'Unknown error')}")
            except ValueError:
                response.failure(f"Invalid JSON response: {response.text}")

    @task(1) 
    def best_sellers(self):
        """Test GET /orders/reports/best-sellers endpoint (read)"""
        with self.client.get("/orders/reports/best-sellers", catch_response=True) as response:
            try:
                data = response.json()
                if response.status_code == 200: 
                    if str(type(data)) == "<class 'list'>":
                        response.success()
                    else:
                        response.failure("Le resultat n'est pas une liste : " + str(data))
                else:
                    response.failure(f"Erreur : {response.status_code} - {data.get('error', 'Unknown error')}")
            except ValueError:
                response.failure(f"Invalid JSON response: {response.text}")