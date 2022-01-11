DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS games;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS players;

CREATE TABLE accounts (
	id			INT AUTO_INCREMENT NOT NULL,
	username	VARCHAR(128) NOT NULL,
	password	VARCHAR(255) NOT NULL,
	wins		INT DEFAULT 0,
	PRIMARY KEY (`id`)
);

INSERT INTO accounts
	(username, password, wins)
VALUES
	('Michael Skyba', '1234567890', 4),
	('Linus Torvalds', 'Hunter1', 0),
	('password', 'password', 912390);

CREATE TABLE games (
	id			INT AUTO_INCREMENT NOT NULL,
	name		VARCHAR(128) NOT NULL,
	password	VARCHAR(128) NOT NULL,
	PRIMARY KEY (`id`)
);

INSERT INTO games
	(name, password)
VALUES
	('you are going to lose', '');

CREATE TABLE questions (
	id			INT AUTO_INCREMENT NOT NULL,
	game_id		INT NOT NULL,
	text		VARCHAR(128) NOT NULL,
	answer		INT NOT NULL,
	progress	INT NOT NULL,
	PRIMARY KEY (`id`)
);

INSERT INTO questions
	(game_id, text, answer, progress)
VALUES
	(1, '1 × 1 + 1 - 1 test', 1, 0),
	(1, '1 × 2', 2, 1),
	(1, '1 × 3', 3, 2),
	(1, '1 × 4', 4, 3),
	(1, '1 × 5', 5, 4),
	(1, '1 × 6', 6, 5),
	(1, '1 × 7', 7, 6),
	(1, '1 × 8', 8, 7),
	(1, '1 × 9', 9, 8),
	(1, '1 × 10', 10, 9);

CREATE TABLE players (
	id			INT AUTO_INCREMENT NOT NULL,
	game_id		INT NOT NULL,
	user_id		INT NOT NULL,
	progress	INT DEFAULT -1,
	PRIMARY KEY (`id`)
);

INSERT INTO players
	(game_id, user_id, progress)
VALUES
	(1, 2, -1);
