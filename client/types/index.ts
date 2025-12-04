// Common types used across the application

export interface Job {
  id: string;
  image: string;
  status: "RUNNING" | "DELAYED" | "COMPLETED" | "FAILED";
  carbonImpact?: number;
  slaDeadline: string;
  submittedAt: string;
  startTime?: string;
  estimatedStart?: string;
  logs?: string;
}

export interface WorkerNode {
  id: string;
  name: string;
  status: "online" | "offline" | "busy";
  lastHeartbeat: string;
  jobsProcessed: number;
  cpuUsage: number;
  memoryUsage: number;
}

export interface Region {
  id: string;
  name: string;
  intensity: number;
  status: "green" | "yellow" | "red";
}

export interface ForecastData {
  timeWindow: string;
  wind: number;
  solar: number;
  gas: number;
  coal: number;
  nuclear: number;
  intensity: number;
}

export interface QueueData {
  ready: number;
  delayed: number;
  active: number;
  completed: number;
  failed: number;
}

export interface KPIData {
  totalCO2Saved: number;
  activeJobs: number;
  pendingOptimized: number;
  currentIntensity: number;
  trend: "rising" | "falling";
}

export interface SystemStatus {
  scheduler: boolean;
  redis: boolean;
}

export interface ChartDataPoint {
  hour: number;
  intensity: number;
  isOptimal: boolean;
}

export interface JobSubmission {
  dockerImage: string;
  command: string;
  estimatedDuration: number;
  slaDeadline: number;
  simulationMode: boolean;
}

export interface SimulationResult {
  immediateExecution: {
    startTime: string;
    carbonIntensity: number;
    estimatedCO2: number;
  };
  optimizedExecution: {
    startTime: string;
    carbonIntensity: number;
    estimatedCO2: number;
    delaySavings: number;
  };
}
