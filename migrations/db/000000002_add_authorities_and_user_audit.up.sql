CREATE TABLE IF NOT EXISTS roles (
    role VARCHAR (50) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_authority serial PRIMARY KEY,
    user_id int references users(user_id),
    role VARCHAR (50) references roles(role)
);