
-- Создание таблицы password
CREATE TABLE password (
                          id SERIAL PRIMARY KEY,
                          value BYTEA
);

-- Создание таблицы profile_role
CREATE TABLE profile_role (
                              id SERIAL PRIMARY KEY,
                              profile_id INT,
                              role_id INT
);

-- Создание таблицы profile
CREATE TABLE profile (
                         id SERIAL PRIMARY KEY,
                         login TEXT NOT NULL UNIQUE,
                         password_id INT NOT NULL,
                         profile_role_id INT,
                         CONSTRAINT fk_password FOREIGN KEY (password_id) REFERENCES password (id),
                         CONSTRAINT fk_profile_role FOREIGN KEY (profile_role_id) REFERENCES profile_role (id)
);

-- Создание таблицы role
CREATE TABLE role (
                      id SERIAL PRIMARY KEY,
                      value TEXT
);


INSERT INTO role(value) VALUES ('user'), ('admin');
