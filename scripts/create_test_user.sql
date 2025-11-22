-- ============================================================
-- 创建本地测试用户
-- ============================================================
-- 使用方法: psql $DATABASE_URL -f scripts/create_test_user.sql
-- ============================================================

-- 创建测试用户
-- 密码: test123 (已经过bcrypt加密)
-- OTP Secret: JBSWY3DPEHPK3PXP (固定的测试secret)
-- OTP验证码: 使用Google Authenticator扫描或手动输入secret
INSERT INTO users (id, email, password_hash, is_admin, otp_verified, otp_secret, created_at)
VALUES (
    'test-user-local',
    'test@local.dev',
    '$2a$10$aT1VB5d99Do30MzhBK.b6uUIZu/N/d8WUTnZbsXodykwuseQr9fsC', -- 密码: test123
    false,
    true,
    'JBSWY3DPEHPK3PXP', -- 固定的OTP secret，方便测试
    CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    otp_verified = EXCLUDED.otp_verified,
    otp_secret = EXCLUDED.otp_secret;

-- 为测试用户创建AI模型配置
INSERT INTO ai_models (id, user_id, name, provider, enabled, api_key, custom_model_name, created_at)
VALUES (
    'test-deepseek',
    'test-user-local',
    'DeepSeek Test',
    'deepseek',
    true,
    'sk-test-key',
    'deepseek-chat',
    CURRENT_TIMESTAMP
)
ON CONFLICT (id, user_id) DO UPDATE SET
    name = EXCLUDED.name,
    enabled = EXCLUDED.enabled;

-- 为测试用户创建交易所配置
-- 注意：exchanges表使用复合主键 (id, user_id, api_key_name)
INSERT INTO exchanges (id, user_id, api_key_name, name, type, enabled, testnet, api_key, secret_key, created_at)
VALUES (
    'test-binance',
    'test-user-local',
    'default',
    'Binance Test',
    'binance',
    true,
    true,
    'test-api-key',
    'test-secret-key',
    CURRENT_TIMESTAMP
)
ON CONFLICT (id, user_id, api_key_name) DO UPDATE SET
    name = EXCLUDED.name,
    enabled = EXCLUDED.enabled,
    testnet = EXCLUDED.testnet;

-- 显示创建结果
SELECT '✅ 测试用户创建成功！' as status;
SELECT '' as blank;
SELECT '📊 登录信息:' as info;
SELECT '  邮箱: test@local.dev' as email;
SELECT '  密码: test123' as password;
SELECT '  OTP Secret: JBSWY3DPEHPK3PXP' as otp_secret;
SELECT '' as blank2;
SELECT '🔐 OTP设置:' as otp_setup;
SELECT '  1. 打开Google Authenticator' as step1;
SELECT '  2. 添加账户 → 输入设置密钥' as step2;
SELECT '  3. 账户名: test@local.dev' as step3;
SELECT '  4. 密钥: JBSWY3DPEHPK3PXP' as step4;
SELECT '  5. 或使用在线工具: https://totp.danhersam.com/' as step5;
SELECT '' as blank3;
SELECT '💡 提示: 登录时需要输入6位OTP验证码' as tip;
