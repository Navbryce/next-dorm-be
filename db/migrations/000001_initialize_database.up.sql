CREATE TABLE IF NOT EXISTS person
(
    firebase_id  VARCHAR(36)  NOT NULL,
    display_name VARCHAR(300) NOT NULL UNIQUE,
    is_admin     BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (firebase_id)
);

CREATE TABLE IF NOT EXISTS community
(
    id         MEDIUMINT    NOT NULL AUTO_INCREMENT,
    name       VARCHAR(500) NOT NULL,
    parent_id  MEDIUMINT,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX IDX_BY_PARENT (parent_id),
    UNIQUE INDEX U_IDX_NAME_TO_PARENT (parent_id, name)
);

CREATE TABLE IF NOT EXISTS image
(
    id         INT           NOT NULL AUTO_INCREMENT,
    blob_name  VARCHAR(2048) NOT NULL,
    created_at DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS subscription
(
    user_id      VARCHAR(36) NOT NULL,
    community_id MEDIUMINT   NOT NULL,
    PRIMARY KEY (user_id, community_id)
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
    INDEX IDX_VOTE_TOTAL (vote_total DESC, id DESC),
    INDEX IDX_CREATOR (creator_id)
);

CREATE TABLE IF NOT EXISTS content_image
(
    metadata_id INT NOT NULL,
    image_id    INT NOT NULL,
    PRIMARY KEY (metadata_id, image_id)
);

CREATE TABLE IF NOT EXISTS post
(
    id            INT          NOT NULL AUTO_INCREMENT,
    metadata_id   INT          NOT NULL UNIQUE,
    title         VARCHAR(500) NOT NULL,
    content       TEXT         NOT NULL,
    comment_count INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
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
    created_at      DATETIME    NOT NULL                       DEFAULT CURRENT_TIMESTAMP,
    status          ENUM ('SUBMITTED', 'ACCEPTED', 'REJECTED') DEFAULT 'SUBMITTED',
    PRIMARY KEY (id),
    INDEX IDX_BY_POST (tgt_metadata_id)
)
