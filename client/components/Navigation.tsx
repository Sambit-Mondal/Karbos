"use client";

import React, { useState } from "react";
import { motion } from "framer-motion";

interface NavigationProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

const Navigation: React.FC<NavigationProps> = ({ activeTab, onTabChange }) => {
  const [region, setRegion] = useState("us-east-1");
  const systemStatus = {
    scheduler: true,
    redis: true,
  };

  const tabs = [
    { id: "overview", label: "Overview" },
    { id: "workloads", label: "Workloads" },
    { id: "grid", label: "Grid Intelligence" },
    { id: "infrastructure", label: "Infrastructure" },
    { id: "playground", label: "Playground" },
  ];

  const regions = [
    "us-east-1",
    "us-west-2",
    "eu-central-1",
    "eu-west-1",
    "ap-southeast-1",
  ];

  return (
    <motion.nav 
      initial={{ y: -20, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.4 }}
      className="bg-karbos-navy border-b border-karbos-indigo"
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Brand */}
          <div className="flex items-center">
            <h1 className="text-2xl font-bold text-karbos-light-blue">
              Karbos
            </h1>
          </div>

          {/* Tabs */}
          <div className="hidden md:flex space-x-4">
            {tabs.map((tab) => (
              <motion.button
                key={tab.id}
                onClick={() => onTabChange(tab.id)}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                className={`px-3 py-2 rounded-md text-sm font-medium transition-colors relative ${
                  activeTab === tab.id
                    ? "bg-karbos-blue-purple text-white"
                    : "text-karbos-lavender hover:bg-karbos-indigo hover:text-white"
                }`}
              >
                {tab.label}
                {activeTab === tab.id && (
                  <motion.div
                    layoutId="activeTab"
                    className="absolute inset-0 bg-karbos-blue-purple rounded-md -z-10"
                    transition={{ type: "spring", stiffness: 300, damping: 30 }}
                  />
                )}
              </motion.button>
            ))}
          </div>

          {/* Region Selector & System Status */}
          <div className="flex items-center space-x-4">
            <select
              value={region}
              onChange={(e) => setRegion(e.target.value)}
              className="bg-karbos-indigo text-karbos-light-blue px-3 py-1 rounded-md text-sm border border-karbos-blue-purple focus:outline-none focus:ring-2 focus:ring-karbos-lavender"
            >
              {regions.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>

            <div className="flex items-center space-x-2">
              <div className="flex items-center space-x-1">
                <motion.div
                  animate={systemStatus.scheduler ? { scale: [1, 1.2, 1] } : {}}
                  transition={{ duration: 2, repeat: Infinity }}
                  className={`w-2 h-2 rounded-full ${
                    systemStatus.scheduler ? "bg-green-500" : "bg-red-500"
                  }`}
                  title={`Scheduler: ${systemStatus.scheduler ? "Online" : "Offline"}`}
                />
                <span className="text-xs text-karbos-lavender">Scheduler</span>
              </div>
              <div className="flex items-center space-x-1">
                <motion.div
                  animate={systemStatus.redis ? { scale: [1, 1.2, 1] } : {}}
                  transition={{ duration: 2, repeat: Infinity, delay: 0.5 }}
                  className={`w-2 h-2 rounded-full ${
                    systemStatus.redis ? "bg-green-500" : "bg-red-500"
                  }`}
                  title={`Redis: ${systemStatus.redis ? "Online" : "Offline"}`}
                />
                <span className="text-xs text-karbos-lavender">Redis</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Mobile Tabs */}
      <div className="md:hidden px-4 pb-3 space-y-1">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => onTabChange(tab.id)}
            className={`block w-full text-left px-3 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? "bg-karbos-blue-purple text-white"
                : "text-karbos-lavender hover:bg-karbos-indigo hover:text-white"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>
    </motion.nav>
  );
};

export default Navigation;
