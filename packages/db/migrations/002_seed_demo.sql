-- 002_seed_demo.sql — demo data for local dev (optional, idempotent)

INSERT INTO users (email, password_hash) VALUES
('alex@outpost.engine', '$2a$10$demo_hash_placeholder')
ON CONFLICT (email) DO NOTHING;

-- need user id for FK
DO $$
DECLARE
  demo_user_id INT;
  linkedin_id INT;
BEGIN
  SELECT id INTO demo_user_id FROM users WHERE email = 'alex@outpost.engine' LIMIT 1;

  INSERT INTO social_accounts (user_id, platform, platform_user_id, access_token, refresh_token, token_expires_at)
  VALUES (demo_user_id, 'linkedin', 'demo-linkedin-uid', 'enc_access_token_demo', 'enc_refresh_token_demo', NOW() + INTERVAL '60 days')
  ON CONFLICT (user_id, platform, platform_user_id) DO NOTHING
  RETURNING id INTO linkedin_id;

  -- fallback if already exists
  SELECT id INTO linkedin_id FROM social_accounts WHERE user_id = demo_user_id AND platform = 'linkedin' LIMIT 1;

  INSERT INTO posts (user_id, social_account_id, content, media_urls, status, scheduled_at)
  VALUES
  (demo_user_id, linkedin_id, 'Building a queue system with Go + Redis taught me: reliability is a feature.', '[]'::jsonb, 'PENDING', NOW() + INTERVAL '2 hours'),
  (demo_user_id, linkedin_id, 'Editorial Minimal Lux — why scheduler tools should feel like Linear, not Canva.', '[]'::jsonb, 'PENDING', NOW() + INTERVAL '1 day')
  ON CONFLICT DO NOTHING;
END $$;
