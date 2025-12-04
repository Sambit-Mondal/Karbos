/**
 * Utility functions for the Karbos dashboard
 */

/**
 * Format a timestamp to relative time (e.g., "2 min ago")
 */
export function formatRelativeTime(
  timestamp: string,
  now: Date = new Date(),
): string {
  const then = new Date(timestamp);
  const diff = Math.floor((now.getTime() - then.getTime()) / 1000);

  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

/**
 * Format a duration in minutes to human-readable format
 */
export function formatDuration(minutes: number): string {
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`;
}

/**
 * Calculate carbon savings percentage
 */
export function calculateSavings(baseline: number, optimized: number): number {
  return Math.round(((baseline - optimized) / baseline) * 100);
}

/**
 * Get intensity level classification
 */
export function getIntensityLevel(intensity: number): {
  label: string;
  color: string;
} {
  if (intensity < 350) {
    return { label: "Low", color: "text-green-400" };
  }
  if (intensity < 450) {
    return { label: "Medium", color: "text-yellow-400" };
  }
  return { label: "High", color: "text-red-400" };
}

/**
 * Format large numbers with commas
 */
export function formatNumber(num: number): string {
  return num.toLocaleString();
}

/**
 * Format CO2 amount with appropriate units
 */
export function formatCO2(grams: number): string {
  if (grams < 1000) return `${Math.round(grams)}g`;
  return `${(grams / 1000).toFixed(2)}kg`;
}

/**
 * Get status color class
 */
export function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    RUNNING: "bg-green-500",
    DELAYED: "bg-yellow-500",
    COMPLETED: "bg-blue-500",
    FAILED: "bg-red-500",
    online: "bg-green-500",
    busy: "bg-yellow-500",
    offline: "bg-red-500",
    green: "bg-green-500",
    yellow: "bg-yellow-500",
    red: "bg-red-500",
  };
  return colors[status] || "bg-gray-500";
}

/**
 * Calculate time until a future date
 */
export function getTimeUntil(
  futureDate: string,
  now: Date = new Date(),
): string {
  const future = new Date(futureDate);
  const diff = future.getTime() - now.getTime();
  const minutes = Math.floor(diff / 60000);

  if (minutes < 0) return "Started";
  if (minutes < 60) return `Starts in ${minutes}m`;

  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  return mins > 0 ? `Starts in ${hours}h ${mins}m` : `Starts in ${hours}h`;
}

/**
 * Generate mock data for development
 */
export function generateMockChartData(hours: number = 24): Array<{
  hour: number;
  intensity: number;
  isOptimal: boolean;
}> {
  return Array.from({ length: hours }, (_, i) => ({
    hour: i,
    intensity: 300 + Math.sin(i / 3) * 100 + Math.random() * 50,
    isOptimal: (i >= 2 && i <= 6) || (i >= 14 && i <= 18),
  }));
}

/**
 * Validate Docker image name format
 */
export function isValidDockerImage(image: string): boolean {
  // Basic validation for Docker image format
  const regex =
    /^[a-z0-9]+([\.\-_][a-z0-9]+)*\/[a-z0-9]+([\.\-_][a-z0-9]+)*(:[a-z0-9]+([\.\-_][a-z0-9]+)*)?$/i;
  return (
    regex.test(image) ||
    /^[a-z0-9]+([\.\-_][a-z0-9]+)*(:[a-z0-9]+([\.\-_][a-z0-9]+)*)?$/i.test(
      image,
    )
  );
}

/**
 * Calculate percentage for progress bars
 */
export function calculatePercentage(value: number, max: number): number {
  return Math.min(Math.round((value / max) * 100), 100);
}
