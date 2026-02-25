-- Add co-op game state columns to td_rooms table

ALTER TABLE td_rooms 
ADD COLUMN player1_state JSON COMMENT 'Player 1 game state (round, hp, gold, kills)',
ADD COLUMN player2_state JSON COMMENT 'Player 2 game state (round, hp, gold, kills)',
ADD COLUMN last_p1_update TIMESTAMP NULL COMMENT 'Last update time for player 1',
ADD COLUMN last_p2_update TIMESTAMP NULL COMMENT 'Last update time for player 2';

-- Example JSON structure for player_state:
-- {
--   "round": 5,
--   "hp": 87,
--   "gold": 1200,
--   "kills": 45,
--   "timestamp": 1234567890,
--   "is_alive": true
-- }
