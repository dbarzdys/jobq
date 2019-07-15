CREATE OR REPLACE FUNCTION jobq_notify_task_created() RETURNS TRIGGER AS $$
DECLARE 
    notification jsonb;
BEGIN
    notification = json_build_object(
        'job_name', NEW.job_name,
        'timeout', NEW.timeout,
        'start_at', NEW.start_at
    );
    PERFORM pg_notify('jobq_task_created', notification::text);
    RETURN NULL; 
END;
$$ LANGUAGE plpgsql;
