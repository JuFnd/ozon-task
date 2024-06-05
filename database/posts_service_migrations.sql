DROP TABLE IF EXISTS posts CASCADE;
CREATE TABLE IF NOT EXISTS posts (
                                     id SERIAL PRIMARY KEY,
                                     user_id INT NOT NULL DEFAULT 0,
                                     content TEXT NOT NULL DEFAULT '',
                                     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                     comments_allowed bool NOT NULL DEFAULT true
);

DROP TABLE IF EXISTS comments CASCADE;
CREATE TABLE IF NOT EXISTS comments (
                                        id SERIAL PRIMARY KEY,
                                        user_id INT NOT NULL DEFAULT 0,
                                        post_id INT NOT NULL DEFAULT 0,
                                        parent_id INT NOT NULL DEFAULT 0,
                                        content TEXT NOT NULL DEFAULT '',
                                        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);