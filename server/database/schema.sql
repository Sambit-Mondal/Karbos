-- Karbos Database Schema
-- PostgreSQL with Supabase

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Job Status Enum
CREATE TYPE job_status AS ENUM (
    'PENDING',
    'DELAYED',
    'RUNNING',
    'COMPLETED',
    'FAILED'
);

-- Jobs Table
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    docker_image VARCHAR(500) NOT NULL,
    command TEXT,
    status job_status NOT NULL DEFAULT 'PENDING',
    scheduled_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    estimated_duration INTEGER, -- in seconds
    region VARCHAR(50),
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Constraints
    CONSTRAINT jobs_deadline_future CHECK (deadline > created_at)
);

-- Execution Logs Table
CREATE TABLE IF NOT EXISTS execution_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    output TEXT,
    error_output TEXT,
    exit_code INTEGER,
    duration INTEGER, -- in seconds
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    worker_node_id VARCHAR(100),
    
    CONSTRAINT execution_logs_job_fk FOREIGN KEY (job_id) REFERENCES jobs(id)
);

-- Carbon Cache Table
CREATE TABLE IF NOT EXISTS carbon_cache (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    region VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    intensity_value DECIMAL(10, 2) NOT NULL, -- gCO2/kWh
    forecast_window INTEGER, -- hours ahead
    source VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Composite unique constraint for region + timestamp
    CONSTRAINT carbon_cache_unique UNIQUE (region, timestamp, forecast_window)
);

-- Indexes for performance
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_user_id ON jobs(user_id);
CREATE INDEX idx_jobs_scheduled_time ON jobs(scheduled_time);
CREATE INDEX idx_jobs_created_at ON jobs(created_at DESC);
CREATE INDEX idx_jobs_deadline ON jobs(deadline);

CREATE INDEX idx_execution_logs_job_id ON execution_logs(job_id);
CREATE INDEX idx_execution_logs_started_at ON execution_logs(started_at DESC);

CREATE INDEX idx_carbon_cache_region_timestamp ON carbon_cache(region, timestamp DESC);
CREATE INDEX idx_carbon_cache_region ON carbon_cache(region);

-- Function to update job status timestamp
CREATE OR REPLACE FUNCTION update_job_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'RUNNING' AND OLD.status != 'RUNNING' THEN
        NEW.started_at = NOW();
    END IF;
    
    IF (NEW.status = 'COMPLETED' OR NEW.status = 'FAILED') AND 
       (OLD.status != 'COMPLETED' AND OLD.status != 'FAILED') THEN
        NEW.completed_at = NOW();
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update timestamps
CREATE TRIGGER job_status_timestamp_trigger
    BEFORE UPDATE ON jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_job_timestamp();

-- Comments for documentation
COMMENT ON TABLE jobs IS 'Stores all job submissions with scheduling and status information';
COMMENT ON TABLE execution_logs IS 'Stores execution logs and results for each job run';
COMMENT ON TABLE carbon_cache IS 'Caches carbon intensity forecasts for different regions';

COMMENT ON COLUMN jobs.scheduled_time IS 'The optimized time when the job should be executed';
COMMENT ON COLUMN jobs.deadline IS 'The SLA deadline by which the job must complete';
COMMENT ON COLUMN jobs.estimated_duration IS 'Estimated job duration in seconds';
COMMENT ON COLUMN carbon_cache.intensity_value IS 'Carbon intensity in grams of CO2 per kilowatt-hour';
