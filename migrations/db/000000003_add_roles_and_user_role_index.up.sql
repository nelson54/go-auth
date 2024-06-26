INSERT INTO roles VALUES (role='ROLE_USER'),
                         (role='ROLE_ADMIN'),
                         (role='ROLE_SYSTEM');

CREATE UNIQUE INDEX user_roles_idx ON user_roles(user_id, role);
