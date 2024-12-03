CREATE TABLE IF NOT EXISTS users(
	id VARCHAR(36) PRIMARY KEY,
	firstname VARCHAR(60) DEFAULT "",
	lastname VARCHAR(60) DEFAULT "",
	email VARCHAR(120) DEFAULT "",
    extra VARCHAR(256) DEFAULT ""
);

INSERT IGNORE INTO users(id, firstname, lastname, email, extra) VALUES('5bf79a64-40bb-9161-5385-b02dcb8948d5', 'Dick', 'Hardt', 'admin@oauth.labs', 'flag{871d1dec99b890a924ebd803e77c4ea1ccb6f8c7}');
