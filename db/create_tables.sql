DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS games;
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

CREATE TABLE players (
	id			INT AUTO_INCREMENT NOT NULL,
	game_id		INT NOT NULL,
	user_id		INT NOT NULL,
	progress	INT DEFAULT 0,
	PRIMARY KEY (`id`)
);

-- So Linus is joining the 'you are going to lose game'
-- Since he's the only player, he must have supposedly been the host
INSERT INTO players
	(game_id, user_id)
VALUES
	(0, 1);
