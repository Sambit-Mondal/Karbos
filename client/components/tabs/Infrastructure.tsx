"use client";

import React from "react";
import { motion } from "framer-motion";
import useSWR from "swr";
import { apiClient } from "@/lib/api";
import type { SystemHealthResponse } from "@/lib/types";
import { 
  Server, 
  Activity, 
  Clock,
  Database,
  TrendingUp,
  Zap,
  AlertCircle,
  CheckCircle2
} from "lucide-react";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from "recharts";

const Infrastructure = () => {
  // Poll every 3 seconds for real-time updates
  const { data: health, error, isLoading } = useSWR<SystemHealthResponse>(
    '/api/system/health',
    () => apiClient.getSystemHealth(),
    {
      refreshInterval: 3000,
      revalidateOnFocus: true,
      revalidateOnReconnect: true,
    }
  );

  const queueData = health ? [
    {
      name: 'Immediate',
      jobs: health.queue_depth_immediate,
      color: '#FFD700', // Yellow/Gold
    },
    {
      name: 'Delayed',
      jobs: health.queue_depth_delayed,
      color: '#8B5CF6', // Purple
    }
  ] : [];

  if (error) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="text-center">
          <AlertCircle className="w-16 h-16 text-red-400 mx-auto mb-4" />
          <p className="text-xl font-semibold text-white">Failed to load system health</p>
          <p className="text-sm text-gray-400 mt-2">{error.message}</p>
        </div>
      </div>
    );
  }

  if (isLoading || !health) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="text-center">
          <Activity className="w-16 h-16 text-karbos-blue-purple animate-pulse mx-auto mb-4" />
          <p className="text-xl font-semibold text-white">Loading system health...</p>
        </div>
      </div>
    );
  }

  const totalQueueDepth = health.queue_depth_immediate + health.queue_depth_delayed;
  const isHealthy = health.active_workers > 0;

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
      className="space-y-6"
    >
      {/* Header */}
      <div className="text-center">
        <h2 className="text-3xl font-bold text-karbos-light-blue">Infrastructure Console</h2>
        <p className="text-sm text-gray-400 mt-2">
          Real-time monitoring of distributed worker nodes and job queues
        </p>
      </div>

      {/* Status Banner */}
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className={`rounded-lg p-4 border ${
          isHealthy 
            ? 'bg-green-500/10 border-green-500/20' 
            : 'bg-red-500/10 border-red-500/20'
        }`}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            {isHealthy ? (
              <CheckCircle2 className="w-6 h-6 text-green-400" />
            ) : (
              <AlertCircle className="w-6 h-6 text-red-400" />
            )}
            <div>
              <p className={`font-bold ${isHealthy ? 'text-green-400' : 'text-red-400'}`}>
                System Status: {isHealthy ? 'Operational' : 'Degraded'}
              </p>
              <p className="text-sm text-gray-400">
                Last updated: {new Date(health.timestamp).toLocaleTimeString()}
              </p>
            </div>
          </div>
          <div className="text-right">
            <p className="text-sm text-gray-400">Auto-refresh: 3s</p>
          </div>
        </div>
      </motion.div>

      {/* Metrics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Active Workers Card */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.1 }}
          className="bg-karbos-dark-gray border border-green-500/20 rounded-lg p-6 relative overflow-hidden"
        >
          <div className="absolute top-0 right-0 w-32 h-32 bg-green-500/5 rounded-full blur-3xl" />
          <div className="relative">
            <div className="flex items-center justify-between mb-4">
              <Server className="w-8 h-8 text-green-400" />
              <span className={`px-3 py-1 rounded-full text-xs font-semibold ${
                health.active_workers > 0 
                  ? 'bg-green-500/20 text-green-400' 
                  : 'bg-red-500/20 text-red-400'
              }`}>
                {health.active_workers > 0 ? 'ONLINE' : 'OFFLINE'}
              </span>
            </div>
            <p className="text-5xl font-bold text-white mb-2">{health.active_workers}</p>
            <p className="text-sm text-gray-400">Active Worker Nodes</p>
          </div>
        </motion.div>

        {/* Pending Jobs Card */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.2 }}
          className="bg-karbos-dark-gray border border-yellow-500/20 rounded-lg p-6 relative overflow-hidden"
        >
          <div className="absolute top-0 right-0 w-32 h-32 bg-yellow-500/5 rounded-full blur-3xl" />
          <div className="relative">
            <div className="flex items-center justify-between mb-4">
              <TrendingUp className="w-8 h-8 text-yellow-400" />
              <span className={`px-3 py-1 rounded-full text-xs font-semibold ${
                totalQueueDepth > 50 
                  ? 'bg-red-500/20 text-red-400' 
                  : totalQueueDepth > 20
                    ? 'bg-yellow-500/20 text-yellow-400'
                    : 'bg-green-500/20 text-green-400'
              }`}>
                {totalQueueDepth > 50 ? 'HIGH' : totalQueueDepth > 20 ? 'MEDIUM' : 'LOW'}
              </span>
            </div>
            <p className="text-5xl font-bold text-white mb-2">{totalQueueDepth}</p>
            <p className="text-sm text-gray-400">Total Pending Jobs</p>
            <div className="flex gap-4 mt-3 text-xs">
              <span className="text-yellow-400">⚡ {health.queue_depth_immediate} Immediate</span>
              <span className="text-purple-400">⏰ {health.queue_depth_delayed} Delayed</span>
            </div>
          </div>
        </motion.div>

        {/* Redis Latency Card */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.3 }}
          className="bg-karbos-dark-gray border border-blue-500/20 rounded-lg p-6 relative overflow-hidden"
        >
          <div className="absolute top-0 right-0 w-32 h-32 bg-blue-500/5 rounded-full blur-3xl" />
          <div className="relative">
            <div className="flex items-center justify-between mb-4">
              <Database className="w-8 h-8 text-blue-400" />
              <span className={`px-3 py-1 rounded-full text-xs font-semibold ${
                health.redis_latency_ms < 5 
                  ? 'bg-green-500/20 text-green-400' 
                  : health.redis_latency_ms < 20
                    ? 'bg-yellow-500/20 text-yellow-400'
                    : 'bg-red-500/20 text-red-400'
              }`}>
                {health.redis_latency_ms < 5 ? 'FAST' : health.redis_latency_ms < 20 ? 'NORMAL' : 'SLOW'}
              </span>
            </div>
            <div className="flex items-baseline gap-2">
              <p className="text-5xl font-bold text-white">{health.redis_latency_ms}</p>
              <span className="text-xl text-gray-400">ms</span>
            </div>
            <p className="text-sm text-gray-400 mt-2">Redis Ping Latency</p>
          </div>
        </motion.div>
      </div>

      {/* Node Grid Section */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4 }}
        className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-6"
      >
        <div className="flex items-center gap-2 mb-6">
          <Server className="w-6 h-6 text-karbos-blue-purple" />
          <h3 className="text-xl font-semibold text-white">Worker Node Fleet</h3>
          <span className="ml-auto text-sm text-gray-400">
            {health.active_workers} / {health.active_workers} nodes online
          </span>
        </div>

        {health.active_workers === 0 ? (
          <div className="text-center py-12">
            <Server className="w-16 h-16 text-gray-600 mx-auto mb-4" />
            <p className="text-lg text-gray-400">No active worker nodes</p>
            <p className="text-sm text-gray-500 mt-2">Start a worker to see it appear here</p>
          </div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
            {health.worker_ids.map((workerId, index) => (
              <motion.div
                key={workerId}
                initial={{ opacity: 0, scale: 0.8 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.5 + index * 0.05 }}
                className="bg-karbos-dark-gray/50 border border-green-500/30 rounded-lg p-4 flex flex-col items-center justify-center relative group hover:border-green-500/60 transition-all"
              >
                {/* Pulse animation */}
                <div className="absolute inset-0 rounded-lg bg-green-500/10 animate-pulse" />
                
                <div className="relative">
                  <Server className="w-10 h-10 text-green-400 mb-2" />
                  <div className="absolute -top-1 -right-1 w-3 h-3 bg-green-400 rounded-full animate-ping" />
                  <div className="absolute -top-1 -right-1 w-3 h-3 bg-green-400 rounded-full" />
                </div>
                
                <p className="text-xs text-gray-400 mt-2 text-center break-all">
                  {workerId.substring(0, 8)}...
                </p>
                
                {/* Tooltip on hover */}
                <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-black/90 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-10">
                  {workerId}
                </div>
              </motion.div>
            ))}
          </div>
        )}
      </motion.div>

      {/* Queue Visualization */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5 }}
        className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-6"
      >
        <div className="flex items-center gap-2 mb-6">
          <Activity className="w-6 h-6 text-karbos-blue-purple" />
          <h3 className="text-xl font-semibold text-white">Queue Visualization</h3>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Bar Chart */}
          <div>
            <h4 className="text-sm font-semibold text-gray-400 mb-4">Queue Depth Comparison</h4>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={queueData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                <XAxis 
                  dataKey="name" 
                  stroke="#9CA3AF"
                  style={{ fontSize: '14px' }}
                />
                <YAxis 
                  stroke="#9CA3AF"
                  style={{ fontSize: '14px' }}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1F2937',
                    border: '1px solid #374151',
                    borderRadius: '8px',
                    color: '#fff'
                  }}
                  cursor={{ fill: 'rgba(139, 92, 246, 0.1)' }}
                />
                <Bar dataKey="jobs" radius={[8, 8, 0, 0]}>
                  {queueData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>

          {/* Queue Stats Cards */}
          <div className="space-y-4">
            <div className="bg-karbos-dark-gray/50 border border-yellow-500/20 rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Zap className="w-5 h-5 text-yellow-400" />
                  <h4 className="text-sm font-semibold text-gray-300">Immediate Queue</h4>
                </div>
                <span className="text-2xl font-bold text-yellow-400">
                  {health.queue_depth_immediate}
                </span>
              </div>
              <p className="text-xs text-gray-500">
                Jobs scheduled to run immediately (FIFO)
              </p>
            </div>

            <div className="bg-karbos-dark-gray/50 border border-purple-500/20 rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Clock className="w-5 h-5 text-purple-400" />
                  <h4 className="text-sm font-semibold text-gray-300">Delayed Queue</h4>
                </div>
                <span className="text-2xl font-bold text-purple-400">
                  {health.queue_depth_delayed}
                </span>
              </div>
              <p className="text-xs text-gray-500">
                Carbon-aware jobs waiting for optimal time
              </p>
            </div>

            <div className="bg-karbos-dark-gray/50 border border-blue-500/20 rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <TrendingUp className="w-5 h-5 text-blue-400" />
                  <h4 className="text-sm font-semibold text-gray-300">Total Workload</h4>
                </div>
                <span className="text-2xl font-bold text-blue-400">
                  {totalQueueDepth}
                </span>
              </div>
              <p className="text-xs text-gray-500">
                Combined jobs across all queues
              </p>
            </div>
          </div>
        </div>
      </motion.div>

      {/* System Info Footer */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.6 }}
        className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4"
      >
        <p className="text-sm text-gray-300">
          <strong className="text-blue-400">📊 Real-time Monitoring:</strong> This dashboard updates every 3 seconds. 
          Worker nodes send heartbeats every 10 seconds with a 15-second TTL. 
          Queue depths are queried directly from Redis (LLEN for immediate, ZCARD for delayed).
        </p>
      </motion.div>
    </motion.div>
  );
};

export default Infrastructure;
