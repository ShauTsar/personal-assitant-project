-- Подключаемся к базе данных
\c novaDB;

-- Создаем таблицу пользователей
CREATE TABLE users (
                       id serial PRIMARY KEY,
                       username VARCHAR(255) NOT NULL,
                       password VARCHAR(255) NOT NULL,
                       email VARCHAR(255) NOT NULL,
                       timezone VARCHAR(255) NOT NULL,
                       avatar_url VARCHAR(255),
                       phone VARCHAR(255)
);
