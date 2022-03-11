CREATE TABLE IF NOT EXISTS person (
    firebase_id VARCHAR(36) NOT NULL,
    display_name VARCHAR(300) NOT NULL,
    PRIMARY KEY(firebase_id)
);

CREATE TABLE IF NOT EXISTS community (
    id MEDIUMINT NOT NULL AUTO_INCREMENT,
    name VARCHAR(500) NOT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS post (
    id INT NOT NULL AUTO_INCREMENT,
    creator_id VARCHAR(36) NOT NULL,
    content TEXT NOT NULL,
    visibility ENUM('NORMAL', 'HIDDEN') NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    vote_total INT DEFAULT 0,
    num_votes INT DEFAULT 0,
    PRIMARY KEY(id),
    INDEX IDX_BY_CREATED_AT (created_at DESC, id DESC)
);

CREATE TABLE IF NOT EXISTS post_communities (
    post_id INT NOT NULL,
    community_id MEDIUMINT NOT NULL,
    PRIMARY KEY(post_id, community_id)
);

CREATE TABLE IF NOT EXISTS vote (
    value TINYINT NOT NULL,
    post_id INT NOT NULL,
    voter_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (post_id, voter_id)
);