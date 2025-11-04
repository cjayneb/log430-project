<img src="https://upload.wikimedia.org/wikipedia/commons/2/2a/Ets_quebec_logo.png" width="250"> \
Jean-Christophe Benoit \
Rapport de projet Phase 2 \
LOG430 — Architecture logicielle \
28 octobre 2025, Montréal \
École de technologie supérieure

<!-- TOC start (generated with https://github.com/derlin/bitdowntoc) -->

- [Run book](#run-book)
   * [Prerequisites](#prerequisites)
   * [Installation](#installation)
   * [Running the project](#running-the-project)
      + [Run locally (without Docker)](#run-locally-without-docker)
      + [Run with Docker Compose](#run-with-docker-compose)
   * [Running tests](#running-tests)
      + [Running all tests (with coverage report)](#running-all-tests-with-coverage-report)
      + [Generate HTML coverage report](#generate-html-coverage-report)
   * [Deployment](#deployment)
      + [Deploying locally](#deploying-locally)
      + [Deploying remotely](#deploying-remotely)
- [User Guide](#user-guide)
   * [Access BrokerX](#access-brokerx)
   * [Registering as a new user](#registering-as-a-new-user)
   * [Logging in](#logging-in)
   * [Placing an order](#placing-an-order)
   * [Adding funds to your wallet](#adding-funds-to-your-wallet)
- [Performance tests results](#performance-tests-results)
   * [Monolithic BrokerX from phase 1](#monolithic-brokerx-from-phase-1)
      + [No caching, no load balancing](#no-caching-no-load-balancing)
         - [100 users with peak ~65 orders placed/second](#100-users-with-peak-65-orders-placedsecond)
         - [200 users with peak ~105 orders placed/second](#200-users-with-peak-105-orders-placedsecond)
         - [100 users with peak ~65 orders placed/second](#100-users-with-peak-65-orders-placedsecond-1)
      + [Added Redis Order book (caching), no load balancing](#added-redis-order-book-caching-no-load-balancing)
         - [Load test up to 400 RPS](#load-test-up-to-400-rps)
         - [Performance test result after implementing regular dirty order syncing (7 minutes, 500 users, 5 users/second)](#performance-test-result-after-implementing-regular-dirty-order-syncing-7-minutes-500-users-5-userssecond)
         - [Performance test result after removing unused composite indexes (7 minutes, 500 users, 5 users/second)](#performance-test-result-after-removing-unused-composite-indexes-7-minutes-500-users-5-userssecond)
         - [Performance test result after implementing Redis order persistence queue and execution record persistence queue (7 minutes, 500 users, 5 users/second) :](#performance-test-result-after-implementing-redis-order-persistence-queue-and-execution-record-persistence-queue-7-minutes-500-users-5-userssecond-)
      + [Added load balancing](#added-load-balancing)
         - [Performance test result 2 instances (7 minutes, 500 users, 5 users/second)](#performance-test-result-2-instances-7-minutes-500-users-5-userssecond)
         - [Performance test result 2 instances (10 minutes, 1000 users, 5 users/second)](#performance-test-result-2-instances-10-minutes-1000-users-5-userssecond)
         - [Performance test result 2 instances (10 minutes, 1600 users, 5 users/second)](#performance-test-result-2-instances-10-minutes-1600-users-5-userssecond)
   * [Micro-services architecture BrokerX](#micro-services-architecture-brokerx)
      + [No load balancing](#no-load-balancing)
         - [Run #1 (300 RPS)](#run-1-300-rps)
         - [Run #2 (300 RPS)](#run-2-300-rps)
         - [Run #3 (300 RPS)](#run-3-300-rps)
         - [Run #4 (up to 600 RPS)](#run-4-up-to-600-rps)
         - [Run #5 (up to 900 RPS)](#run-5-up-to-900-rps)
      + [With load balancing](#with-load-balancing)
   * [Load Tests Conclusions](#load-tests-conclusions)
- [Arc42](#arc42)
- [1. Introduction and Goals](#1-introduction-and-goals)
   * [1.1 Requirements Overview](#11-requirements-overview)
      + [The system **must** support the following core functionalities:](#the-system-must-support-the-following-core-functionalities)
      + [The system **should** support the following functionalities:](#the-system-should-support-the-following-functionalities)
      + [The system **could** support the following functionalities:](#the-system-could-support-the-following-functionalities)
   * [1.2 Quality Goals](#12-quality-goals)
   * [1.3 Stakeholders](#13-stakeholders)
- [2. Architecture Constraints](#2-architecture-constraints)
- [3. System Scope and Context](#3-system-scope-and-context)
   * [Business Context](#business-context)
   * [Technical Context](#technical-context)
- [4. Solution Strategy](#4-solution-strategy)
- [5. Building Block View](#5-building-block-view)
- [6. Runtime View](#6-runtime-view)
   * [General Use Case Diagram](#general-use-case-diagram)
   * [UC-01 Sign up](#uc-01-sign-up)
      + [Goal:](#goal)
      + [Actors:](#actors)
      + [Preconditions:](#preconditions)
      + [Main Flow:](#main-flow)
      + [Postconditions:](#postconditions)
      + [Exceptions / Alternatives:](#exceptions-alternatives)
   * [UC-02 Login](#uc-02-login)
      + [Goal:](#goal-1)
      + [Actors:](#actors-1)
      + [Preconditions:](#preconditions-1)
      + [Main Flow:](#main-flow-1)
      + [Postconditions:](#postconditions-1)
      + [Exceptions / Alternatives:](#exceptions-alternatives-1)
   * [UC-05 Place order](#uc-05-place-order)
      + [Goal:](#goal-2)
      + [Actors:](#actors-2)
      + [Preconditions:](#preconditions-2)
      + [Main Flow:](#main-flow-2)
      + [Postconditions:](#postconditions-2)
      + [Exceptions / Alternatives:](#exceptions-alternatives-2)
   * [UC-07 Find match and execute order](#uc-07-find-match-and-execute-order)
      + [Actors:](#actors-3)
      + [Preconditions:](#preconditions-3)
      + [Main Flow:](#main-flow-3)
      + [Postconditions:](#postconditions-3)
      + [Exceptions / Alternatives:](#exceptions-alternatives-3)
- [7. Deployment View](#7-deployment-view)
- [8. Cross-cutting Concepts](#8-cross-cutting-concepts)
- [9. Design Decisions](#9-design-decisions)
   * [ADR-01: Hexagonal architecture](#adr-01-hexagonal-architecture)
      + [Context](#context)
      + [Decision](#decision)
      + [Status](#status)
      + [Consequences](#consequences)
   * [ADR-02: Persistence with MySQL and Repository/Transaction Manager Pattern](#adr-02-persistence-with-mysql-and-repositorytransaction-manager-pattern)
      + [Context](#context-1)
      + [Decision](#decision-1)
      + [Status](#status-1)
      + [Consequences](#consequences-1)
   * [ADR-03: Use of Go as Implementation Language](#adr-03-use-of-go-as-implementation-language)
      + [Context](#context-2)
      + [Decision](#decision-2)
      + [Status](#status-2)
      + [Consequences](#consequences-2)
   * [ADR-04: NGINX API Gateway](#adr-04-nginx-api-gateway)
      + [Context](#context-3)
      + [Decision](#decision-3)
      + [Status](#status-3)
      + [Consequences](#consequences-3)
   * [ADR-05: Redi-Based Order Book](#adr-05-redi-based-order-book)
      + [Context](#context-4)
      + [Decision](#decision-4)
      + [Status](#status-4)
      + [Consequences](#consequences-4)
- [10. Quality Requirements](#10-quality-requirements)
   * [Latency](#latency)
   * [Throughput](#throughput)
   * [Data Integrity](#data-integrity)
   * [Security](#security)
   * [Testability](#testability)
   * [Interoperability](#interoperability)
- [11. Risks and Technical Debts](#11-risks-and-technical-debts)
- [12. Glossary](#12-glossary)

<!-- TOC end -->

# Run book
## Prerequisites

To run this project locally, you need the following tools installed:

- [Go 1.21+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/)

## Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/cjayneb/log430-project.git
   cd log430-project
   ```

2. **Configure environment variables**

   Change the values in the `backend/services/{service-name}/config.go` files if you want to override defaults.
   Example:

   ```env
    Port  string `env:"APP_PORT" envDefault:"8181"`
    DBUrl string `env:"DATABASE_URL" envDefault:"user:pass@tcp(127.0.0.1:3306)/brokerx"`
   ```

   > The variables in `config.go` are used when running the app locally. The environnement variables set in the `docker-compose.yml` will override the values set in `backend/services/{service-name}/config.go` when running the project with Docker Compose.

## Running the project

### Run locally (without Docker)

From on of the `backend/services/{service-name}` directories:

```bash
go run .
```

This will start the microservice's API on http://127.0.0.1:8080.

- Health endpoint: http://127.0.0.1:8080/health (GET) (each microservice has one)

More examples
- Login endpoint: http://127.0.0.1:8080/api/user/auth/login (POST)
- Home page endpoint: http://127.0.0.1:8080/ (GET)
- Orders page endpoint: http://127.0.0.1:8080/api/order (GET)

> You must have a MySQL instance and a Redis instance running on your machine for all uses cases to work

### Run with Docker Compose

From the project root:

```bash
docker compose down -v # To remove existing containers and their volumes
docker compose up --build -d
```

This starts:

- The Go backend microservices on [http://localhost/api](http://localhost/api)
- A MySQL database (`brokerx_db`) on port `3306`
- Nginx API Gateway
- Redis DB

## Running tests

**Important** : Before running all tests, you must have the test database up, because some of the tests are integration tests needing a real database.

Run the following command to start the test database:

```bash
#From the root of the project
docker compose -f docker-compose.test.yml up -d
```

### Running all tests (with coverage report)

Inside the `backend/services/{service-name}` folder of each micro-service:

```bash
go test ./... -coverprofile=coverage
```

### Generate HTML coverage report

```bash
go tool cover -html=coverage
```

## Deployment

At this stage, the application is deployed locally or remotely using Docker Compose.
A production-ready deployment would likely use Kubernetes or cloud-based services, but that is outside the current scope.

### Deploying locally

To deploy locally, you just have to run the following commands

```bash
docker compose down -v # Ensure Docker is clean with no previous deployment
docker compose up --build -d # Build the Docker image and run docker compose in detached mode
```

### Deploying remotely

The GitHub Actions Workflow should take care of deploying the application to the ETS Virtual Machine self hosted runner automatically on every push. See `.github/workflows/ci_cd.yml`

_The current deployment pipeline is broken because of issues with the storage on the ETS VM._

> To access the remote deployment, you must be connected to the ETS Cisco Secure Client via accesvpn.etsmtl.ca


# User Guide
This document show syou how to use BrokerX once it has been deployed either locally or remotely.

## Access BrokerX

To access BrokerX, use your web browser and navigate to :

- Locally : http://localhost/
- Remotely : http://10.194.32.206/

## Registering as a new user

Access the following link in your browser : http://localhost/register.html
![alt text](image-8.png)

Enter you first and last name as well as an email and a password and submit : 
![alt text](image-9.png)

## Logging in

Enter the email and password and click on _Login_ to access the account of a seller :

- Email : seller@email.com
- Password : password

![alt text](image.png)

## Placing an order

Logging in will bring you to this _Orders_ page:

![alt text](image-1.png)

Then you can fill the form to sell AAPL stocks at market price and click _Submit Order_ :
![alt text](image-2.png)

Then you should see your order appear at the top of the page and this confirmation message at the bottom of the form.

> Error messages will also appear at the bottom of the form.

![alt text](image-3.png)

Next, if you refresh the page, you should see that your order status, and remaining quantity have been updated if a matching order has been found :
![alt text](image-4.png)

> Other users (email and buyer@email.com) have different base positions, orders and wallet balances, which make for different outputs in results.

## Adding funds to your wallet

After logging in, it is possible to add funds to your wallet by clicking on _Wallet_:
![alt text](image-5.png)

Then entering the amount you want to deposit and submitting : 
![alt text](image-6.png)

The balance automatically updates itself and you will be able to make bigger purchases of stocks : 
![alt text](image-7.png)

# Performance tests results
During the second phase of the Brokerx project, a myriad of load tests were done to assess the throughput, latency and availability of the app depending on the architecture and configuration. A Locust container was used to send send requests to the BrokerX endpoint responsible for processing orders.

Here are the test results in chronological order with a description of the improvements made at each iteration.

## Monolithic BrokerX from phase 1

### No caching, no load balancing
#### 100 users with peak ~65 orders placed/second

![alt text](lt1.png)

Stays usable to whole time, but a slight increase in latency is observed over time.

#### 200 users with peak ~105 orders placed/second

![alt text](lt2.png)

Quickly becomes unusable after reaching 100 RPS due to latency surpassing 500 ms and degrading over time.

> The performance tests above revealed the following issues which will be fixed in the next iterations :
>- One goroutine created per request to submit to matching engine (adressed in Introduce caching #5)
>- Requests to fetch orders every time an order is placed (adressed in Introduce caching #5)
>- Heavy load on CPU caused by html template rendering for every single request (adressed in Transform into RESTful API #1)
>- Lack of Db connection pool management (connections not closing, not being reused) (adressed in Introduce caching #5)
>- Heavy logging (adressed in Structure logs #7)


#### 100 users with peak ~65 orders placed/second

The main culprit was found to be the html rendering for every single request. After removing html rendering, the cpu load is constant with constant requests.
Image

![alt text](lt3.png)

However, the cpu load on the mysql database container is steadily increasing with constant requests. This suggests that order fetching becomes more and more demanding, the more our tables grow. Investigation is underway

> With the current schema creation script, there is no index on the fields used by the matching engine to fetch, making fetching time of orders grow linearly with growing orders table.

> Here are the solutions that encourage a logarithmic increase in database cpu usage (load stabilizes with growing tables):
> - Added two indexes on the orders table
CREATE INDEX idx_orders_match_asc ON orders(symbol, action, status, unit_price ASC, created_at ASC);
CREATE INDEX idx_orders_match_desc ON orders(symbol, action, status, unit_price DESC, created_at ASC);
(To be implemented) 
> - Add Redis queue that created orders are sent to and make matching engine read from that queue instead of always reading from database, therefore reducing load on mysql database container

With the two indexes added, we can observe a logarithmic increase of cpu load on the db between 10:59 and 11:06 when the Request load was constant. At 11:06, the RPS was doubled, which explains the sudden rise :

![alt text](lt4.png)

Implemented batch writes for execution records. Performance is stable over time, but db usage still keeps rising the more orders are stored in the orders table.

![alt text](lt5.png)

![alt text](lt6.png)

![alt text](lt7.png)

### Added Redis Order book (caching), no load balancing
The problem observed in the previous tests was that the mysql database or any relationnal database will be the bottleneck of this application if we continue fetching and creating rows continuously in a growing orders table. It is inevitable. Therefore, the next logical step is using Redis to keep a live in memory order book :

- Redis Sorted Set = “live order book” (fast, concurrent, transient state)
- MySQL = “persistent ledger” (durable, slower, authoritative snapshot)

Adding redis to take care of the live order book really helped to stabilize the cpu load on the mysql container since most operations are done using the redis ordered set which is a lot more performant than constant sql queries to fetch and set data.

#### Load test up to 400 RPS
Here are the results of the first performance test after these redis changes :

![alt text](lt8.png)

![alt text](lt9.png)

![alt text](lt10.png)

![alt text](lt11.png)

With the load increasing and reaching almost 500 RPS on only 1 instance! It is normal to see the cpu load increasing since the load was also increasing every few minutes. But what is important is that the latency stayed way under the required 500 ms and the cpu load somewhat stabilized after reaching each plateau of request load.

Next steps are :

- Implement regular order syncing from redis to mysql so that incomplete orders are visible to the user
- OPtimize sql performance by changing order and execution persistance by sending to queue and flushing every x seconds
- Remove unecessary composite sql indexes
- Optimize locust performance so that it takes less cpu % to eventually reach 1000 RPS
- Try load balancing go api to reduce stress on each instance
- Try distributed sql database? Or other database (see class slides)

#### Performance test result after implementing regular dirty order syncing (7 minutes, 500 users, 5 users/second)

![alt text](lt12.png)

![alt text](lt13.png)

![alt text](lt14.png)

![alt text](lt15.png)

#### Performance test result after removing unused composite indexes (7 minutes, 500 users, 5 users/second)

![alt text](lt16.png)

![alt text](lt17.png)

![alt text](lt18.png)

![alt text](lt19.png)

It seems db cpu usage decreased a little bit. Overall performance seems more stable.

#### Performance test result after implementing Redis order persistence queue and execution record persistence queue (7 minutes, 500 users, 5 users/second) :

![alt text](lt20.png)

![alt text](lt21.png)

![alt text](lt22.png)

![alt text](lt23.png)

The goal of these improvements was to reduce the over reliance on the sql database and I think this goal was achieved. As we can see, with a stable load of requests (~300 RPS), the cpu load of each key component seems to stabilize (even though the number of stored orders is constantly increasing) and especially the database cpu load has been reduced significantly with plenty of room for higher load testing.

### Added load balancing

#### Performance test result 2 instances (7 minutes, 500 users, 5 users/second)

![alt text](lt24.png)

There seems to be no need for more than two instances at this level of request load. All metrics are very stable with each brokerx instance using around 10% cpu.

#### Performance test result 2 instances (10 minutes, 1000 users, 5 users/second)

![alt text](lt25.png)

VERY acceptable latency, cpu loads are very stable.

#### Performance test result 2 instances (10 minutes, 1600 users, 5 users/second)

![alt text](lt26.png)

With a request load of 1000+ RPS, the app is still comfortably responding within 250 ms with an average latency of 5 ms! The closest bottleneck still seems to be the sql database, meaning that future improvements should focus on increasing batch sizes when persisting to database and less reliance on database for order processing and matching. 

We settle at 2 instances of the go api with the current monolith setup.

## Micro-services architecture BrokerX

We expect that transitionning to micro services should help distribute the load between the various services and therefore improve performance.

However, micro-services bring about a new challenges such as http client connection pooling and increased latency because of calls between the services.

Below are the results of the few performance tests that were performed after the transformation of BrokerX into RESTful API micro-services with an NGINX API Gateway.

### No load balancing

#### Run #1 (300 RPS)
First trying out performance tests on new microservice api gateway architecture, and the results are not great when reaching over 300 RPS. The order service becomes overwhelmed because the requests to market-service and portfolio-service dont work anymore due to nginx 512 worker_connections are not enough

![alt text](lt27.png)

#### Run #2 (300 RPS)
First improvement : increase nginx worker connections

![alt text](lt28.png)

Less errors but the delay is unacceptable

#### Run #3 (300 RPS)
Second improvement : allow each service to keep up to 100 idle tcp connections

![alt text](lt29.png)

Improvement but failures and high latency start to appear again once reaching 400 RPS. The culprits are the following in the order service :
- level=error msg="matching failed for order #132124: error making request to matching service: Post "http://nginx/api/matching/\": read tcp 172.18.0.7:38756->172.18.0.10:80: read: connection reset by peer"
- level=error msg="matching failed for order #133737: error making request to matching service: Post "http://nginx/api/matching/\": context canceled"
- level=error msg="matching failed for order #132440: error making request to matching service: Post "http://nginx/api/matching/\": http: server closed idle connection"

#### Run #4 (up to 600 RPS)

Third improvement : Add shared http client for order service

![alt text](lt30.png)

System gets unstable quickly after 600 RPS.

#### Run #5 (up to 900 RPS)
Fourth improvement : Make calls from one microservice to another internal calls that dont go through nginx.

![alt text](lt31.png)

System gets abit unstable when reaching around 800 RPS. Surprisingly high cpu usage on order service and matching service. They are the main service doing work but monolith brokerx never reached this high of a cpu load.

### With load balancing
> Load balancing was not tested on the microservices architecture due to the lack of CPU resources available.

## Load Tests Conclusions

The load tests showed how BrokerX's architecture and code evolved to achieve the desired results. The initial monolithic version from phase 1 had a clear overreliance on the database and took care of too much static html rendering.

Progressive improvements were implemented, such as Redis caching and batch writes to the database, and this is where most of the performance was improved. 

When transitionning to microservices, new problems arose, but were fixed to allow for 800+ RPS an average latency under 250 ms and an availability over 99%.


# Arc42
# 1. Introduction and Goals

## 1.1 Requirements Overview

The BrokerX project is an online brokerage platform targeting individual investors, developed in response to the rapidly evolving financial services sector. It is made to allow retail investors to trade stocks securely and rapidly from the comfort of their home.

### The system **must** support the following core functionalities:

- **Authentication** : Secure login with session cookies.
  - Justification: Secure access to the system is mandatory, especially considering the sensitivity of the data residing on BrokerX.
- **Order placement** : Allow users to place market or limit orders on various stocks.
  - Justification: This is a core business function. Without order placement, Brokerx has no reason to exist.
- **Matching & Execution** : Ensure matching of buy and sell orders using an internal matching engine, generating an execution report.
  - Justification: This is another feature that is too important to be ignored, placing an order to buy stocks is a must, but it is not useful if the order cannot be matched with a corresponding selling order.

### The system **should** support the following functionalities:

- **Market data subscription** : Realtime market data updates (top-of-book, trades, OHLC).
  - Justififcation: This feature should definitely be a part of BrokerX, so that users do not have to use another platform to consult stock prices, but it is not a necessity to have a brokerage system that works.
- **Registration** : Account creation and identity verification via an OTP sent by email.
  - Justification: For now, user accounts can be manually added to the system and tests of the whole system can still be done without a registration page.
- **Wallet funding** : Allow users to add money to their BrokerX wallet in order to buy stocks and get compensated for selling stocks.
  - Justification: For now, wallets can be also be manually created and given any amount as a test balance, since the system does not yet interact with the real stock market.
- **Order modification/cancellation** : Allow users to modify or cancel order that have been placed.
  - Justification: This is not a necessity as it does not pose any real risk to lack the ability to modify or cancel orders during this phase. However, a brokerage platform should definitely have this feature in order to rectify input errors.

### The system **could** support the following functionalities:

- **Notifications** : Send email notifications to the user when and order is confirmed.
  - Justification: This is a nice-to-have and in no way necessary to a working brokerage platform.a

> These functions cover the core trading workflow and will be expanded in later iterations.

[LOG430 - 2025.3 - Projet - Cahier de Charge.pdf](Other%20docs/cahier_de_charge.pdf)

## 1.2 Quality Goals

| Priority | Quality goal | Scenario                                                   |
| -------- | ------------ | ---------------------------------------------------------- |
| 1        | Latency      | ≤ 250 ms to get an acknowledgement after placing an order. |
| 2        | Throughput   | ≥ 800 orders successfully placed per second.               |
| 3        | Availability | The system must be available at least 95.5% of the time.     |

These goals are to be met during the second iteration (micro-services architecture) of BrokerX.

**Notes / roadmap:** The project specification defines phased targets (monolith → microservices → event-driven) where the latency/throughput and availability targets tighten in later phases. Observability (logs + metrics) is required from phase 2 and distributed tracing from phase 3. These constraints drive the choice of architecture and incremental migration plan

## 1.3 Stakeholders

**Contents.**

| Role/Name                     | Contact                | Expectations                                                                                               |
| ----------------------------- | ---------------------- | ---------------------------------------------------------------------------------------------------------- |
| Clients (web users)           | _N/A_                  | Fast, correct order handling; clear execution confirmations; accurate portfolio balances.                  |
| Developers of BrokerX         | Jean-Christophe Benoit | Clear modular code structure (domain vs infra), testability, reproducible CI/CD, containerized deployment. |
| Back-Office Operations        | _N/A_                  | Accurate execution records, trustworthy reports for accounting.                                            |
| Compliance /Risk Employees    | _N/A_                  | Control over pre-trade checks, post-trade surveillance, immutable logs, idempotency guarantees.            |
| External Market Data Provider | _N/A_                  | Well-defined streaming API (subscribe/unsubscribe); order routing contracts for simulated exchanges.       |



# 2. Architecture Constraints

| Constraint  | Description                                                                                                                                                                                  |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Technology  | C#, Go, Rus or C++ permitted. Python or JavaScript/TypeScript are strongly discouraged for the backend part of the system.                                                                   |
| Performance | The system must meet latency, throughput and availability targets.                                                                                                                           |
| Deployment  | The system prototype must be containerized and deployed on a public or semi publicly accessible platform via an automatic CI/CD pipeline. Multiple artifacts must be deployed during phase 2. |



# 3. System Scope and Context

## Business Context

![SVG Image](bounded_context.svg)

## Technical Context

# 4. Solution Strategy

| Problem                        | Solution                                                                                                                                                                                                                                                                                      |
| ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Maintainability and evolution  | Internally, the system code is organized in a modular fashion using an architecture similar to hexagonal. This isolates the domain logic from infrastructure adapters. This allows the monolithic prototype to evolve somewhat easily in later phases.                                        |
| Data consistency and integrity | A MySQL database is used to store all data and acts as a persistent ledger. Entities and data queries and commands are managed with repositories which abstract the database specific logic into interfaces. Transactions are used in critical flows like order executions to maintain consistency. A Redis sorted set acts as a live order order and orders statuses are synced to the database at an interval. Redis queues are also used to insert order data in batches. |
| Delivering data to client      | The Go server supports multiple http endpoints for server rendered html templates and data retrieval.                                                                                                                                                                                         |
| Latency and throughput         | Using the Go language because it is well-suited for high concurrency, low latency environments. Goroutines are easily implementable for asynchronous processing, making order acknowledgement faster.                                                                                         |
| Error Handling & Observability | The Go programming language is made for functions to return multiple return values, making it very easy to propagate errors from any layers back to the client. It also comes with a an integrated logging library, ensuring the system's behaviors and faults are observable at any point.   |


# 5. Building Block View

![alt text](c4_level1.png)

![alt text](c4-service-level.png)

![alt text](package-level-diagram.png)

![SVG Image](class_diagram.svg)

# 6. Runtime View

## General Use Case Diagram

![SVG Image](use_cases/allusecases.png)

> The use cases colored in green are the ones that are currently implemented in the system

> Note that not all use case scenarios and runtime diagrams have been updated for phase 2 due to lack of time.

## UC-01 Sign up

### Goal:

Allow a registered user to create a user account so that they can later authenticate to BrokerX and use it.

### Actors:

- Primary: Client (user)
- Supporting: MySQL DB

### Preconditions:
N.A.

### Main Flow:

**1.** The Client submits email, password, first name and last name to the system.

**2.** The system receives the credentials, validates them and creates a new user in the database.

**3.** The system returns a confirmation of the user registration to the user.

### Postconditions:

- User account is created is the system

### Exceptions / Alternatives:

**E2.** Database unavailable → The system returns an error 500 Internal Server Error.

**A1.1** User id header not present in request → The system returns an error 401 Unauthorized.

**A1.2** Invalid credentials (wrong JSON formatting) → The system returns an error 400 Bad Request.

**A2.** User account with given email already exists → The system returns an error 409

## UC-02 Login

### Goal:

Allow a registered user to securely authenticate into the BrokerX system.

### Actors:

- Primary: Client (user)
- Supporting: MySQL DB

### Preconditions:

- The user is already registered in the system.
- Credentials (email, password hash) are stored in the database.

### Main Flow:

**1.** The Client submits email and password to the system.

**2.** The system receives the credentials and validates them.

**3.** The system returns the home page, along with a signed session cookie to the client

### Postconditions:

- User is authenticated and can call secured endpoints with the cookie.

### Exceptions / Alternatives:

**E2.** Database unavailable → The system returns an error 500 Internal Server Error.

**A3.** Invalid credentials → The system returns an error 401 Unauthorized.

**A3.** Third failed authentication attempt → The system returns an error 401 Unauthorized and locks the account for 30 minutes.

![SVG Image](use_cases/uc02_sequence.svg)

## UC-05 Place order

### Goal:

Allow a user to place a buy or sell order on a stock.

### Actors:

- Primary: Client (user)
- Supporting: Market Data Provider

### Preconditions:

- The user is logged in.
- The Market Data Provider is up and running.
- The client is on the Orders page

### Main Flow:

**1.** The Client completes the form to place an order containing the symbol, quantity, action, type, timing and unit price.

**2.** The system receives the order data and validates that the inputs are not empty.

**3.** The system starts verification the compliance of the order.

**4.** The system requests instrument and price data from the Market Data Provider.

**5.** The Market Data Provider responds with instrument tick size and price band.

**6.** The system finishes verifying the compliance of the order.

**7.** The system creates an order.

**8.** The system submits the order to the internal matching engine.

**9.** The system send an acknowledgement to the client that the order has been placed.

**10.** The client receives the acknowledgement.

### Postconditions:

- The order is stored in the database with open status.
- The client can see the new placed order with all its information.

### Exceptions / Alternatives:

**E3.** The order has an invalid quantity → The system propagates the error back to the client. The order is not created.

**E5.** The Market Data Provider does not respond → The system propagates the error back to the client. The order is not created.

**E5.** The user does not have enough funds → The system propagates the error back to the client. The order is not created.

**E5.** The user does not have enough stocks → The system propagates the error back to the client. The order is not created.

**E5.** The order tick size or price band does not match the instrument's required tick size or price band → The system propagates the error back to the client. The order is not created.

![SVG Image](use_cases/uc05_sequence.svg)

## UC-07 Find match and execute order

Match incoming buy and sell orders and generates execution records.

### Actors:

- Primary: Client (user)
- Supporting: _None_

### Preconditions:

- A valid order was placed by the client and submitted to the internal matching engine

### Main Flow:

**1.** The matching engine receives the order.

**2.** The matching engine fetches all matching orders (symbol, action, type) from the database.

**3.** The matching engine finds a match or multiple matches for the submitted order.

**4.** The matching engine generates execution records for each match order.

**5.** The matching engine updates order(s) quantities and status.

**5.** The matching engine saves the execution records to the database.

### Postconditions:

- The order changes are saved to the database.
- The client can see the submitted order with all its information and the status and quantities updated.
- The client can see each execution record for the order

### Exceptions / Alternatives:

**E2.** There is a database error when fetching the orders → The system logs the error. The order stays in the database. The order is not updated.

**A2.** The systeme finds no match for the order → The order stays in the database. The order is not updated. No execution records are saved.

**E5.** There is a database error when updating the order(s) → The system logs the error. The order stays in the database. The order is not updated. The execution records are not saved.

**E6.** There is a database error when saving the execution records → The system logs the error. The order stays in the database. The order is not updated. The execution records are not saved.

![SVG Image](use_cases/uc07_sequence.svg)


# 7. Deployment View

![SVG Image](deployment.svg)


# 8. Cross-cutting Concepts

- Architecture semi-hexagonale
- Micro-services
- Server-side HTML rendring
- Interfaces
- MySQL relational database
- Repository pattern
- Transaction pattern


# 9. Design Decisions

---

## ADR-01: Hexagonal architecture

### Context

BrokerX must be evolvable across phases: from monolithic prototype to micro-services and eventually event-driven. A tightly coupled MVC design would limit flexibility. We need an architecture that clearly separates domain logic from infrastructure to make refactoring manageable.

### Decision

Adopt a hexagonal style architecture (ports and adapters):

- **Core domain layer**: entities, matching engine, business rules
- **Ports**: interfaces for any internal/external service, data access objects
- **Adapters**: implementations of the interfaces (REST API handlers, SQL repositories, mock data provider)

### Status

Accepted

### Consequences

- Domain logic is independent of delivery and persitence mechanisms
- Easier to swap infrastructure (database, APIs) without touching core logic
- Simplifies testing (domain logic can be tested in isolation)
- Slightly higher complexity (more abstractions, interfaces, files)
- May be over engineering for phase 1 but will most likely pay off in later phases

---

## ADR-02: Persistence with MySQL and Repository/Transaction Manager Pattern

### Context

BrokerX must maintain strong consistency for orders, executions and balances. Persistent storage was required by the project and we want to avoid spreading raw SQL queries across the codebase.

### Decision

Use MySQL relational database as the system of record, accessed through repository interfaces in the ports layers. Repositories abstract database-specific logic into adapters. Transactions are applied for critical operations (order matching).

### Status

Accepted

### Consequences

- Centralized persistence logic, easier to maintain
- Enforces data integrity using SQL constraints and transactions
- Repository interfaces decouple domain logic from data access details, allowing easier replacement of persistent storage in the future.
- Less flexibility in schema evolution if requirements change rapidly

---

## ADR-03: Use of Go as Implementation Language

### Context

The backend must support low latency and high throughput for order placement and matching. It must also be portable across environments (developer laptops, CI/CD pipelines, VMs). The following languages were considered :

- Java/C#
- Python/Node.js
- Rust/C

### Decision

Implement BrokerX backend API server in Go (Golang)

### Status

Accepted

### Consequences

- Very fast compilation and deployment cycle.
- Excellent concurrency support for handling many requests/orders.
- Small memory footprint, portable binaries.
- Clean error handling via multiple return values.
- Newer language, meaning lack of libraries in certain domains


## ADR-04: NGINX API Gateway

### Context


### Decision

Implement API Gateway with NGINX.

### Status

Accepted

### Consequences

## ADR-05: Redi-Based Order Book

### Context


### Decision

Implement Order Book using Redis Sorted Set.

### Status

Accepted

### Consequences


# 10. Quality Requirements

## Latency

## Throughput

## Data Integrity

## Security

## Testability

## Interoperability


# 11. Risks and Technical Debts

- Writing a lot of code in a short amount of time can lead to lack of unit testing and therefore, quality degradation.
- Lack of code review due to the project being individually developped.
- There is a risk of long refactoring time between phases depending on the evolvability of the system.


# 12. Glossary

| Term                       | Definition                                                                                                                                         |
| -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| **ADR**                    | Architecture Decision Record: a short document that captures an important architectural choice, its context, and consequences.                     |
| **Broker**                 | An intermediary that executes buy/sell orders on behalf of investors. BrokerX acts as an online broker for retail clients.                         |
| **CI**                     | Continuous Integration: automation of integrating new code into the main branch with quality checks and automated tests.                           |
| **CD**                     | Continuous Deployment: automation of delivering tested code to production environments.                                                            |
| **Execution**              | The successful match of a buy and sell order, resulting in a trade and an execution report.                                                        |
| **Execution Report**       | A message generated by the matching engine that details the outcome of an order execution (price, quantity, time).                                 |
| **Hexagonal Architecture** | Also called Ports and Adapters: an architectural style that separates the domain logic (core) from external systems (DB, UI, APIs).                |
| **Latency**                | The time it takes for a user action (e.g., placing an order) to receive an acknowledgement from the system.                                        |
| **Limit Order**            | An order to buy or sell a stock at a specified price or better.                                                                                    |
| **Market Data Provider**   | External or simulated service that supplies price and instrument data to BrokerX.                                                                  |
| **Market Order**           | An order to buy or sell immediately at the current market price.                                                                                   |
| **Matching Engine**        | The core system component that pairs buy and sell orders and generates execution records.                                                          |
| **Monolith**               | An application built as a single deployable unit. BrokerX phase 1 is implemented as a Go monolith.                                                 |
| **MySQL**                  | Relational database used by BrokerX to persist orders, executions, users, and balances.                                                            |
| **Order**                  | A request placed by a user to buy or sell a stock (can be market or limit, buy or sell, ioc or day).                                               |
| **IOC**                    | Immediate Or Cancel. Refers to the timing of an order that must be immediately fully filled, if not, it must be cancelled                          |
| **DAY**                    | Refers to the timing of an order that is open until the end of the trading day.                                                                    |
| **Order Book**             | The collection of all open buy and sell orders for a given instrument, managed by the matching engine.                                             |
| **Port**                   | Interface in hexagonal architecture that represents an entry/exit point for the core domain logic (e.g., repository interface, service interface). |
| **Adapter**                | Implementation of a port, connecting the domain logic to infrastructure (e.g., SQL repository, HTTP services).                                     |
| **Pre-trade Checks**       | Compliance rules applied before accepting an order (e.g., tick size validation, price bands, sufficient balance/holdings).                         |
| **Session Cookie**         | Small piece of data stored on the client after login, used to authenticate subsequent requests.                                                    |
| **Tick Size**              | The minimum price movement allowed for a given instrument.                                                                                         |
| **Throughput**             | Number of orders the system can successfully handle per second.                                                                                    |
| **Transaction**            | A group of one or more database operations executed as a single unit of work, ensuring atomicity and consistency.                                  |
| **Wallet**                 | Virtual balance associated with a user account, used to fund purchases and receive proceeds from sales.                                            |
