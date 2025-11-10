CREATE TABLE IF NOT EXISTS user_positions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    ticket INTEGER NOT NULL UNIQUE,
    platform TEXT,
    account_name TEXT,
    account_login TEXT,
    broker_server TEXT,
    broker_name TEXT,
    balance REAL,
    status TEXT NOT NULL DEFAULT 'open',
    symbol TEXT,
    side TEXT,
    volume REAL,
    entry_time DATETIME,
    entry_price REAL,
    stop_loss REAL,
    take_profit REAL,
    profit REAL,
    risk_percent REAL,
    tick_value REAL,
    tick_size REAL,
    magic INTEGER,
    reason TEXT,
    comment TEXT,
    last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    raw_payload TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_positions_user_id ON user_positions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_positions_status ON user_positions (status);

CREATE TABLE IF NOT EXISTS user_trades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    ticket INTEGER NOT NULL,
    position_ticket INTEGER,
    platform TEXT,
    symbol TEXT,
    side TEXT,
    entry_type TEXT,
    volume REAL,
    entry_time DATETIME,
    entry_price REAL,
    stop_loss REAL,
    take_profit REAL,
    profit REAL,
    risk_percent REAL,
    tick_value REAL,
    tick_size REAL,
    magic INTEGER,
    reason TEXT,
    comment TEXT,
    raw_payload TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, ticket)
);

CREATE INDEX IF NOT EXISTS idx_user_trades_user_id ON user_trades (user_id);
CREATE INDEX IF NOT EXISTS idx_user_trades_position_ticket ON user_trades (position_ticket);
CREATE INDEX IF NOT EXISTS idx_user_trades_entry_time ON user_trades (entry_time);

