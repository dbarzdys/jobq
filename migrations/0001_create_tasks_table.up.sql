CREATE TABLE IF NOT EXISTS jobq_tasks (
    id BIGSERIAL,
    job_name varchar(100) NOT NULL,
    body jsonb NOT NULL,
    retries int NOT NULL,
    timeout timestamp,
    start_at timestamp,
    PRIMARY KEY(id)
);
