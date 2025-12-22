-- Initial schema for verdict-agent

CREATE TABLE decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    input TEXT NOT NULL,
    verdict JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_final BOOLEAN DEFAULT TRUE
);

CREATE TABLE todos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id UUID REFERENCES decisions(id),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_decisions_created_at ON decisions(created_at DESC);
CREATE INDEX idx_todos_decision_id ON todos(decision_id);
