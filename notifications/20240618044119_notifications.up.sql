CREATE TABLE notifications (
  id UUID PRIMARY KEY,
  revision TEXT NOT NULL UNIQUE,
  topic TEXT NOT NULL DEFAULT '',
  nlevel TEXT NOT NULL DEFAULT '',
  template TEXT NOT NULL DEFAULT '',
  vars JSONB NOT NULL DEFAULT '[]',
  target TEXT NOT NULL DEFAULT '',
  organisation_id UUID REFERENCES organisations (id) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
  user_id UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
  emailed BOOLEAN NOT NULL DEFAULT false,
  emailed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  seen BOOLEAN NOT NULL DEFAULT false,
  seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE notifications ADD COLUMN ts tsvector
  GENERATED ALWAYS AS
    (  to_tsvector('english', coalesce(topic, ''))
  ) STORED;

CREATE INDEX notifications_ts_idx ON notifications USING GIN (ts);
