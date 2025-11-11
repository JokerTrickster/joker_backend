-- Add indexes for efficient alarm notification queries
CREATE INDEX idx_alarm_time ON user_alarms(alarm_time, is_enabled, deleted_at);
CREATE INDEX idx_last_sent ON user_alarms(last_sent);
CREATE INDEX idx_user_tokens ON weather_service_tokens(user_id, deleted_at);
