DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS games;

CREATE TABLE accounts (
	id			INT AUTO_INCREMENT NOT NULL,
	username	VARCHAR(128) NOT NULL,
	password	VARCHAR(255) NOT NULL,
	PRIMARY KEY (`id`)
);

INSERT INTO accounts
	(username, password)
VALUES
	('Michael Skyba', '1234567890'),
	('Linus Torvalds', 'Hunter1'),
	('password', 'password');
