CREATE TABLE snapshots (
    kind varchar(255) not null,
    identity uuid not null,
    data blob not null,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE blueprints (
    id uuid not null,
    kind varchar(255) not null,
    data json not null,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);