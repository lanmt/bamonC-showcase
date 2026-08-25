//已改为使用grom 迁移，该sql（可做参考）已弃用

-- 创建 schema
CREATE SCHEMA IF NOT EXISTS bc;

-- 删除旧表（注意外键顺序）
DROP TABLE IF EXISTS bc.courts CASCADE;
DROP TABLE IF EXISTS bc.buddies CASCADE;
DROP TABLE IF EXISTS bc.users CASCADE;

-- 用户表
CREATE TABLE bc.users (
                          id BIGSERIAL PRIMARY KEY,
                          username TEXT UNIQUE NOT NULL,
                          point TEXT,
                          auth TEXT,
                          auth_updated_at TIMESTAMP WITH TIME ZONE,
                          relay_until TIMESTAMP WITH TIME ZONE,
                          relay_enabled BOOLEAN DEFAULT FALSE,
                          enabled BOOLEAN DEFAULT TRUE,
                          buddy BIGINT,
                          updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 好友表
CREATE TABLE bc.buddies (
                            id BIGSERIAL PRIMARY KEY,
                            user_id BIGINT NOT NULL REFERENCES bc.users(id) ON DELETE CASCADE,
                            buddy_id BIGINT NOT NULL,
                            buddy_name TEXT
);

-- 球场表
CREATE TABLE bc.courts (
                           id BIGSERIAL PRIMARY KEY,
                           user_id BIGINT NOT NULL REFERENCES bc.users(id) ON DELETE CASCADE,
                           venue_site_id BIGINT NOT NULL,
                           court_id BIGINT NOT NULL,
                           time1_id BIGINT NOT NULL,
                           time2_id BIGINT
);
