"use client";

import React, { useState } from "react";
import { motion } from "framer-motion";
import useSWR from "swr";
import { apiClient } from "@/lib/api";
import type { CarbonForecastResponse } from "@/lib/types";
import { 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  Legend
} from "recharts";
import { 
  RefreshCw, 
  TrendingDown, 
  TrendingUp, 
  Activity,
  Database,
  Zap,
  AlertCircle
} from "lucide-react";

interface Region {
  id: string;
  name: string;
  description: string;
}

const REGIONS: Region[] = [
  { id: "", name: "All Regions", description: "Global carbon intensity data" },
  { id: "US-EAST", name: "US East (Virginia)", description: "Coal heavy region" },
  { id: "US-WEST", name: "US West (Oregon)", description: "Hydro powered region" },
  { id: "US-CENTRAL", name: "US Central", description: "Wind energy region" },
  { id: "EU-WEST", name: "EU West", description: "Mixed renewable" },
  { id: "EU-CENTRAL", name: "EU Central", description: "Solar + Wind" },
  { id: "EU-NORTH", name: "EU North", description: "Hydro + Wind" },
  { id: "ASIA-EAST", name: "Asia East", description: "Coal dominant" },
  { id: "ASIA-SOUTH", name: "Asia South", description: "Solar growing" },
  { id: "AU-EAST", name: "Australia East", description: "Coal transitioning" },
];

// Helper function to get intensity color
const getIntensityColor = (intensity: number): string => {
  if (intensity < 150) return "text-green-400";
  if (intensity < 250) return "text-yellow-400";
  if (intensity < 400) return "text-orange-400";
  return "text-red-400";
};

// Custom tooltip component (defined outside to avoid recreation during render)
const CustomTooltip = ({ active, payload }: { active?: boolean; payload?: Array<{ payload: { fullTimestamp: string; intensity: number; region: string } }> }) => {
  if (active && payload && payload.length) {
    const data = payload[0].payload;
    return (
      <div className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-3 shadow-lg">
        <p className="text-sm text-gray-300 mb-1">{data.fullTimestamp}</p>
        <p className={`text-lg font-bold ${getIntensityColor(data.intensity)}`}>
          {data.intensity} gCO₂/kWh
        </p>
        {data.region && (
          <p className="text-xs text-gray-400 mt-1">Region: {data.region}</p>
        )}
      </div>
    );
  }
  return null;
};

const GridIntelligence = () => {
  const [selectedRegion, setSelectedRegion] = useState<string>("");
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  // Fetch carbon forecast data for selected region
  const { 
    data: carbonData, 
    error, 
    isLoading, 
    mutate 
  } = useSWR<CarbonForecastResponse>(
    `/api/carbon-forecast?region=${selectedRegion}`,
    () => apiClient.getCarbonForecast(selectedRegion || undefined),
    {
      refreshInterval: 60000, // Refresh every minute
      revalidateOnFocus: true,
    }
  );

  // Calculate statistics
  const stats = React.useMemo(() => {
    if (!carbonData?.forecasts || carbonData.forecasts.length === 0) {
      return {
        avgIntensity: 0,
        minIntensity: 0,
        maxIntensity: 0,
        dataPoints: 0,
        cleanHours: 0,
        dirtyHours: 0,
      };
    }

    const intensities = carbonData.forecasts.map(f => f.intensity_value);
    const avg = intensities.reduce((a, b) => a + b, 0) / intensities.length;
    const min = Math.min(...intensities);
    const max = Math.max(...intensities);
    const clean = intensities.filter(i => i < 150).length;
    const dirty = intensities.filter(i => i > 400).length;

    return {
      avgIntensity: Math.round(avg),
      minIntensity: Math.round(min),
      maxIntensity: Math.round(max),
      dataPoints: intensities.length,
      cleanHours: clean,
      dirtyHours: dirty,
    };
  }, [carbonData]);

  // Prepare chart data
  const chartData = React.useMemo(() => {
    if (!carbonData?.forecasts) return [];
    
    return carbonData.forecasts.map(forecast => ({
      timestamp: new Date(forecast.timestamp).toLocaleTimeString('en-US', { 
        hour: '2-digit', 
        minute: '2-digit',
        hour12: false 
      }),
      fullTimestamp: new Date(forecast.timestamp).toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      }),
      intensity: forecast.intensity_value,
      region: forecast.region,
    }));
  }, [carbonData]);

  // Get intensity background color
  const getIntensityBg = (intensity: number): string => {
    if (intensity < 150) return "bg-green-500/10 border-green-500/20";
    if (intensity < 250) return "bg-yellow-500/10 border-yellow-500/20";
    if (intensity < 400) return "bg-orange-500/10 border-orange-500/20";
    return "bg-red-500/10 border-red-500/20";
  };

  // Pagination
  const totalPages = Math.ceil(chartData.length / itemsPerPage);
  const paginatedData = chartData.slice(
    (currentPage - 1) * itemsPerPage,
    currentPage * itemsPerPage
  );

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
      className="space-y-6"
    >
      {/* Header */}
      <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4">
        <div>
          <h2 className="text-2xl font-bold text-karbos-light-blue">Grid Intelligence</h2>
          <p className="text-sm text-gray-400 mt-1">
            Real-time carbon intensity forecasting across global regions
          </p>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={selectedRegion}
            onChange={(e) => {
              setSelectedRegion(e.target.value);
              setCurrentPage(1);
            }}
            className="px-4 py-2 bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-md text-black focus:outline-none focus:border-karbos-blue-purple transition-colors"
          >
            {REGIONS.map(region => (
              <option key={region.id} value={region.id}>
                {region.name}
              </option>
            ))}
          </select>
          <button
            onClick={() => mutate()}
            disabled={isLoading}
            className="px-4 py-2 bg-karbos-blue-purple text-white rounded-md hover:bg-opacity-80 transition-colors flex items-center gap-2 disabled:opacity-50"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </div>
      </div>

      {/* Region Description */}
      {selectedRegion && (
        <div className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-4">
          <p className="text-sm text-gray-300">
            <span className="font-semibold text-karbos-light-blue">
              {REGIONS.find(r => r.id === selectedRegion)?.name}:
            </span>{' '}
            {REGIONS.find(r => r.id === selectedRegion)?.description}
          </p>
        </div>
      )}

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-4"
        >
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-400">Average Intensity</span>
            <Activity className="w-4 h-4 text-karbos-blue-purple" />
          </div>
          <p className={`text-2xl font-bold ${getIntensityColor(stats.avgIntensity)}`}>
            {stats.avgIntensity}
          </p>
          <p className="text-xs text-gray-500 mt-1">gCO₂/kWh</p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-4"
        >
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-400">Lowest Intensity</span>
            <TrendingDown className="w-4 h-4 text-green-400" />
          </div>
          <p className="text-2xl font-bold text-green-400">{stats.minIntensity}</p>
          <p className="text-xs text-gray-500 mt-1">Optimal scheduling window</p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-4"
        >
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-400">Peak Intensity</span>
            <TrendingUp className="w-4 h-4 text-red-400" />
          </div>
          <p className="text-2xl font-bold text-red-400">{stats.maxIntensity}</p>
          <p className="text-xs text-gray-500 mt-1">Avoid scheduling here</p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
          className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-4"
        >
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-400">Data Points</span>
            <Database className="w-4 h-4 text-karbos-blue-purple" />
          </div>
          <p className="text-2xl font-bold text-white">{stats.dataPoints}</p>
          <p className="text-xs text-gray-500 mt-1">
            {stats.cleanHours} clean / {stats.dirtyHours} dirty
          </p>
        </motion.div>
      </div>

      {/* Error State */}
      {error && (
        <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 flex items-start gap-3">
          <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
          <div>
            <p className="text-red-400 font-semibold">Failed to load carbon data</p>
            <p className="text-sm text-gray-400 mt-1">
              Please check your connection and try again.
            </p>
          </div>
        </div>
      )}

      {/* Loading State */}
      {isLoading && !carbonData && (
        <div className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-12 flex flex-col items-center justify-center">
          <RefreshCw className="w-8 h-8 text-karbos-blue-purple animate-spin mb-4" />
          <p className="text-gray-400">Loading carbon intensity data...</p>
        </div>
      )}

      {/* No Data State */}
      {!isLoading && !error && chartData.length === 0 && (
        <div className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-12 flex flex-col items-center justify-center">
          <Database className="w-12 h-12 text-gray-600 mb-4" />
          <p className="text-gray-400 text-lg mb-2">No carbon forecast data available</p>
          <p className="text-sm text-gray-500">
            Try selecting a different region or refresh the data
          </p>
        </div>
      )}

      {/* Chart Section */}
      {chartData.length > 0 && (
        <>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
            className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-6"
          >
            <div className="mb-4 flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold text-white">Carbon Intensity Forecast</h3>
                <p className="text-sm text-gray-400 mt-1">
                  Time-series visualization of grid carbon intensity
                </p>
              </div>
              <Zap className="w-6 h-6 text-karbos-blue-purple" />
            </div>

            <ResponsiveContainer width="100%" height={400}>
              <AreaChart data={chartData}>
                <defs>
                  <linearGradient id="intensityGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#ef4444" stopOpacity={0.8} />
                    <stop offset="30%" stopColor="#f59e0b" stopOpacity={0.6} />
                    <stop offset="60%" stopColor="#eab308" stopOpacity={0.4} />
                    <stop offset="100%" stopColor="#22c55e" stopOpacity={0.3} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                <XAxis 
                  dataKey="timestamp" 
                  stroke="#9ca3af"
                  tick={{ fill: '#9ca3af', fontSize: 12 }}
                />
                <YAxis 
                  stroke="#9ca3af"
                  tick={{ fill: '#9ca3af', fontSize: 12 }}
                  label={{ 
                    value: 'gCO₂/kWh', 
                    angle: -90, 
                    position: 'insideLeft',
                    style: { fill: '#9ca3af' }
                  }}
                />
                <Tooltip content={<CustomTooltip />} />
                <Legend 
                  wrapperStyle={{ color: '#9ca3af' }}
                  formatter={() => 'Carbon Intensity'}
                />
                <Area
                  type="monotone"
                  dataKey="intensity"
                  stroke="#8b5cf6"
                  strokeWidth={2}
                  fill="url(#intensityGradient)"
                  animationDuration={1000}
                />
              </AreaChart>
            </ResponsiveContainer>
          </motion.div>

          {/* Data Table */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6 }}
            className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg overflow-hidden"
          >
            <div className="p-6 border-b border-karbos-blue-purple/20">
              <h3 className="text-lg font-semibold text-white">Raw Forecast Data</h3>
              <p className="text-sm text-gray-400 mt-1">
                Auditability table showing all forecast values
              </p>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-karbos-dark-gray/50 border-b border-karbos-blue-purple/20">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                      Timestamp
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                      Region
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                      Intensity (gCO₂/kWh)
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                      Status
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-karbos-blue-purple/10">
                  {paginatedData.map((data, index) => (
                    <tr 
                      key={index}
                      className="hover:bg-karbos-blue-purple/5 transition-colors"
                    >
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                        {data.fullTimestamp}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                        {data.region}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`text-sm font-semibold ${getIntensityColor(data.intensity)}`}>
                          {data.intensity}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`px-2 py-1 rounded-full text-xs font-medium border ${getIntensityBg(data.intensity)} ${getIntensityColor(data.intensity)}`}>
                          {data.intensity < 150 ? '🟢 Clean' : 
                           data.intensity < 250 ? '🟡 Moderate' :
                           data.intensity < 400 ? '🟠 High' : '🔴 Very High'}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="px-6 py-4 border-t border-karbos-blue-purple/20 flex flex-col sm:flex-row items-center justify-between gap-4">
                <div className="text-sm text-gray-400">
                  Showing {((currentPage - 1) * itemsPerPage) + 1} to {Math.min(currentPage * itemsPerPage, chartData.length)} of {chartData.length} entries
                </div>
                <div className="flex gap-2 items-center">
                  <button
                    onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                    disabled={currentPage === 1}
                    className="px-3 py-1 bg-karbos-blue-purple/20 text-white rounded-md hover:bg-karbos-blue-purple/30 transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-sm"
                  >
                    Previous
                  </button>
                  
                  <div className="flex gap-1 items-center">
                    {/* First page */}
                    {currentPage > 3 && (
                      <>
                        <button
                          onClick={() => setCurrentPage(1)}
                          className="px-3 py-1 rounded-md text-sm bg-karbos-blue-purple/20 text-gray-300 hover:bg-karbos-blue-purple/30 transition-colors"
                        >
                          1
                        </button>
                        <span className="text-gray-500 px-1">...</span>
                      </>
                    )}
                    
                    {/* Page numbers around current page */}
                    {Array.from({ length: totalPages }, (_, i) => i + 1)
                      .filter(page => {
                        // Show first 3 pages, last 3 pages, and 2 pages around current page
                        return (
                          page === currentPage ||
                          (page >= currentPage - 1 && page <= currentPage + 1) ||
                          (currentPage <= 2 && page <= 3) ||
                          (currentPage >= totalPages - 1 && page >= totalPages - 2)
                        );
                      })
                      .map(page => (
                        <button
                          key={page}
                          onClick={() => setCurrentPage(page)}
                          className={`px-3 py-1 rounded-md text-sm transition-colors ${
                            currentPage === page
                              ? 'bg-karbos-blue-purple text-white'
                              : 'bg-karbos-blue-purple/20 text-gray-300 hover:bg-karbos-blue-purple/30'
                          }`}
                        >
                          {page}
                        </button>
                      ))}
                    
                    {/* Last page */}
                    {currentPage < totalPages - 2 && (
                      <>
                        <span className="text-gray-500 px-1">...</span>
                        <button
                          onClick={() => setCurrentPage(totalPages)}
                          className="px-3 py-1 rounded-md text-sm bg-karbos-blue-purple/20 text-gray-300 hover:bg-karbos-blue-purple/30 transition-colors"
                        >
                          {totalPages}
                        </button>
                      </>
                    )}
                  </div>
                  
                  <button
                    onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                    disabled={currentPage === totalPages}
                    className="px-3 py-1 bg-karbos-blue-purple/20 text-white rounded-md hover:bg-karbos-blue-purple/30 transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-sm"
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </motion.div>

          {/* Legend Explanation */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.7 }}
            className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-4"
          >
            <h4 className="text-sm font-semibold text-white mb-3">Intensity Classification</h4>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-green-400"></div>
                <span className="text-xs text-gray-400">&lt; 150: Clean</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-yellow-400"></div>
                <span className="text-xs text-gray-400">150-250: Moderate</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-orange-400"></div>
                <span className="text-xs text-gray-400">250-400: High</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-red-400"></div>
                <span className="text-xs text-gray-400">&gt; 400: Very High</span>
              </div>
            </div>
          </motion.div>
        </>
      )}
    </motion.div>
  );
};

export default GridIntelligence;
