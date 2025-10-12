CREATE TABLE IF NOT EXISTS auth_event_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_event_logs_user_id ON auth_event_logs(user_id);
CREATE INDEX idx_auth_event_logs_created_at ON auth_event_logs(created_at);
CREATE INDEX idx_auth_event_logs_event_type ON auth_event_logs(event_type);
