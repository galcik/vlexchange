CREATE TABLE account
(
    id         SERIAL PRIMARY KEY,
    username   varchar          NOT NULL,
    token      varchar(50)      NOT NULL UNIQUE,
    usd_amount bigint DEFAULT 0 NOT NULL,
    btc_amount bigint DEFAULT 0 NOT NULL
);

CREATE TYPE order_type AS ENUM ('buy', 'sell');
CREATE TYPE order_state AS ENUM ('live', 'fulfilled', 'cancelled');

CREATE TABLE standing_order
(
    id                  SERIAL PRIMARY KEY,
    account_id          integer                    NOT NULL REFERENCES account (id),
    type                order_type                 NOT NULL,
    state               order_state DEFAULT 'live' NOT NULL,
    quantity            bigint      DEFAULT 0      NOT NULL,
    filled_quantity     bigint      DEFAULT 0      NOT NULL,
    filled_price        bigint      DEFAULT 0      NOT NULL,
    limit_price         bigint      DEFAULT 0      NOT NULL,
    reserved_usd_amount bigint      DEFAULT 0      NOT NULL,
    reserved_btc_amount bigint      DEFAULT 0      NOT NULL,
    webhook_url         text
);

CREATE
    INDEX standing_order_account_id_idx ON standing_order (account_id);