ALTER TABLE users
    ADD created_at timestamp,
    ADD updated_at timestamp default current_timestamp,
    ADD deleted_at timestamp;

CREATE TABLE IF NOT EXISTS roles (
    role VARCHAR (50) PRIMARY KEY
);

INSERT INTO roles VALUES (role='ROLE_USER'),
                         (role='ROLE_ADMIN'),
                         (role='ROLE_SYSTEM');

CREATE TABLE IF NOT EXISTS user_roles (
    user_authority serial PRIMARY KEY,
    user_id int references users(user_id),
    role VARCHAR (50) references roles(role)
);

CREATE UNIQUE INDEX user_roles_idx ON user_roles(user_id, role);


