DROP DATABASE IF EXISTS np1finder;

CREATE DATABASE IF NOT EXISTS np1finder;

USE np1finder;

CREATE TABLE users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL
);

CREATE TABLE posts (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  description TEXT NOT NULL
);

CREATE TABLE comments (
  id INT AUTO_INCREMENT PRIMARY KEY,
  post_id INT NOT NULL,
  description TEXT NOT NULL
);

INSERT INTO users (name) VALUES ('rrreeeyyy'), ('Shirai Kuroko');
INSERT INTO posts (user_id, description) VALUES (1, 'Hello, world!'), (2, 'Judgement desuno!'), (2, 'Level 4 Teleporter desuno!');
INSERT INTO comments (post_id, description) VALUES (1, 'Nice to meet you!'), (2, 'Judgement desuno!'), (2, 'Level 4 Teleporter desuno!');
