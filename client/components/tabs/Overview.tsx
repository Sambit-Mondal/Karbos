"use client";

import React from "react";

// Generate deterministic mock data
const generateChartData = () => {
  const data = [];
  for (let i = 0; i < 24; i++) {
    data.push({
      hour: i,
      intensity: 300 + Math.sin(i / 3) * 100 + ((i * 7) % 50),
      isOptimal: (i >= 2 && i <= 6) || (i >= 14 && i <= 18),
    });
  }
  return data;
};

const Overview = () => {
  // Mock data
  const kpiData = {
    totalCO2Saved: 4.2,
    activeJobs: 12,
    pendingOptimized: 8,
    currentIntensity: 450,
    trend: "falling",
  };

  const chartData = generateChartData();

  return (
    <div className="space-y-6">
      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
          <h3 className="text-karbos-lavender text-sm font-medium">
            Total CO₂ Saved
          </h3>
          <p className="text-3xl font-bold text-karbos-light-blue mt-2">
            {kpiData.totalCO2Saved} kg
          </p>
          <p className="text-green-400 text-sm mt-1">↑ 12% from last week</p>
        </div>

        <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
          <h3 className="text-karbos-lavender text-sm font-medium">
            Active Jobs
          </h3>
          <p className="text-3xl font-bold text-karbos-light-blue mt-2">
            {kpiData.activeJobs}
          </p>
          <p className="text-karbos-lavender text-sm mt-1">Running now</p>
        </div>

        <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
          <h3 className="text-karbos-lavender text-sm font-medium">
            Pending (Optimized)
          </h3>
          <p className="text-3xl font-bold text-karbos-light-blue mt-2">
            {kpiData.pendingOptimized}
          </p>
          <p className="text-yellow-400 text-sm mt-1">
            Waiting for green window
          </p>
        </div>

        <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
          <h3 className="text-karbos-lavender text-sm font-medium">
            Current Grid Intensity
          </h3>
          <p className="text-3xl font-bold text-karbos-light-blue mt-2">
            {kpiData.currentIntensity}
          </p>
          <p
            className={`text-sm mt-1 ${
              kpiData.trend === "falling" ? "text-green-400" : "text-red-400"
            }`}
          >
            {kpiData.trend === "falling" ? "↓" : "↑"} gCO₂/kWh
          </p>
        </div>
      </div>

      {/* Eco-Curve Chart */}
      <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
        <h2 className="text-xl font-semibold text-karbos-light-blue mb-4">
          24-Hour Carbon Intensity Forecast (Eco-Curve)
        </h2>
        <div className="relative h-96 bg-karbos-navy/30 rounded-lg p-4">
          <svg className="w-full h-full" viewBox="0 0 1000 400">
            {/* Background */}
            <rect
              x="80"
              y="20"
              width="910"
              height="360"
              className="fill-karbos-navy opacity-50"
            />

            {/* Y-axis labels */}
            <text
              x="30"
              y="30"
              className="text-xs fill-karbos-lavender font-medium"
            >
              500
            </text>
            <text
              x="30"
              y="130"
              className="text-xs fill-karbos-lavender font-medium"
            >
              400
            </text>
            <text
              x="30"
              y="230"
              className="text-xs fill-karbos-lavender font-medium"
            >
              300
            </text>
            <text
              x="30"
              y="330"
              className="text-xs fill-karbos-lavender font-medium"
            >
              200
            </text>

            {/* Optimal windows (continuous regions) */}
            <rect
              x={80 + 2 * 38}
              y="20"
              width={5 * 38}
              height="360"
              className="fill-green-500 opacity-10"
            />
            <rect
              x={80 + 14 * 38}
              y="20"
              width={5 * 38}
              height="360"
              className="fill-green-500 opacity-10"
            />

            {/* High carbon zones (only for high intensity points) */}
            {chartData.map((point, i) => {
              if (!point.isOptimal && point.intensity > 400) {
                return (
                  <rect
                    key={`high-${i}`}
                    x={80 + i * 38}
                    y="20"
                    width="38"
                    height="360"
                    className="fill-red-500 opacity-10"
                  />
                );
              }
              return null;
            })}

            {/* Grid lines */}
            {[0, 1, 2, 3, 4].map((i) => (
              <line
                key={`grid-${i}`}
                x1="80"
                y1={30 + i * 90}
                x2="990"
                y2={30 + i * 90}
                className="stroke-karbos-blue-purple opacity-30"
                strokeWidth="1"
              />
            ))}

            {/* Carbon intensity curve */}
            <polyline
              points={chartData
                .map((point, i) => {
                  const x = 80 + i * 38;
                  const y = 380 - ((point.intensity - 200) / 300) * 360;
                  return `${x},${y}`;
                })
                .join(" ")}
              className="fill-none stroke-karbos-lavender"
              strokeWidth="3"
            />

            {/* Data points */}
            {chartData.map((point, i) => {
              const x = 80 + i * 38;
              const y = 380 - ((point.intensity - 200) / 300) * 360;
              return (
                <circle
                  key={`point-${i}`}
                  cx={x}
                  cy={y}
                  r="4"
                  className="fill-karbos-light-blue cursor-pointer hover:r-6 transition-all"
                >
                  <title>
                    {`Hour ${point.hour}: ${Math.round(point.intensity)} gCO₂/kWh`}
                  </title>
                </circle>
              );
            })}

            {/* X-axis labels */}
            {[0, 6, 12, 18, 24].map((hour) => (
              <text
                key={`x-label-${hour}`}
                x={80 + hour * 38}
                y="395"
                className="text-xs fill-karbos-lavender text-anchor-middle"
                textAnchor="middle"
              >
                {hour}h
              </text>
            ))}
          </svg>

          <div className="mt-4 flex justify-center space-x-6 text-sm">
            <div className="flex items-center space-x-2">
              <div className="w-4 h-4 bg-green-900 opacity-40 rounded"></div>
              <span className="text-karbos-lavender">Optimal Windows</span>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-4 h-4 bg-red-900 opacity-40 rounded"></div>
              <span className="text-karbos-lavender">High Carbon Zones</span>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
        <h2 className="text-xl font-semibold text-karbos-light-blue mb-4">
          Recent Activity
        </h2>
        <div className="space-y-3">
          <div className="flex items-center justify-between p-3 bg-karbos-navy rounded">
            <div>
              <p className="text-karbos-light-blue font-medium">
                Job abc-123 completed
              </p>
              <p className="text-karbos-lavender text-sm">
                Saved 45g CO₂ by delaying 2 hours
              </p>
            </div>
            <span className="text-green-400 text-sm">2 min ago</span>
          </div>
          <div className="flex items-center justify-between p-3 bg-karbos-navy rounded">
            <div>
              <p className="text-karbos-light-blue font-medium">
                Job def-456 scheduled
              </p>
              <p className="text-karbos-lavender text-sm">
                Will run in optimal window (14:00)
              </p>
            </div>
            <span className="text-yellow-400 text-sm">5 min ago</span>
          </div>
          <div className="flex items-center justify-between p-3 bg-karbos-navy rounded">
            <div>
              <p className="text-karbos-light-blue font-medium">
                Worker node-03 connected
              </p>
              <p className="text-karbos-lavender text-sm">
                Available for job execution
              </p>
            </div>
            <span className="text-blue-400 text-sm">12 min ago</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Overview;
