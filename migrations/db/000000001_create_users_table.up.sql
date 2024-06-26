CREATE TABLE IF NOT EXISTS users(
    user_id serial PRIMARY KEY,
    username VARCHAR (50) UNIQUE NOT NULL,
    password VARCHAR (128) NOT NULL,
    created_at timestamp,
    updated_at timestamp default current_timestamp,
    deleted_at timestamp
);


















