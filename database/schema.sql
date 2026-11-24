-- SIPASDOS: skema database PostgreSQL
-- Jalankan sekali: psql -U postgres -d sipasdos -f database/schema.sql

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(150) NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    role          VARCHAR(20)  NOT NULL CHECK (role IN ('admin', 'dosen', 'mahasiswa')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS study_programs (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS students (
    id               SERIAL PRIMARY KEY,
    user_id          INT          NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    nim              VARCHAR(20)  NOT NULL UNIQUE,
    study_program_id INT          NOT NULL REFERENCES study_programs(id),
    semester         INT          NOT NULL CHECK (semester BETWEEN 1 AND 14),
    ipk              NUMERIC(3,2) NOT NULL CHECK (ipk BETWEEN 0 AND 4)
);

CREATE TABLE IF NOT EXISTS lecturers (
    id               SERIAL PRIMARY KEY,
    user_id          INT         NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    nidn             VARCHAR(20) NOT NULL UNIQUE,
    study_program_id INT         NOT NULL REFERENCES study_programs(id)
);

CREATE TABLE IF NOT EXISTS courses (
    id               SERIAL PRIMARY KEY,
    code             VARCHAR(20)  NOT NULL UNIQUE,
    name             VARCHAR(100) NOT NULL,
    study_program_id INT          NOT NULL REFERENCES study_programs(id)
);

CREATE TABLE IF NOT EXISTS recruitment_periods (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    start_date DATE         NOT NULL,
    end_date   DATE         NOT NULL,
    status     VARCHAR(10)  NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'open', 'closed')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CHECK (end_date >= start_date)
);

CREATE TABLE IF NOT EXISTS vacancies (
    id          SERIAL PRIMARY KEY,
    period_id   INT          NOT NULL REFERENCES recruitment_periods(id),
    course_id   INT          NOT NULL REFERENCES courses(id),
    lecturer_id INT          NOT NULL REFERENCES lecturers(id),
    quota       INT          NOT NULL CHECK (quota > 0),
    min_ipk     NUMERIC(3,2) NOT NULL DEFAULT 0 CHECK (min_ipk BETWEEN 0 AND 4),
    description TEXT         NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS applications (
    id         SERIAL PRIMARY KEY,
    vacancy_id INT         NOT NULL REFERENCES vacancies(id),
    student_id INT         NOT NULL REFERENCES students(id),
    motivation TEXT        NOT NULL DEFAULT '',
    status     VARCHAR(15) NOT NULL DEFAULT 'submitted'
               CHECK (status IN ('submitted', 'verified', 'interview', 'accepted', 'rejected')),
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (vacancy_id, student_id)
);

CREATE TABLE IF NOT EXISTS application_documents (
    id             SERIAL PRIMARY KEY,
    application_id INT         NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    type           VARCHAR(20) NOT NULL CHECK (type IN ('cv', 'transkrip', 'lainnya')),
    file_url       TEXT        NOT NULL
);

CREATE TABLE IF NOT EXISTS interviews (
    id             SERIAL PRIMARY KEY,
    application_id INT          NOT NULL UNIQUE REFERENCES applications(id) ON DELETE CASCADE,
    scheduled_at   TIMESTAMPTZ  NOT NULL,
    location       VARCHAR(150) NOT NULL,
    notes          TEXT         NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS assessments (
    id             SERIAL PRIMARY KEY,
    application_id INT          NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    lecturer_id    INT          NOT NULL REFERENCES lecturers(id),
    score          NUMERIC(5,2) NOT NULL CHECK (score BETWEEN 0 AND 100),
    comment        TEXT         NOT NULL DEFAULT '',
    UNIQUE (application_id, lecturer_id)
);

CREATE TABLE IF NOT EXISTS notifications (
    id         SERIAL PRIMARY KEY,
    user_id    INT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message    TEXT        NOT NULL,
    is_read    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
