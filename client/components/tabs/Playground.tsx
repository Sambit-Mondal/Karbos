"use client";

import React, { useState, useEffect, useMemo, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import useSWR from "swr";
import { apiClient } from "@/lib/api";
import type { CarbonForecastResponse, SubmitJobRequest, SubmitJobResponse } from "@/lib/types";
import { 
  Play, 
  Loader, 
  CheckCircle, 
  AlertCircle,
  TrendingDown,
  Clock,
  Leaf,
  Zap,
  Settings
} from "lucide-react";
import { useRouter } from "next/navigation";

const DOCKER_IMAGES = [
  { value: "python:3.9-slim", label: "Python 3.9 Slim", description: "Lightweight Python runtime" },
  { value: "python:3.11-slim", label: "Python 3.11 Slim", description: "Latest Python 3.11" },
  { value: "node:18-alpine", label: "Node.js 18 Alpine", description: "Minimal Node.js 18" },
  { value: "node:20-alpine", label: "Node.js 20 Alpine", description: "Latest Node.js 20" },
  { value: "alpine:latest", label: "Alpine Linux", description: "Minimal Linux distribution" },
  { value: "golang:1.21-alpine", label: "Go 1.21 Alpine", description: "Go programming language" },
  { value: "ubuntu:22.04", label: "Ubuntu 22.04", description: "Full Ubuntu environment" },
  { value: "nginx:alpine", label: "Nginx Alpine", description: "Web server" },
];

const REGIONS = [
  { value: "US-EAST", label: "US East (Virginia)" },
  { value: "US-WEST", label: "US West (Oregon)" },
  { value: "US-CENTRAL", label: "US Central" },
  { value: "EU-WEST", label: "EU West" },
  { value: "EU-CENTRAL", label: "EU Central" },
  { value: "EU-NORTH", label: "EU North" },
  { value: "ASIA-EAST", label: "Asia East" },
  { value: "AU-EAST", label: "Australia East" },
];

const Playground = () => {
  const router = useRouter();
  const [dockerImage, setDockerImage] = useState(DOCKER_IMAGES[0].value);
  const [command, setCommand] = useState("echo 'Hello, Carbon-Aware World!'");
  const [duration, setDuration] = useState(30); // minutes
  const [deadlineHours, setDeadlineHours] = useState(4); // hours
  const [region, setRegion] = useState("US-EAST");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showSuccess, setShowSuccess] = useState(false);
  const [showError, setShowError] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [prediction, setPrediction] = useState<SubmitJobResponse | null>(null);

  // Fetch carbon forecast for selected region
  const { data: carbonData } = useSWR<CarbonForecastResponse>(
    `/api/carbon-forecast?region=${region}`,
    () => apiClient.getCarbonForecast(region),
    {
      refreshInterval: 60000,
    }
  );

  // Calculate CO2 savings prediction client-side
  const impactPrediction = useMemo(() => {
    if (!carbonData?.forecasts || carbonData.forecasts.length === 0) {
      return null;
    }

    const now = new Date();
    const deadlineTime = new Date(now.getTime() + deadlineHours * 60 * 60 * 1000);
    
    // Filter forecasts within deadline
    const viableForecasts = carbonData.forecasts.filter(f => {
      const forecastTime = new Date(f.timestamp);
      return forecastTime >= now && forecastTime <= deadlineTime;
    });

    if (viableForecasts.length === 0) return null;

    // Current intensity (now or nearest future)
    const currentIntensity = viableForecasts[0]?.intensity_value || 300;
    
    // Find greenest window
    const optimalWindow = viableForecasts.reduce((min, curr) => 
      curr.intensity_value < min.intensity_value ? curr : min
    );

    const savingsPercent = ((currentIntensity - optimalWindow.intensity_value) / currentIntensity) * 100;
    const co2Savings = ((currentIntensity - optimalWindow.intensity_value) * duration) / 60; // gCO2eq

    return {
      currentIntensity,
      optimalIntensity: optimalWindow.intensity_value,
      optimalTime: new Date(optimalWindow.timestamp),
      savingsPercent: Math.max(0, savingsPercent),
      co2Savings: Math.max(0, co2Savings),
      isImmediate: savingsPercent < 10, // Less than 10% savings, run immediately
    };
  }, [carbonData, deadlineHours, duration]);

  const handleDryRun = useCallback(async () => {
    const deadline = new Date();
    deadline.setHours(deadline.getHours() + deadlineHours);

    const request: SubmitJobRequest = {
      user_id: "playground-user",
      docker_image: dockerImage,
      command: [command],
      deadline: deadline.toISOString(),
      estimated_duration: duration * 60, // Convert to seconds
      region: region,
    };

    try {
      const result = await apiClient.submitJob(request, true); // dry_run = true
      setPrediction(result);
    } catch (error) {
      console.error("Dry run failed:", error);
      setPrediction(null);
    }
  }, [dockerImage, command, duration, deadlineHours, region]);

  // Auto-run dry-run when inputs change
  useEffect(() => {
    if (!dockerImage || !command || duration <= 0) return;

    const timer = setTimeout(() => {
      handleDryRun();
    }, 1000); // Debounce 1 second

    return () => clearTimeout(timer);
  }, [dockerImage, command, duration, deadlineHours, region, handleDryRun]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!dockerImage || !command) {
      setErrorMessage("Please fill in all required fields");
      setShowError(true);
      setTimeout(() => setShowError(false), 3000);
      return;
    }

    setIsSubmitting(true);

    const deadline = new Date();
    deadline.setHours(deadline.getHours() + deadlineHours);

    const request: SubmitJobRequest = {
      user_id: "playground-user",
      docker_image: dockerImage,
      command: [command],
      deadline: deadline.toISOString(),
      estimated_duration: duration * 60,
      region: region,
    };

    try {
      const result = await apiClient.submitJob(request, false); // actual submission
      setPrediction(result);
      setShowSuccess(true);
      
      // Auto-redirect to Workloads after 2 seconds
      setTimeout(() => {
        router.push("/?tab=workloads");
      }, 2000);
    } catch (error: unknown) {
      console.error("Job submission failed:", error);
      const message = error instanceof Error && 'response' in error 
        ? (error as { response?: { data?: { message?: string } } }).response?.data?.message 
        : error instanceof Error 
        ? error.message 
        : "Failed to submit job";
      setErrorMessage(message || "Failed to submit job");
      setShowError(true);
      setTimeout(() => setShowError(false), 5000);
    } finally {
      setIsSubmitting(false);
    }
  };

  const formatTime = (date: Date) => {
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
      className="space-y-6 max-w-4xl mx-auto"
    >
      {/* Header */}
      <div className="text-center">
        <h2 className="text-3xl font-bold text-karbos-light-blue">Job Playground</h2>
        <p className="text-sm text-gray-400 mt-2">
          Submit carbon-aware jobs and see real-time CO₂ savings predictions
        </p>
      </div>

      {/* Success Toast */}
      <AnimatePresence>
        {showSuccess && (
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="bg-green-500/10 border border-green-500/20 rounded-lg p-4 flex items-start gap-3"
          >
            <CheckCircle className="w-5 h-5 text-green-400 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-green-400 font-semibold">Job Scheduled Successfully!</p>
              <p className="text-sm text-gray-300 mt-1">
                Redirecting to Workloads tab...
              </p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Error Toast */}
      <AnimatePresence>
        {showError && (
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 flex items-start gap-3"
          >
            <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-red-400 font-semibold">Submission Failed</p>
              <p className="text-sm text-gray-300 mt-1">{errorMessage}</p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Left Column - Form */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.1 }}
        >
          <form onSubmit={handleSubmit} className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-6 space-y-6">
            <div className="flex items-center gap-2 mb-4">
              <Settings className="w-5 h-5 text-karbos-blue-purple" />
              <h3 className="text-lg font-semibold text-white">Job Configuration</h3>
            </div>

            {/* Docker Image Selector */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Docker Image *
              </label>
              <select
                value={dockerImage}
                onChange={(e) => setDockerImage(e.target.value)}
                className="w-full px-4 py-3 bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-md text-black focus:outline-none focus:border-karbos-blue-purple transition-colors"
                required
              >
                {DOCKER_IMAGES.map(img => (
                  <option key={img.value} value={img.value}>
                    {img.label} - {img.description}
                  </option>
                ))}
              </select>
            </div>

            {/* Command Input */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Command *
              </label>
              <input
                type="text"
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                placeholder='e.g., python script.py'
                className="w-full px-4 py-3 bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-md text-black focus:outline-none focus:border-karbos-blue-purple transition-colors"
                required
              />
              <p className="text-xs text-gray-500 mt-1">Shell command to execute in the container</p>
            </div>

            {/* Duration Input */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Estimated Duration: {duration} minutes
              </label>
              <input
                type="range"
                min="5"
                max="180"
                step="5"
                value={duration}
                onChange={(e) => setDuration(parseInt(e.target.value))}
                className="w-full h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-karbos-blue-purple"
              />
              <div className="flex justify-between text-xs text-gray-500 mt-1">
                <span>5 min</span>
                <span>180 min (3 hrs)</span>
              </div>
            </div>

            {/* Region Selector */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Region
              </label>
              <select
                value={region}
                onChange={(e) => setRegion(e.target.value)}
                className="w-full px-4 py-3 bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-md text-black focus:outline-none focus:border-karbos-blue-purple transition-colors"
              >
                {REGIONS.map(r => (
                  <option key={r.value} value={r.value}>
                    {r.label}
                  </option>
                ))}
              </select>
            </div>

            {/* SLA Deadline Slider */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                <Clock className="w-4 h-4 inline mr-2" />
                Deadline: I need this done by {formatTime(new Date(Date.now() + deadlineHours * 60 * 60 * 1000))}
              </label>
              <input
                type="range"
                min="1"
                max="24"
                step="1"
                value={deadlineHours}
                onChange={(e) => setDeadlineHours(parseInt(e.target.value))}
                className="w-full h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-karbos-blue-purple"
              />
              <div className="flex justify-between text-xs text-gray-500 mt-1">
                <span>1 Hour</span>
                <span>24 Hours</span>
              </div>
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full px-6 py-3 bg-karbos-blue-purple text-white rounded-md hover:bg-opacity-80 transition-colors flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed font-semibold"
            >
              {isSubmitting ? (
                <>
                  <Loader className="w-5 h-5 animate-spin" />
                  Submitting...
                </>
              ) : (
                <>
                  <Play className="w-5 h-5" />
                  Submit Job
                </>
              )}
            </button>
          </form>
        </motion.div>

        {/* Right Column - Impact Prediction */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.2 }}
          className="space-y-4"
        >
          {/* Impact Predictor */}
          {impactPrediction && (
            <div className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-6">
              <div className="flex items-center gap-2 mb-4">
                <Leaf className="w-5 h-5 text-green-400" />
                <h3 className="text-lg font-semibold text-white">Impact Prediction</h3>
              </div>

              <div className="space-y-4">
                {/* Savings Badge */}
                {impactPrediction.savingsPercent > 5 ? (
                  <div className="bg-green-500/10 border border-green-500/20 rounded-lg p-4">
                    <div className="flex items-start gap-3">
                      <TrendingDown className="w-6 h-6 text-green-400 flex-shrink-0" />
                      <div>
                        <p className="text-2xl font-bold text-green-400">
                          ~{Math.round(impactPrediction.savingsPercent)}% CO₂ Savings
                        </p>
                        <p className="text-sm text-gray-300 mt-1">
                          By delaying until {formatTime(impactPrediction.optimalTime)}
                        </p>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
                    <div className="flex items-start gap-3">
                      <Zap className="w-6 h-6 text-blue-400 flex-shrink-0" />
                      <div>
                        <p className="text-lg font-bold text-blue-400">
                          Run Immediately
                        </p>
                        <p className="text-sm text-gray-300 mt-1">
                          Current grid is already optimal
                        </p>
                      </div>
                    </div>
                  </div>
                )}

                {/* Details Grid */}
                <div className="grid grid-cols-2 gap-4">
                  <div className="bg-karbos-dark-gray/50 rounded-lg p-3">
                    <p className="text-xs text-gray-400 mb-1">Current Intensity</p>
                    <p className="text-xl font-bold text-orange-400">
                      {Math.round(impactPrediction.currentIntensity)}
                    </p>
                    <p className="text-xs text-gray-500">gCO₂/kWh</p>
                  </div>

                  <div className="bg-karbos-dark-gray/50 rounded-lg p-3">
                    <p className="text-xs text-gray-400 mb-1">Optimal Intensity</p>
                    <p className="text-xl font-bold text-green-400">
                      {Math.round(impactPrediction.optimalIntensity)}
                    </p>
                    <p className="text-xs text-gray-500">gCO₂/kWh</p>
                  </div>

                  <div className="bg-karbos-dark-gray/50 rounded-lg p-3">
                    <p className="text-xs text-gray-400 mb-1">Estimated Savings</p>
                    <p className="text-xl font-bold text-purple-400">
                      {Math.round(impactPrediction.co2Savings)}g
                    </p>
                    <p className="text-xs text-gray-500">CO₂ equivalent</p>
                  </div>

                  <div className="bg-karbos-dark-gray/50 rounded-lg p-3">
                    <p className="text-xs text-gray-400 mb-1">Optimal Start</p>
                    <p className="text-sm font-bold text-white">
                      {impactPrediction.optimalTime.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                    </p>
                    <p className="text-xs text-gray-500">{impactPrediction.optimalTime.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Server Prediction (from dry-run) */}
          {prediction && (
            <div className="bg-karbos-dark-gray border border-karbos-blue-purple/20 rounded-lg p-6">
              <div className="flex items-center gap-2 mb-4">
                <CheckCircle className="w-5 h-5 text-karbos-blue-purple" />
                <h3 className="text-lg font-semibold text-white">Server Calculation</h3>
              </div>

              <div className="space-y-3 text-sm">
                <div className="flex justify-between items-center pb-2 border-b border-gray-700">
                  <span className="text-gray-400">Execution Mode:</span>
                  <span className={`font-semibold ${prediction.immediate ? 'text-yellow-400' : 'text-green-400'}`}>
                    {prediction.immediate ? 'Immediate' : 'Delayed (Optimized)'}
                  </span>
                </div>

                <div className="flex justify-between items-center pb-2 border-b border-gray-700">
                  <span className="text-gray-400">Scheduled Time:</span>
                  <span className="font-semibold text-white">
                    {new Date(prediction.scheduled_time).toLocaleTimeString()}
                  </span>
                </div>

                {prediction.expected_intensity && (
                  <div className="flex justify-between items-center pb-2 border-b border-gray-700">
                    <span className="text-gray-400">Expected Intensity:</span>
                    <span className="font-semibold text-purple-400">
                      {Math.round(prediction.expected_intensity)} gCO₂/kWh
                    </span>
                  </div>
                )}

                {prediction.carbon_savings && prediction.carbon_savings > 0 && (
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400">Carbon Savings:</span>
                    <span className="font-semibold text-green-400">
                      {Math.round(prediction.carbon_savings)} gCO₂
                    </span>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Info Box */}
          <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
            <p className="text-sm text-gray-300">
              <strong className="text-blue-400">How it works:</strong> The scheduler analyzes the carbon intensity forecast for your region and deadline. Jobs are automatically delayed to cleaner energy windows when possible, reducing your carbon footprint without compromising SLAs.
            </p>
          </div>
        </motion.div>
      </div>
    </motion.div>
  );
};

export default Playground;
