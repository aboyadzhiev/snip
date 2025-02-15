CREATE TABLE IF NOT EXISTS url_map
(
    id           BIGINT                              NOT NULL
        CONSTRAINT url_map_pk
            PRIMARY KEY,
    slug         TEXT                                NOT NULL,
    original_url TEXT                                NOT NULL,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
