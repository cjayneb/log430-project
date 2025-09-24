CREATE DATABASE IF NOT EXISTS brokerx;

USE brokerx;

CREATE TABLE IF NOT EXISTS users (
    id CHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
);
CREATE INDEX idx_users_email ON users(email);

INSERT INTO users (id, email, password, role) VALUES
(UUID(), 'email', '$2a$14$VWlwuLF38a4lcpkmsBk9Bulkanjd2mauqYDkU9Y5OziSgbA9CryZG')
(UUID(), 'buyer@email.com', '$2a$14$VWlwuLF38a4lcpkmsBk9Bulkanjd2mauqYDkU9Y5OziSgbA9CryZG')
(UUID(), 'seller@email.com', '$2a$14$VWlwuLF38a4lcpkmsBk9Bulkanjd2mauqYDkU9Y5OziSgbA9CryZG')
