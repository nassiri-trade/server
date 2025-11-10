CREATE TABLE IF NOT EXISTS calendar_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_hash TEXT NOT NULL UNIQUE,
    event_date DATETIME NOT NULL,
    currency TEXT NOT NULL,
    impact TEXT NOT NULL,
    detail TEXT NOT NULL,
    actual TEXT,
    forecast TEXT,
    previous TEXT,
    source_url TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_calendar_events_date ON calendar_events (event_date DESC);

