-- Remove alarm notification indexes
DROP INDEX idx_alarm_time ON user_alarms;
DROP INDEX idx_last_sent ON user_alarms;
DROP INDEX idx_user_tokens ON weather_service_tokens;
