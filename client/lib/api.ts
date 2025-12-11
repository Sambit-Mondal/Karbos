import axios from 'axios';
import type {
  Job,
  ExecutionLog,
  CarbonCacheEntry,
  SubmitJobRequest,
  SubmitJobResponse,
  HealthResponse,
  CarbonForecastResponse,
  SystemHealthResponse
} from './types';

// Configure axios instance
const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// API Methods
export const apiClient = {
  // Health check
  health: async (): Promise<HealthResponse> => {
    const { data } = await api.get('/health');
    return data;
  },

  // Jobs
  getJobs: async (): Promise<Job[]> => {
    const { data } = await api.get('/api/jobs');
    return data;
  },

  getJob: async (jobId: string): Promise<Job> => {
    const { data } = await api.get(`/api/jobs/${jobId}`);
    return data;
  },

  submitJob: async (request: SubmitJobRequest, dryRun: boolean = false): Promise<SubmitJobResponse> => {
    const params = dryRun ? { dry_run: 'true' } : {};
    const { data } = await api.post('/api/submit', request, { params });
    return data;
  },

  // Execution Logs
  getExecutionLog: async (jobId: string): Promise<ExecutionLog> => {
    const { data } = await api.get(`/api/jobs/${jobId}/logs`);
    return data;
  },

  // Carbon Forecast
  getCarbonForecast: async (region?: string): Promise<CarbonForecastResponse> => {
    const params = region ? { region } : {};
    const { data } = await api.get('/api/carbon-forecast', { params });
    return data;
  },

  // Carbon Cache (all entries)
  getCarbonCache: async (): Promise<CarbonCacheEntry[]> => {
    const { data } = await api.get('/api/carbon-cache');
    return data;
  },

  // Metrics endpoint (Prometheus)
  getMetrics: async (): Promise<string> => {
    const { data } = await api.get('/metrics', {
      headers: { 'Accept': 'text/plain' }
    });
    return data;
  },

  // System Health
  getSystemHealth: async (): Promise<SystemHealthResponse> => {
    const { data } = await api.get('/api/system/health');
    return data;
  },
};

export default api;
