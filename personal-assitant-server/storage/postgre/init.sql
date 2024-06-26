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
    timezone   VARCHAR(255), --NOT NULL,
    avatar_url VARCHAR(255),
    phone      VARCHAR(255)
);


CREATE TABLE tasks
(
    id serial PRIMARY KEY,
    user_id INT REFERENCES users(id),
    start_date timestamp,
    planned_date timestamp,
    finished_date timestamp,
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
--TODO не забыть использовать это в коде
ALTER TABLE  tasks add column  attachment varchar (255);
ALTER TABLE  tasks add column  title varchar (255);

CREATE TABLE users_notification_settings
(
    id serial PRIMARY KEY,
    user_id   INT REFERENCES users(id),
    for_morning int,
    for_evening int,
    for_task int,
    is_dark_them boolean
);
CREATE TABLE finances_categories
(
    id serial PRIMARY KEY,
    name varchar(255),
    is_for_all_users boolean default false,
    user_id INT REFERENCES users(id)
);
CREATE TABLE finances
(
    id serial PRIMARY KEY,
    category_id INT REFERENCES finances_categories(id),
    fin_date date,
    description varchar(550),
    price varchar(250),
    is_expense boolean default true,
    user_id INT REFERENCES users(id)

);


INSERT INTO finances_categories (name, is_for_all_users) VALUES
    ('Подписки',true),
    ('Фитнес',true),
    ('Продукты',true),
    ('Рестораны',true),
    ('Транспорт',true),
    ('Автомобиль',true),
    ('Подарки',true),
    ('Развлечения',true),
    ('Одежда',true),
    ('Обувь',true),
    ('Путешествия',true);

ALTER TABLE  finances add column  is_expense boolean default true;
ALTER TABLE  finances add column  user_id INT REFERENCES users(id);
ALTER TABLE  finances_categories add column  user_id INT REFERENCES users(id);
CREATE TABLE archived_tasks
(
    id serial PRIMARY KEY,
    user_id INT REFERENCES users(id),
    start_date timestamp,
    planned_date timestamp,
    finished_date timestamp,
    description VARCHAR(550),
    is_finished boolean default false,
    archived_date timestamp


);
ALTER TABLE  archived_tasks add column  attachment varchar (255);
ALTER TABLE  archived_tasks add column  title varchar (255);

