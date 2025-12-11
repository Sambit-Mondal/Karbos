"use client";

import React from "react";
import { motion } from "framer-motion";
import useSWR from "swr";
import { apiClient } from "@/lib/api";
import type { Job, CarbonForecastResponse } from "@/lib/types";
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Area, AreaChart } from "recharts";
import { RefreshCw, TrendingDown, TrendingUp, Activity, Clock, CheckCircle, Loader } from "lucide-react";

const Overview = () => {
  // Fetch jobs data
  const { data: jobs, error: jobsError, isLoading: jobsLoading, mutate: mutateJobs } = useSWR<Job[]>(
    '/api/jobs',
    () => apiClient.getJobs(),
    {
      refreshInterval: 5000,
      revalidateOnFocus: true,
    }
  );

  // Fetch carbon forecast data
  const { data: carbonData, error: carbonError, isLoading: carbonLoading, mutate: mutateCarbonData } = useSWR<CarbonForecastResponse>(
    '/api/carbon-forecast',
    () => apiClient.getCarbonForecast(),
    {
      refreshInterval: 30000, // Refresh every 30 seconds
      revalidateOnFocus: true,
    }
  );

  const handleRefreshAll = () => {
    mutateJobs();
    mutateCarbonData();
  };

  // Calculate KPIs from jobs data
  const kpiData = {
    activeJobs: jobs?.filter(j => j.status === 'RUNNING').length || 0,
    pendingOptimized: jobs?.filter(j => j.status === 'DELAYED' || j.status === 'PENDING').length || 0,
    completedJobs: jobs?.filter(j => j.status === 'COMPLETED').length || 0,
    failedJobs: jobs?.filter(j => j.status === 'FAILED').length || 0,
    currentIntensity: carbonData?.current_intensity || 0,
    totalCO2Saved: (jobs?.filter(j => j.status === 'COMPLETED').length || 0) * 100, // Estimate: 100g per job
  };

  // Prepare chart data from carbon forecast
  const chartData = carbonData?.forecasts.map(forecast => ({
    timestamp: new Date(forecast.timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
    intensity: forecast.intensity_value,
    region: forecast.region,
  })) || [];

  // Group by region if multiple regions exist
  const regions = Array.from(new Set(chartData.map(d => d.region)));

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
      className="space-y-6"
    >
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-karbos-light-blue">System Overview</h2>
          <p className="text-sm text-gray-400 mt-1">
            Carbon-aware job scheduling dashboard
          </p>
        </div>
        <button
          onClick={handleRefreshAll}
          disabled={jobsLoading || carbonLoading}
          className="px-4 py-2 bg-karbos-blue-purple text-white rounded-md hover:bg-opacity-80 transition-colors flex items-center gap-2 disabled:opacity-50"
        >
          <RefreshCw className={`w-4 h-4 ${(jobsLoading || carbonLoading) ? 'animate-spin' : ''}`} />
          Refresh
        </button>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.1 }}
          whileHover={{ scale: 1.02, y: -4 }}
          className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple hover:border-karbos-lavender transition-colors"
        >
          <div className="flex items-center justify-between">
            <h3 className="text-karbos-lavender text-sm font-medium">Total CO₂ Saved</h3>
            <TrendingDown className="w-5 h-5 text-green-400" />
          </div>
          <p className="text-3xl font-bold text-karbos-light-blue mt-2">
            {(kpiData.totalCO2Saved / 1000).toFixed(2)} kg
          </p>
          <p className="text-green-400 text-sm mt-1">Est. from {kpiData.completedJobs} completed jobs</p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.2 }}
          whileHover={{ scale: 1.02, y: -4 }}
          className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple hover:border-karbos-lavender transition-colors"
        >
          <div className="flex items-center justify-between">
            <h3 className="text-karbos-lavender text-sm font-medium">Active Jobs</h3>
            <Activity className="w-5 h-5 text-blue-400" />
          </div>
          <p className="text-3xl font-bold text-karbos-light-blue mt-2">
            {kpiData.activeJobs}
          </p>
          <p className="text-karbos-lavender text-sm mt-1">Running now</p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.3 }}
          whileHover={{ scale: 1.02, y: -4 }}
          className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple hover:border-karbos-lavender transition-colors"
        >
          <div className="flex items-center justify-between">
            <h3 className="text-karbos-lavender text-sm font-medium">Pending (Optimized)</h3>
            <Clock className="w-5 h-5 text-yellow-400" />
          </div>
          <p className="text-3xl font-bold text-karbos-light-blue mt-2">
            {kpiData.pendingOptimized}
          </p>
          <p className="text-yellow-400 text-sm mt-1">Waiting for green window</p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.4 }}
          whileHover={{ scale: 1.02, y: -4 }}
          className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple hover:border-karbos-lavender transition-colors"
        >
          <div className="flex items-center justify-between">
            <h3 className="text-karbos-lavender text-sm font-medium">Grid Intensity</h3>
            <TrendingUp className="w-5 h-5 text-orange-400" />
          </div>
          <p className="text-3xl font-bold text-karbos-light-blue mt-2">
            {kpiData.currentIntensity > 0 ? Math.round(kpiData.currentIntensity) : '—'}
          </p>
          <p className="text-karbos-lavender text-sm mt-1">gCO₂/kWh</p>
        </motion.div>
      </div>

      {/* Carbon Intensity Chart */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.5 }}
        className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple"
      >
        <div className="flex justify-between items-center mb-4">
          <div>
            <h3 className="text-xl font-bold text-karbos-light-blue">Carbon Intensity Forecast</h3>
            <p className="text-sm text-gray-400 mt-1">
              {carbonLoading ? 'Loading...' : carbonData ? `${chartData.length} data points` : 'No data'}
            </p>
          </div>
          {carbonData?.optimal_time && (
            <div className="text-right">
              <p className="text-xs text-gray-400">Optimal Time</p>
              <p className="text-sm text-green-400 font-semibold">
                {new Date(carbonData.optimal_time).toLocaleTimeString()}
              </p>
            </div>
          )}
        </div>

        {carbonError && (
          <div className="bg-red-900/20 border border-red-500/30 rounded-lg p-4 text-center">
            <p className="text-red-400">Failed to load carbon data: {carbonError.message}</p>
          </div>
        )}

        {carbonLoading && !carbonData && (
          <div className="flex justify-center items-center py-12">
            <Loader className="w-8 h-8 text-karbos-blue-purple animate-spin" />
          </div>
        )}

        {carbonData && chartData.length === 0 && !carbonLoading && (
          <div className="bg-yellow-900/20 border border-yellow-500/30 rounded-lg p-4 text-center">
            <p className="text-yellow-400">No carbon forecast data available</p>
          </div>
        )}

        {chartData.length > 0 && (
          <div className="mt-4">
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={chartData}>
                <defs>
                  <linearGradient id="colorIntensity" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.8} />
                    <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0.1} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                <XAxis
                  dataKey="timestamp"
                  stroke="#9ca3af"
                  tick={{ fill: '#9ca3af', fontSize: 12 }}
                  angle={-45}
                  textAnchor="end"
                  height={80}
                />
                <YAxis
                  stroke="#9ca3af"
                  tick={{ fill: '#9ca3af', fontSize: 12 }}
                  label={{ value: 'gCO₂/kWh', angle: -90, position: 'insideLeft', fill: '#9ca3af' }}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1e1b4b',
                    border: '1px solid #4c1d95',
                    borderRadius: '8px',
                    color: '#e0e7ff'
                  }}
                  labelStyle={{ color: '#a5b4fc' }}
                />
                <Area
                  type="monotone"
                  dataKey="intensity"
                  stroke="#8b5cf6"
                  strokeWidth={2}
                  fillOpacity={1}
                  fill="url(#colorIntensity)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        )}
      </motion.div>

      {/* Job Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.6 }}
          className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple"
        >
          <h3 className="text-lg font-bold text-karbos-light-blue mb-4">Job Status Breakdown</h3>
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <span className="text-gray-400 flex items-center gap-2">
                <Activity className="w-4 h-4 text-blue-400" />
                Running
              </span>
              <span className="text-blue-400 font-semibold">{kpiData.activeJobs}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-400 flex items-center gap-2">
                <Clock className="w-4 h-4 text-yellow-400" />
                Pending/Delayed
              </span>
              <span className="text-yellow-400 font-semibold">{kpiData.pendingOptimized}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-400 flex items-center gap-2">
                <CheckCircle className="w-4 h-4 text-green-400" />
                Completed
              </span>
              <span className="text-green-400 font-semibold">{kpiData.completedJobs}</span>
            </div>
            {kpiData.failedJobs > 0 && (
              <div className="flex justify-between items-center">
                <span className="text-gray-400 flex items-center gap-2">
                  <span className="w-4 h-4 text-red-400">✕</span>
                  Failed
                </span>
                <span className="text-red-400 font-semibold">{kpiData.failedJobs}</span>
              </div>
            )}
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.7 }}
          className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple"
        >
          <h3 className="text-lg font-bold text-karbos-light-blue mb-4">System Health</h3>
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <span className="text-gray-400">Total Jobs</span>
              <span className="text-karbos-light-blue font-semibold">{jobs?.length || 0}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-400">Success Rate</span>
              <span className="text-green-400 font-semibold">
                {jobs && jobs.length > 0
                  ? Math.round((kpiData.completedJobs / jobs.length) * 100)
                  : 0}%
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-400">Regions Monitored</span>
              <span className="text-karbos-light-blue font-semibold">{regions.length || 0}</span>
            </div>
          </div>
        </motion.div>
      </div>
    </motion.div>
  );
};

export default Overview;
