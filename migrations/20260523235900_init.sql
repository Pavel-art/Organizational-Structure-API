-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS departments (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    parent_id   INT NULL REFERENCES departments(id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS employees (
    id             SERIAL PRIMARY KEY,
    department_id  INT NOT NULL REFERENCES departments(id) ON UPDATE CASCADE ON DELETE CASCADE,
    full_name      VARCHAR(200) NOT NULL,
    position       VARCHAR(200) NOT NULL,
    hired_at       DATE NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- uniqueness within same parent, including root (parent_id IS NULL)
CREATE UNIQUE INDEX IF NOT EXISTS uq_departments_parent_name_not_null
    ON departments(parent_id, name)
    WHERE parent_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_departments_root_name
    ON departments(name)
    WHERE parent_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_departments_parent_id ON departments(parent_id);
CREATE INDEX IF NOT EXISTS idx_employees_department_id ON employees(department_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS departments;
-- +goose StatementEnd

