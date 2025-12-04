"use client";

"use client";

import React from "react";
import { motion } from "framer-motion";

interface WorkerNode {
  id: string;
  name: string;
  status: "online" | "offline" | "busy";
  lastHeartbeat: string;
  jobsProcessed: number;
  cpuUsage: number;
  memoryUsage: number;
}

const Infrastructure = () => {
  // Mock data for worker nodes
  const workerNodes: WorkerNode[] = [
    {
      id: "node-01",
      name: "worker-us-east-1a",
      status: "online",
      lastHeartbeat: "2025-12-04T14:45:23",
      jobsProcessed: 47,
      cpuUsage: 23,
      memoryUsage: 45,
    },
    {
      id: "node-02",
      name: "worker-us-east-1b",
      status: "busy",
      lastHeartbeat: "2025-12-04T14:45:20",
      jobsProcessed: 62,
      cpuUsage: 87,
      memoryUsage: 72,
    },
    {
      id: "node-03",
      name: "worker-us-west-2a",
      status: "online",
      lastHeartbeat: "2025-12-04T14:45:18",
      jobsProcessed: 34,
      cpuUsage: 15,
      memoryUsage: 38,
    },
    {
      id: "node-04",
      name: "worker-eu-central-1a",
      status: "offline",
      lastHeartbeat: "2025-12-04T14:12:45",
      jobsProcessed: 28,
      cpuUsage: 0,
      memoryUsage: 0,
    },
    {
      id: "node-05",
      name: "worker-eu-central-1b",
      status: "online",
      lastHeartbeat: "2025-12-04T14:45:22",
      jobsProcessed: 51,
      cpuUsage: 41,
      memoryUsage: 56,
    },
    {
      id: "node-06",
      name: "worker-ap-southeast-1a",
      status: "busy",
      lastHeartbeat: "2025-12-04T14:45:19",
      jobsProcessed: 39,
      cpuUsage: 93,
      memoryUsage: 81,
    },
  ];

  const queueData = {
    ready: 12,
    delayed: 8,
    active: 5,
    completed: 247,
    failed: 3,
  };

  const getStatusColor = (status: WorkerNode["status"]) => {
    switch (status) {
      case "online":
        return "bg-green-500";
      case "busy":
        return "bg-yellow-500";
      case "offline":
        return "bg-red-500";
    }
  };

  const getStatusText = (status: WorkerNode["status"]) => {
    switch (status) {
      case "online":
        return "text-green-400";
      case "busy":
        return "text-yellow-400";
      case "offline":
        return "text-red-400";
    }
  };

  const formatLastSeen = (timestamp: string) => {
    const now = new Date("2025-12-04T14:45:30"); // Mock current time
    const then = new Date(timestamp);
    const diff = Math.floor((now.getTime() - then.getTime()) / 1000);

    if (diff < 60) return `${diff}s ago`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    return `${Math.floor(diff / 3600)}h ago`;
  };

  return (
    <motion.div 
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
      className="space-y-6"
    >
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-karbos-light-blue">
          Infrastructure
        </h2>
        <p className="text-karbos-lavender mt-1">
          Monitor worker nodes and queue health
        </p>
      </div>

      {/* Queue Health */}
      <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
        <h3 className="text-xl font-semibold text-karbos-light-blue mb-4">
          Queue Health
        </h3>

        <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
          <div className="text-center">
            <div className="relative w-32 h-32 mx-auto">
              <svg className="w-full h-full transform -rotate-90">
                <circle
                  cx="64"
                  cy="64"
                  r="56"
                  className="fill-none stroke-karbos-navy"
                  strokeWidth="12"
                />
                <circle
                  cx="64"
                  cy="64"
                  r="56"
                  className="fill-none stroke-blue-500"
                  strokeWidth="12"
                  strokeDasharray={`${(queueData.ready / 50) * 352} ${352 - (queueData.ready / 50) * 352}`}
                />
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-3xl font-bold text-karbos-light-blue">
                  {queueData.ready}
                </span>
                <span className="text-xs text-karbos-lavender">jobs</span>
              </div>
            </div>
            <p className="text-karbos-lavender mt-2 font-medium">Ready Queue</p>
          </div>

          <div className="text-center">
            <div className="relative w-32 h-32 mx-auto">
              <svg className="w-full h-full transform -rotate-90">
                <circle
                  cx="64"
                  cy="64"
                  r="56"
                  className="fill-none stroke-karbos-navy"
                  strokeWidth="12"
                />
                <circle
                  cx="64"
                  cy="64"
                  r="56"
                  className="fill-none stroke-yellow-500"
                  strokeWidth="12"
                  strokeDasharray={`${(queueData.delayed / 50) * 352} ${352 - (queueData.delayed / 50) * 352}`}
                />
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-3xl font-bold text-karbos-light-blue">
                  {queueData.delayed}
                </span>
                <span className="text-xs text-karbos-lavender">jobs</span>
              </div>
            </div>
            <p className="text-karbos-lavender mt-2 font-medium">
              Delayed Queue
            </p>
          </div>

          <div className="text-center">
            <div className="relative w-32 h-32 mx-auto">
              <svg className="w-full h-full transform -rotate-90">
                <circle
                  cx="64"
                  cy="64"
                  r="56"
                  className="fill-none stroke-karbos-navy"
                  strokeWidth="12"
                />
                <circle
                  cx="64"
                  cy="64"
                  r="56"
                  className="fill-none stroke-green-500"
                  strokeWidth="12"
                  strokeDasharray={`${(queueData.active / 50) * 352} ${352 - (queueData.active / 50) * 352}`}
                />
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-3xl font-bold text-karbos-light-blue">
                  {queueData.active}
                </span>
                <span className="text-xs text-karbos-lavender">jobs</span>
              </div>
            </div>
            <p className="text-karbos-lavender mt-2 font-medium">Active</p>
          </div>

          <div className="text-center">
            <div className="relative w-32 h-32 mx-auto">
              <svg className="w-full h-full transform -rotate-90">
                <circle
                  cx="64"
                  cy="64"
                  r="56"
                  className="fill-none stroke-karbos-navy"
                  strokeWidth="12"
                />
                <circle
                  cx="64"
                  cy="64"
                  r="56"
                  className="fill-none stroke-cyan-500"
                  strokeWidth="12"
                  strokeDasharray="352 0"
                />
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-3xl font-bold text-karbos-light-blue">
                  {queueData.completed}
                </span>
                <span className="text-xs text-karbos-lavender">jobs</span>
              </div>
            </div>
            <p className="text-karbos-lavender mt-2 font-medium">Completed</p>
          </div>

          <div className="text-center">
            <div className="relative w-32 h-32 mx-auto">
              <svg className="w-full h-full transform -rotate-90">
                <circle
                  cx="64"
                  cy="64"
                  r="56"
                  className="fill-none stroke-karbos-navy"
                  strokeWidth="12"
                />
                <circle
                  cx="64"
                  cy="64"
                  r="56"
                  className="fill-none stroke-red-500"
                  strokeWidth="12"
                  strokeDasharray={`${(queueData.failed / 50) * 352} ${352 - (queueData.failed / 50) * 352}`}
                />
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-3xl font-bold text-karbos-light-blue">
                  {queueData.failed}
                </span>
                <span className="text-xs text-karbos-lavender">jobs</span>
              </div>
            </div>
            <p className="text-karbos-lavender mt-2 font-medium">Failed</p>
          </div>
        </div>
      </div>

      {/* Worker Nodes Grid */}
      <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-xl font-semibold text-karbos-light-blue">
            Worker Nodes
          </h3>
          <div className="flex items-center space-x-4 text-sm">
            <div className="flex items-center space-x-2">
              <div className="w-3 h-3 rounded-full bg-green-500"></div>
              <span className="text-karbos-lavender">Online</span>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
              <span className="text-karbos-lavender">Busy</span>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-3 h-3 rounded-full bg-red-500"></div>
              <span className="text-karbos-lavender">Offline</span>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {workerNodes.map((node, index) => (
            <motion.div
              key={node.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3, delay: index * 0.05 }}
              whileHover={{ scale: 1.02, y: -4 }}
              className="bg-karbos-navy p-4 rounded-lg border border-karbos-blue-purple hover:border-karbos-lavender transition-colors"
            >
              <div className="flex items-start justify-between mb-3">
                <div>
                  <h4 className="text-karbos-light-blue font-semibold">
                    {node.name}
                  </h4>
                  <p className="text-xs text-karbos-lavender font-mono">
                    {node.id}
                  </p>
                </div>
                <motion.div
                  animate={{ scale: [1, 1.2, 1] }}
                  transition={{ duration: 2, repeat: Infinity, delay: index * 0.2 }}
                  className={`w-3 h-3 rounded-full ${getStatusColor(node.status)}`}
                />
              </div>

              <div className="space-y-2 mb-3">
                <div className="flex justify-between text-sm">
                  <span className="text-karbos-lavender">Status:</span>
                  <span
                    className={`font-medium capitalize ${getStatusText(node.status)}`}
                  >
                    {node.status}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-karbos-lavender">Jobs Processed:</span>
                  <span className="text-karbos-light-blue font-medium">
                    {node.jobsProcessed}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-karbos-lavender">Last Heartbeat:</span>
                  <span className="text-karbos-light-blue font-medium">
                    {formatLastSeen(node.lastHeartbeat)}
                  </span>
                </div>
              </div>

              <div className="space-y-2">
                <div>
                  <div className="flex justify-between text-xs text-karbos-lavender mb-1">
                    <span>CPU</span>
                    <span>{node.cpuUsage}%</span>
                  </div>
                  <div className="w-full bg-karbos-indigo rounded-full h-2">
                    <div
                      className={`h-2 rounded-full transition-all ${
                        node.cpuUsage > 80
                          ? "bg-red-500"
                          : node.cpuUsage > 50
                            ? "bg-yellow-500"
                            : "bg-green-500"
                      }`}
                      style={{ width: `${node.cpuUsage}%` }}
                    />
                  </div>
                </div>
                <div>
                  <div className="flex justify-between text-xs text-karbos-lavender mb-1">
                    <span>Memory</span>
                    <span>{node.memoryUsage}%</span>
                  </div>
                  <div className="w-full bg-karbos-indigo rounded-full h-2">
                    <div
                      className={`h-2 rounded-full transition-all ${
                        node.memoryUsage > 80
                          ? "bg-red-500"
                          : node.memoryUsage > 50
                            ? "bg-yellow-500"
                            : "bg-green-500"
                      }`}
                      style={{ width: `${node.memoryUsage}%` }}
                    />
                  </div>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </div>

      {/* System Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
          <h4 className="text-karbos-lavender text-sm font-medium mb-2">
            Total Nodes
          </h4>
          <p className="text-3xl font-bold text-karbos-light-blue">
            {workerNodes.length}
          </p>
          <p className="text-sm text-green-400 mt-1">
            {workerNodes.filter((n) => n.status !== "offline").length} active
          </p>
        </div>

        <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
          <h4 className="text-karbos-lavender text-sm font-medium mb-2">
            Total Jobs Processed
          </h4>
          <p className="text-3xl font-bold text-karbos-light-blue">
            {workerNodes.reduce((sum, node) => sum + node.jobsProcessed, 0)}
          </p>
          <p className="text-sm text-blue-400 mt-1">Across all nodes</p>
        </div>

        <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
          <h4 className="text-karbos-lavender text-sm font-medium mb-2">
            Average CPU Usage
          </h4>
          <p className="text-3xl font-bold text-karbos-light-blue">
            {Math.round(
              workerNodes.reduce((sum, node) => sum + node.cpuUsage, 0) /
                workerNodes.length,
            )}
            %
          </p>
          <p className="text-sm text-karbos-lavender mt-1">Cluster-wide</p>
        </div>
      </div>
    </motion.div>
  );
};

export default Infrastructure;
