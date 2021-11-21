CREATE DATABASE IF NOT EXISTS ffdb;
USE ffdb;
CREATE TABLE IF NOT EXISTS links (
    id int NOT NULL AUTO_INCREMENT,
    domain VARCHAR(255),
    path VARCHAR(2047),
    destination VARCHAR(4095),
    times_reported int NOT NULL DEFAULT 0,
    hashed_IP CHAR(64),
    votedfordeletion TINYINT(1) DEFAULT 0,
    voted_by VARCHAR(255),
    PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS recycle_bin (
    id int NOT NULL AUTO_INCREMENT,
    domain VARCHAR(255),
    path VARCHAR(2047),
    destination VARCHAR(4095),
    times_reported int NOT NULL DEFAULT 0,
    hashed_IP CHAR(64),
    votedfordeletion TINYINT(1) DEFAULT 0,
    voted_by VARCHAR(255),
    note VARCHAR(255),
    PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS admin_creds (
    id int NOT NULL AUTO_INCREMENT,
    username VARCHAR(255),
    password  CHAR(60) BINARY,
    token_id VARCHAR(255),
    PRIMARY KEY(id)
);
