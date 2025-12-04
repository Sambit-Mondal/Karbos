"use client";

import React, { useState } from "react";
import { motion } from "framer-motion";

interface JobSubmission {
  dockerImage: string;
  command: string;
  estimatedDuration: number;
  slaDeadline: number;
  simulationMode: boolean;
}

interface SimulationResult {
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

const Playground = () => {
  const [formData, setFormData] = useState<JobSubmission>({
    dockerImage: "",
    command: "",
    estimatedDuration: 30,
    slaDeadline: 4,
    simulationMode: true,
  });

  const [simulationResult, setSimulationResult] =
    useState<SimulationResult | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSliderChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData((prev) => ({
      ...prev,
      slaDeadline: parseInt(e.target.value),
    }));
  };

  const handleToggle = () => {
    setFormData((prev) => ({
      ...prev,
      simulationMode: !prev.simulationMode,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    // Simulate API call
    setTimeout(() => {
      if (formData.simulationMode) {
        // Mock simulation result
        const now = new Date("2025-12-04T14:45:00");
        const optimalTime = new Date(now);
        optimalTime.setHours(optimalTime.getHours() + 2);

        setSimulationResult({
          immediateExecution: {
            startTime: now.toISOString(),
            carbonIntensity: 450,
            estimatedCO2: 125,
          },
          optimizedExecution: {
            startTime: optimalTime.toISOString(),
            carbonIntensity: 280,
            estimatedCO2: 78,
            delaySavings: 38,
          },
        });
      } else {
        // Mock job submission
        alert("Job submitted successfully! Check the Workloads tab.");
        setFormData({
          dockerImage: "",
          command: "",
          estimatedDuration: 30,
          slaDeadline: 4,
          simulationMode: true,
        });
      }
      setIsSubmitting(false);
    }, 1500);
  };

  const presetJobs = [
    {
      name: "Python Data Processing",
      image: "python:3.11-slim",
      command: "python process_data.py",
      duration: 45,
    },
    {
      name: "Node.js Build",
      image: "node:18-alpine",
      command: "npm run build",
      duration: 20,
    },
    {
      name: "Docker Image Build",
      image: "docker:24-dind",
      command: "docker build -t myapp:latest .",
      duration: 60,
    },
  ];

  const loadPreset = (preset: (typeof presetJobs)[0]) => {
    setFormData((prev) => ({
      ...prev,
      dockerImage: preset.image,
      command: preset.command,
      estimatedDuration: preset.duration,
    }));
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
          Playground
        </h2>
        <p className="text-karbos-lavender mt-1">
          Test job submissions and see optimization in action
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Job Submission Form */}
        <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
          <h3 className="text-xl font-semibold text-karbos-light-blue mb-4">
            Submit Job
          </h3>

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Docker Image */}
            <div>
              <label className="block text-sm font-medium text-karbos-lavender mb-2">
                Docker Image
              </label>
              <input
                type="text"
                name="dockerImage"
                value={formData.dockerImage}
                onChange={handleInputChange}
                placeholder="e.g., python:3.11-slim"
                required
                className="w-full px-4 py-2 bg-karbos-navy border border-karbos-blue-purple rounded-md text-karbos-light-blue placeholder-karbos-lavender/50 focus:outline-none focus:ring-2 focus:ring-karbos-lavender"
              />
            </div>

            {/* Command */}
            <div>
              <label className="block text-sm font-medium text-karbos-lavender mb-2">
                Command
              </label>
              <textarea
                name="command"
                value={formData.command}
                onChange={handleInputChange}
                placeholder="e.g., python script.py --arg value"
                required
                rows={3}
                className="w-full px-4 py-2 bg-karbos-navy border border-karbos-blue-purple rounded-md text-karbos-light-blue placeholder-karbos-lavender/50 focus:outline-none focus:ring-2 focus:ring-karbos-lavender"
              />
            </div>

            {/* Estimated Duration */}
            <div>
              <label className="block text-sm font-medium text-karbos-lavender mb-2">
                Estimated Duration (minutes)
              </label>
              <input
                type="number"
                name="estimatedDuration"
                value={formData.estimatedDuration}
                onChange={handleInputChange}
                min="1"
                max="180"
                required
                className="w-full px-4 py-2 bg-karbos-navy border border-karbos-blue-purple rounded-md text-karbos-light-blue focus:outline-none focus:ring-2 focus:ring-karbos-lavender"
              />
            </div>

            {/* SLA Deadline Slider */}
            <div>
              <label className="block text-sm font-medium text-karbos-lavender mb-2">
                SLA Deadline: Must finish within {formData.slaDeadline} hours
              </label>
              <input
                type="range"
                name="slaDeadline"
                value={formData.slaDeadline}
                onChange={handleSliderChange}
                min="1"
                max="24"
                className="w-full h-2 bg-karbos-navy rounded-lg appearance-none cursor-pointer accent-karbos-blue-purple"
              />
              <div className="flex justify-between text-xs text-karbos-lavender mt-1">
                <span>1h</span>
                <span>6h</span>
                <span>12h</span>
                <span>18h</span>
                <span>24h</span>
              </div>
            </div>

            {/* Simulation Mode Toggle */}
            <div className="flex items-center justify-between p-4 bg-karbos-navy rounded-md">
              <div>
                <p className="text-karbos-light-blue font-medium">
                  Simulation Mode
                </p>
                <p className="text-xs text-karbos-lavender mt-1">
                  {formData.simulationMode
                    ? "Calculate optimal time without execution"
                    : "Actually submit and execute the job"}
                </p>
              </div>
              <button
                type="button"
                onClick={handleToggle}
                className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                  formData.simulationMode
                    ? "bg-karbos-blue-purple"
                    : "bg-karbos-lavender/30"
                }`}
              >
                <span
                  className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    formData.simulationMode ? "translate-x-6" : "translate-x-1"
                  }`}
                />
              </button>
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full px-4 py-3 bg-karbos-blue-purple text-white rounded-md font-semibold hover:bg-opacity-80 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              {isSubmitting
                ? "Processing..."
                : formData.simulationMode
                  ? "Run Simulation"
                  : "Submit Job"}
            </button>
          </form>

          {/* Preset Jobs */}
          <div className="mt-6">
            <h4 className="text-sm font-medium text-karbos-lavender mb-3">
              Quick Presets
            </h4>
            <div className="space-y-2">
              {presetJobs.map((preset, index) => (
                <button
                  key={index}
                  onClick={() => loadPreset(preset)}
                  className="w-full text-left px-4 py-2 bg-karbos-navy rounded-md hover:bg-karbos-blue-purple/20 transition-colors"
                >
                  <p className="text-karbos-light-blue text-sm font-medium">
                    {preset.name}
                  </p>
                  <p className="text-karbos-lavender text-xs mt-1">
                    {preset.image} • {preset.duration}min
                  </p>
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Simulation Results */}
        <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
          <h3 className="text-xl font-semibold text-karbos-light-blue mb-4">
            Optimization Analysis
          </h3>

          {simulationResult ? (
            <div className="space-y-6">
              {/* Immediate Execution */}
              <div className="bg-karbos-navy p-4 rounded-lg border border-red-500/30">
                <div className="flex items-center justify-between mb-3">
                  <h4 className="text-red-400 font-semibold">
                    Immediate Execution
                  </h4>
                  <span className="text-xs text-karbos-lavender">
                    (No optimization)
                  </span>
                </div>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-karbos-lavender">Start Time:</span>
                    <span className="text-karbos-light-blue">
                      {new Date(
                        simulationResult.immediateExecution.startTime,
                      ).toLocaleString()}
                    </span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-karbos-lavender">
                      Carbon Intensity:
                    </span>
                    <span className="text-red-400 font-semibold">
                      {simulationResult.immediateExecution.carbonIntensity}{" "}
                      gCO₂/kWh
                    </span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-karbos-lavender">Estimated CO₂:</span>
                    <span className="text-karbos-light-blue font-semibold">
                      {simulationResult.immediateExecution.estimatedCO2}g
                    </span>
                  </div>
                </div>
              </div>

              {/* Optimized Execution */}
              <div className="bg-karbos-navy p-4 rounded-lg border border-green-500/30">
                <div className="flex items-center justify-between mb-3">
                  <h4 className="text-green-400 font-semibold">
                    Optimized Execution
                  </h4>
                  <span className="text-xs text-green-400 bg-green-500/20 px-2 py-1 rounded">
                    Recommended
                  </span>
                </div>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-karbos-lavender">Start Time:</span>
                    <span className="text-karbos-light-blue">
                      {new Date(
                        simulationResult.optimizedExecution.startTime,
                      ).toLocaleString()}
                    </span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-karbos-lavender">
                      Carbon Intensity:
                    </span>
                    <span className="text-green-400 font-semibold">
                      {simulationResult.optimizedExecution.carbonIntensity}{" "}
                      gCO₂/kWh
                    </span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-karbos-lavender">Estimated CO₂:</span>
                    <span className="text-karbos-light-blue font-semibold">
                      {simulationResult.optimizedExecution.estimatedCO2}g
                    </span>
                  </div>
                </div>
              </div>

              {/* Savings Summary */}
              <div className="bg-gradient-to-r from-green-500/20 to-cyan-500/20 p-6 rounded-lg border border-green-500/50">
                <div className="text-center">
                  <p className="text-karbos-lavender text-sm mb-2">
                    Carbon Savings
                  </p>
                  <p className="text-4xl font-bold text-green-400 mb-1">
                    {simulationResult.optimizedExecution.delaySavings}%
                  </p>
                  <p className="text-karbos-lavender text-sm">
                    {simulationResult.immediateExecution.estimatedCO2 -
                      simulationResult.optimizedExecution.estimatedCO2}
                    g CO₂ saved
                  </p>
                  <div className="mt-4 pt-4 border-t border-karbos-lavender/20">
                    <p className="text-xs text-karbos-lavender">
                      By delaying execution by{" "}
                      {Math.round(
                        (new Date(
                          simulationResult.optimizedExecution.startTime,
                        ).getTime() -
                          new Date(
                            simulationResult.immediateExecution.startTime,
                          ).getTime()) /
                          3600000,
                      )}{" "}
                      hours, we can reduce carbon emissions significantly while
                      staying within SLA.
                    </p>
                  </div>
                </div>
              </div>

              {/* Visualization */}
              <div className="bg-karbos-navy p-4 rounded-lg">
                <h5 className="text-karbos-lavender text-sm mb-3">
                  Intensity Comparison
                </h5>
                <div className="space-y-3">
                  <div>
                    <div className="flex justify-between text-xs text-karbos-lavender mb-1">
                      <span>Immediate</span>
                      <span>
                        {simulationResult.immediateExecution.carbonIntensity}{" "}
                        gCO₂/kWh
                      </span>
                    </div>
                    <div className="w-full bg-karbos-indigo rounded-full h-3">
                      <div
                        className="h-3 bg-red-500 rounded-full"
                        style={{
                          width: `${(simulationResult.immediateExecution.carbonIntensity / 500) * 100}%`,
                        }}
                      />
                    </div>
                  </div>
                  <div>
                    <div className="flex justify-between text-xs text-karbos-lavender mb-1">
                      <span>Optimized</span>
                      <span>
                        {simulationResult.optimizedExecution.carbonIntensity}{" "}
                        gCO₂/kWh
                      </span>
                    </div>
                    <div className="w-full bg-karbos-indigo rounded-full h-3">
                      <div
                        className="h-3 bg-green-500 rounded-full"
                        style={{
                          width: `${(simulationResult.optimizedExecution.carbonIntensity / 500) * 100}%`,
                        }}
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center h-96 text-center">
              <svg
                className="w-24 h-24 text-karbos-lavender/30 mb-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                />
              </svg>
              <h4 className="text-karbos-lavender text-lg font-medium mb-2">
                No Simulation Yet
              </h4>
              <p className="text-karbos-lavender/70 text-sm max-w-sm">
                Fill out the form and run a simulation to see how much carbon
                can be saved by optimizing job execution timing.
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Info Box */}
      <div className="bg-karbos-indigo p-4 rounded-lg border border-karbos-blue-purple">
        <div className="flex items-start space-x-3">
          <svg
            className="w-5 h-5 text-karbos-blue-purple mt-0.5"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path
              fillRule="evenodd"
              d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
              clipRule="evenodd"
            />
          </svg>
          <div>
            <p className="text-sm text-karbos-lavender">
              <span className="font-semibold text-karbos-light-blue">Tip:</span>{" "}
              Use simulation mode to test the optimization algorithm without
              actually executing jobs. This is perfect for demonstrations and
              understanding how carbon-aware scheduling works.
            </p>
          </div>
        </div>
      </div>
    </motion.div>
  );
};

export default Playground;
