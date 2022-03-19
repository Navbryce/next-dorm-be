CREATE TABLE IF NOT EXISTS person
(
    firebase_id  VARCHAR(36)  NOT NULL,
    display_name VARCHAR(300) NOT NULL,
    PRIMARY KEY (firebase_id),
    UNIQUE INDEX IDX_display_name (display_name)
);

CREATE TABLE IF NOT EXISTS community
(
    id   MEDIUMINT    NOT NULL AUTO_INCREMENT,
    name VARCHAR(500) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS subscription
(
    user_id VARCHAR(36) NOT NULL,
    community_id MEDIUMINT NOT NULL,
    PRIMARY KEY(user_id, community_id)
);

CREATE TABLE IF NOT EXISTS content_metadata
(
    id            INT                        NOT NULL AUTO_INCREMENT,
    creator_id    VARCHAR(36)                NOT NULL,
    creator_alias VARCHAR(300)               NOT NULL,
    visibility    ENUM ('NORMAL', 'HIDDEN')  NOT NULL,
    status        ENUM ('POSTED', 'DELETED') NOT NULL DEFAULT 'POSTED',
    vote_total    INT                                 DEFAULT 0,
    num_votes     INT                                 DEFAULT 0,
    created_at    DATETIME                   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME                   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX IDX_BY_CREATED_AT (created_at DESC, id DESC),
    INDeX IDX_CREATOR (creator_id)
);

CREATE TABLE IF NOT EXISTS post
(
    id          INT          NOT NULL AUTO_INCREMENT,
    metadata_id INT          NOT NULL,
    title       VARCHAR(500) NOT NULL,
    content     TEXT         NOT NULL,
    PRIMARY KEY (id),
    UNIQUE INDEX IDX_CONTENT_METADATA (metadata_id)
);

CREATE TABLE IF NOT EXISTS post_communities
(
    post_id      INT       NOT NULL,
    community_id MEDIUMINT NOT NULL,
    PRIMARY KEY (post_id, community_id)
);

CREATE TABLE IF NOT EXISTS vote
(
    value           TINYINT     NOT NULL,
    tgt_metadata_id INT         NOT NULL,
    voter_id        VARCHAR(36) NOT NULL,
    PRIMARY KEY (tgt_metadata_id, voter_id)
    # TODO: NEED INDEX ON voter_id here?
);

CREATE TABLE IF NOT EXISTS comment
(
    id                 INT  NOT NULL AUTO_INCREMENT,
    root_metadata_id   INT  NOT NULL,
    parent_metadata_id INT  NOT NULL,
    metadata_id        INT  NOT NULL,
    content            TEXT NOT NULL,
    PRIMARY KEY (id),
    UNIQUE INDEX IDX_CONTENT_METADATA (metadata_id),
    INDEX IDX_PARENT (parent_metadata_id)
);

CREATE TABLE IF NOT EXISTS report
(
    id              INT         NOT NULL AUTO_INCREMENT,
    tgt_metadata_id INT         NOT NULL,
    creator_id      VARCHAR(36) NOT NULL,
    reason          TEXT        NOT NULL,
    created_at      DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX IDX_BY_POST (tgt_metadata_id)
)
