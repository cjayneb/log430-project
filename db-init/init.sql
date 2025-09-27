CREATE DATABASE IF NOT EXISTS brokerx;

USE brokerx;

CREATE TABLE IF NOT EXISTS users (
    id CHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    failed_attempts INT NOT NULL DEFAULT 0,
    locked_until DATETIME NULL
);
CREATE UNIQUE INDEX idx_users_email ON users(email);

INSERT INTO users (id, email, password) VALUES
(UUID(), 'email', '$2a$14$VWlwuLF38a4lcpkmsBk9Bulkanjd2mauqYDkU9Y5OziSgbA9CryZG'),
(UUID(), 'buyer@email.com', '$2a$14$VWlwuLF38a4lcpkmsBk9Bulkanjd2mauqYDkU9Y5OziSgbA9CryZG'),
(UUID(), 'seller@email.com', '$2a$14$VWlwuLF38a4lcpkmsBk9Bulkanjd2mauqYDkU9Y5OziSgbA9CryZG');

CREATE TABLE IF NOT EXISTS orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id CHAR(36) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    type ENUM('buy', 'sell') NOT NULL,
    action ENUM('market', 'limit') NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    timing ENUM('day', 'ioc') NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

INSERT INTO orders (user_id, symbol, type, action, quantity, unit_price, timing, status) VALUES
((SELECT id FROM users WHERE email = 'email'), 'AAPL', 'buy', 'market', 10, 150.00, 'day', 'open');