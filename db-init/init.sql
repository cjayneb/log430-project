CREATE DATABASE IF NOT EXISTS brokerx;
USE brokerx;

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    status ENUM('pending', 'active') NOT NULL DEFAULT 'pending',
    failed_attempts INT NOT NULL DEFAULT 0,
    locked_until DATETIME NULL
);
CREATE UNIQUE INDEX idx_users_email ON users(email);

INSERT INTO users (email, password, first_name, last_name, status) VALUES
('email', '$2a$14$VWlwuLF38a4lcpkmsBk9Bulkanjd2mauqYDkU9Y5OziSgbA9CryZG', 'fn', 'ln', 'active'),
('buyer@email.com', '$2a$14$VWlwuLF38a4lcpkmsBk9Bulkanjd2mauqYDkU9Y5OziSgbA9CryZG', 'buyer', 'man', 'active'),
('seller@email.com', '$2a$14$VWlwuLF38a4lcpkmsBk9Bulkanjd2mauqYDkU9Y5OziSgbA9CryZG', 'seller', 'woman', 'active');

CREATE TABLE IF NOT EXISTS wallets (
    id CHAR(36) PRIMARY KEY,
    user_id INT NOT NULL,
    -- Three-balance model
    settled_funds DECIMAL(14, 2) NOT NULL DEFAULT 0,   -- real ledger cash
    available_funds DECIMAL(14, 2) NOT NULL DEFAULT 0, -- funds user can spend
    reserved_funds DECIMAL(14, 2) NOT NULL DEFAULT 0,  -- funds locked for BUY orders
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX idx_wallets_user ON wallets(user_id);

INSERT INTO wallets (id, user_id, settled_funds, available_funds, reserved_funds) VALUES
(UUID(), 1, 0,        0,        0),
(UUID(), 2, 10000000, 10000000, 0),
(UUID(), 3, 300,      300,      0);

-- ============================================
-- WALLET LEDGER (append-only audit trail)
-- Required by audit & exactly-once rules
-- ============================================
CREATE TABLE IF NOT EXISTS wallet_ledger (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    wallet_id CHAR(36) NOT NULL,
    user_id INT NOT NULL,

    change_type ENUM(
        'deposit',
        'withdrawal',
        'reserve_funds',
        'release_funds',
        'execution_debit',
        'execution_credit'
    ) NOT NULL,

    amount DECIMAL(14, 2) NOT NULL,
    balance_settled DECIMAL(14, 2) NOT NULL,   -- snapshot after change
    balance_available DECIMAL(14, 2) NOT NULL,
    balance_reserved DECIMAL(14, 2) NOT NULL,

    reference_id VARCHAR(64) NULL, -- order_id, exec_id etc.
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (wallet_id) REFERENCES wallets(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX idx_wallet_ledger_wallet ON wallet_ledger(wallet_id);
CREATE INDEX idx_wallet_ledger_user ON wallet_ledger(user_id);

CREATE TABLE IF NOT EXISTS orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    type ENUM('market', 'limit') NOT NULL,
    action ENUM('buy', 'sell') NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    remaining_quantity INT NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    timing ENUM('day', 'ioc') NOT NULL,
    status ENUM('open', 'partially_filled', 'filled', 'canceled') NOT NULL,
    status_reason VARCHAR(300) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX idx_orders_user ON orders(user_id);

CREATE TABLE IF NOT EXISTS positions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    available_quantity INT NOT NULL,
    reserved_quantity INT NOT NULL DEFAULT 0,
    unit_price DECIMAL(10, 2) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX idx_positions_user ON positions(user_id);

INSERT INTO positions (user_id, symbol, available_quantity, unit_price) VALUES
(3, 'AAPL', 150, 400.00);

CREATE TABLE IF NOT EXISTS executions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    buy_order_id INT NOT NULL,
    sell_order_id INT NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (buy_order_id) REFERENCES orders(id),
    FOREIGN KEY (sell_order_id) REFERENCES orders(id)
);

CREATE TABLE IF NOT EXISTS outbox_order_events (
    id INT PRIMARY KEY AUTO_INCREMENT,
    topic VARCHAR(64) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    trace_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NULL,
    jwt_token TEXT NULL,
    payload JSON NOT NULL,
    status ENUM('pending', 'published', 'failed') NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    next_attempt_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    error_message VARCHAR(64) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
CREATE INDEX idx_outbox_status ON outbox_order_events(status);
CREATE INDEX idx_outbox_trace ON outbox_order_events(trace_id);
CREATE INDEX idx_outbox_pending ON outbox_order_events(status, next_attempt_at);
