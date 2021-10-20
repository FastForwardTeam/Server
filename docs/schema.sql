CREATE TABLE IF NOT EXISTS links (
    id int NOT NULL AUTO_INCREMENT,
    domain VARCHAR(255),
    path VARCHAR(2047),
    destination VARCHAR(4095),
    hashed_IP CHAR(64),
    PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS recycle_bin (
    id int NOT NULL AUTO_INCREMENT,
    domain VARCHAR(255),
    path VARCHAR(2047),
    destination VARCHAR(4095),
    hashed_IP CHAR(64),
    note VARCHAR(255),
    PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS admin_creds (
    id int NOT NULL AUTO_INCREMENT,
    password  CHAR(60) BINARY,
    token_id VARCHAR(255),
    PRIMARY KEY(id)
);
