CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscriptions (
    id           UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID        NOT NULL,
    service_name TEXT        NOT NULL,
    price        INTEGER     NOT NULL CHECK (price >= 0),
    start_date   DATE        NOT NULL,
    end_date     DATE        NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT subscriptions_date_range CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id      ON subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions (service_name);
CREATE INDEX IF NOT EXISTS idx_subscriptions_start_date   ON subscriptions (start_date);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS subscriptions_set_updated_at ON subscriptions;
CREATE TRIGGER subscriptions_set_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();