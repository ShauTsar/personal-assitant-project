-- Подключаемся к базе данных
\c
novaDB;

-- Создаем таблицу пользователей
CREATE TABLE users
(
    id         serial PRIMARY KEY,
    username   VARCHAR(255) NOT NULL,
    password   VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL,
    timezone   VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(255),
    phone      VARCHAR(255)
);


CREATE TABLE tasks
(
    id serial PRIMARY KEY,
    user_id INT REFERENCES users(id),
    start_date date,
    planned_date date,
    finished_date date,
    description VARCHAR(550),
    is_finished boolean default false


);
CREATE TABLE tbot
(
    id serial PRIMARY KEY,
    user_id INT REFERENCES users(id),
    tg_user_id INT, -- Идентификатор пользователя в телеграме
    task_id INT REFERENCES tasks(id), -- Связь с задачей
    notification_enabled BOOLEAN DEFAULT true, -- Флаг, включены ли уведомления
    notification_time TIME, -- Время для отправки уведомлений
    last_notified_at TIMESTAMP, -- Время последнего уведомления
    CONSTRAINT tg_user_unique UNIQUE (tg_user_id)
);
