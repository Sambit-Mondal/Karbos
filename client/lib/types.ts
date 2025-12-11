// API Response Types matching backend models

export type JobStatus = 'PENDING' | 'DELAYED' | 'RUNNING' | 'COMPLETED' | 'FAILED';

export interface Job {
  id: string;
  user_id: string;
  docker_image: string;
  command?: string;
  status: JobStatus;
  scheduled_time?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  deadline: string;
  estimated_duration?: number; // seconds
  region?: string;
  metadata?: string;
}

export interface ExecutionLog {
  id: string;
  job_id: string;
  output?: string;
  error_message?: string;
  exit_code?: number;
  duration?: number; // seconds
  started_at: string;
  completed_at?: string;
  worker_node_id?: string;
  created_at: string;
}

export interface CarbonCacheEntry {
  id: string;
  region: string;
  timestamp: string;
  intensity_value: number; // gCO2/kWh
  forecast_window?: number;
  source?: string;
  created_at: string;
}

export interface SubmitJobRequest {
  user_id: string;
  docker_image: string;
  command?: string[];
  deadline: string; // ISO 8601
  estimated_duration?: number; // seconds
  region?: string;
}

export interface SubmitJobResponse {
  job_id: string;
  status: JobStatus;
  created_at: string;
  scheduled_time: string;
  immediate: boolean;
  expected_intensity?: number;
  carbon_savings?: number;
}

export interface HealthResponse {
  status: string;
  timestamp: string;
  database: string;
  redis: string;
  version: string;
}

export interface CarbonForecastResponse {
  region: string;
  forecasts: Array<{
    region: string;
    timestamp: string;
    intensity_value: number;
    unit: string;
  }>;
  current_intensity?: number;
  optimal_time?: string;
}

export interface SystemHealthResponse {
  active_workers: number;
  worker_ids: string[];
  queue_depth_immediate: number;
  queue_depth_delayed: number;
  redis_latency_ms: number;
  timestamp: string;
}
