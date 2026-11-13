-- Инициализация расширений PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 1. Организационная структура
CREATE TABLE IF NOT EXISTS schools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    region VARCHAR(100) DEFAULT 'Республика Башкортостан',
    city VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS classes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    grade_number INT NOT NULL,
    letter VARCHAR(5) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Пользователи
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tg_id BIGINT UNIQUE,
    tg_username VARCHAR(100),
    email VARCHAR(255) UNIQUE,
    hashed_password VARCHAR(255),
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'student',
    school_id UUID REFERENCES schools(id) ON DELETE SET NULL,
    class_id UUID REFERENCES classes(id) ON DELETE SET NULL,
    current_streak INT DEFAULT 0,
    max_streak INT DEFAULT 0,
    last_active_date DATE,
    total_xp INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Источники материалов
CREATE TABLE IF NOT EXISTS sources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    external_url VARCHAR(1000),
    hf_dataset_name VARCHAR(255),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Фрагменты предложений
CREATE TABLE IF NOT EXISTS fragments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_id UUID NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    text_ba TEXT NOT NULL,
    audio_url VARCHAR(1000),
    audio_start_ms INT,
    audio_end_ms INT,
    chapter_info VARCHAR(255),
    paragraph_idx INT,
    sentence_idx INT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Языковые пары и консенсус
CREATE TABLE IF NOT EXISTS fragment_translations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fragment_id UUID NOT NULL REFERENCES fragments(id) ON DELETE CASCADE,
    lang_pair VARCHAR(10) NOT NULL DEFAULT 'ba-ru',
    initial_auto_translation TEXT NOT NULL,
    has_honeypot BOOLEAN DEFAULT FALSE,
    honeypot_variant TEXT,
    confirmed_translation TEXT,
    consensus_score DOUBLE PRECISION DEFAULT 0.0,
    confirmed_count INT DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 6. Ответы пользователей
CREATE TABLE IF NOT EXISTS user_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fragment_translation_id UUID NOT NULL REFERENCES fragment_translations(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    submitted_text TEXT NOT NULL,
    time_spent_seconds INT DEFAULT 0,
    detected_honeypot BOOLEAN DEFAULT FALSE,
    xp_awarded INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT uq_user_fragment UNIQUE(user_id, fragment_translation_id)
);

-- 7. Итоговая согласованная база (Требование раздела 5 ТЗ)
CREATE TABLE IF NOT EXISTS confirmed_translations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fragment_id UUID NOT NULL,
    text_ba TEXT NOT NULL,
    lang_pair VARCHAR(10) NOT NULL,
    translation TEXT NOT NULL,
    source_coord VARCHAR(500) NOT NULL,
    confirmed_count INT DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'confirmed',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_fragments_source ON fragments(source_id);
CREATE INDEX IF NOT EXISTS idx_frag_trans_status ON fragment_translations(status, lang_pair);
CREATE INDEX IF NOT EXISTS idx_submissions_user ON user_submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_users_xp ON users(total_xp DESC);
CREATE INDEX IF NOT EXISTS idx_users_tg_id ON users(tg_id);
