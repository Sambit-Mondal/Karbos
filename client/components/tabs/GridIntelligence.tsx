"use client";

import { getStatusColor } from "@/lib/utils";
import React, { useState, useMemo } from "react";

interface Region {
  id: string;
  name: string;
  intensity: number;
  status: "green" | "yellow" | "red";
}

interface ForecastData {
  timeWindow: string;
  wind: number;
  solar: number;
  gas: number;
  coal: number;
  nuclear: number;
  intensity: number;
}

const GridIntelligence = () => {
  const [selectedRegion, setSelectedRegion] = useState<string | null>(null);

  const regions: Region[] = useMemo(
    () => [
      { id: "US-EAST", name: "US East", intensity: 320, status: "green" },
      { id: "US-WEST", name: "US West", intensity: 380, status: "yellow" },
      { id: "EU-CENTRAL", name: "EU Central", intensity: 290, status: "green" },
      { id: "EU-NORTH", name: "EU North", intensity: 480, status: "red" },
      {
        id: "ASIA-PACIFIC",
        name: "Asia Pacific",
        intensity: 520,
        status: "red",
      },
      { id: "UK", name: "United Kingdom", intensity: 350, status: "yellow" },
    ],
    [],
  );

  const [forecastData] = useState<ForecastData[]>(() => {
    return Array.from({ length: 24 }, (_, i) => {
      const hour = new Date("2025-12-04T00:00:00");
      hour.setHours(hour.getHours() + i);

      const wind = Math.floor(Math.random() * 30) + 10;
      const solar = i >= 6 && i <= 18 ? Math.floor(Math.random() * 25) + 5 : 0;
      const gas = Math.floor(Math.random() * 20) + 25;
      const coal = Math.floor(Math.random() * 15) + 10;
      const nuclear = 100 - wind - solar - gas - coal;

      const intensity =
        (wind * 11 + solar * 48 + gas * 490 + coal * 820 + nuclear * 12) / 100;

      return {
        timeWindow: hour.toLocaleTimeString("en-US", {
          hour: "2-digit",
          minute: "2-digit",
        }),
        wind,
        solar,
        gas,
        coal,
        nuclear,
        intensity,
      };
    });
  });

  const getIntensityLabel = (intensity: number) => {
    if (intensity < 350) return { label: "Low", color: "text-green-400" };
    if (intensity < 450) return { label: "Medium", color: "text-yellow-400" };
    return { label: "High", color: "text-red-400" };
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-karbos-light-blue">
          Grid Intelligence
        </h2>
        <p className="text-karbos-lavender mt-1">
          Real-time carbon intensity data and forecasts across regions
        </p>
      </div>

      {/* Regional Map/Heatmap */}
      <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
        <h3 className="text-xl font-semibold text-karbos-light-blue mb-4">
          Regional Carbon Intensity Map
        </h3>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {regions.map((region) => {
            const { label, color } = getIntensityLabel(region.intensity);
            return (
              <div
                key={region.id}
                onClick={() => setSelectedRegion(region.id)}
                className={`p-4 rounded-lg border-2 cursor-pointer transition-all ${
                  selectedRegion === region.id
                    ? "border-karbos-light-blue bg-karbos-navy"
                    : "border-karbos-blue-purple hover:border-karbos-lavender"
                }`}
              >
                <div className="flex items-start justify-between mb-2">
                  <h4 className="text-karbos-light-blue font-semibold">
                    {region.name}
                  </h4>
                  <div
                    className={`w-3 h-3 rounded-full ${getStatusColor(region.status)}`}
                    title={`Status: ${region.status}`}
                  />
                </div>
                <p className="text-sm text-karbos-lavender mb-1">{region.id}</p>
                <div className="flex items-baseline space-x-2">
                  <span className="text-2xl font-bold text-karbos-light-blue">
                    {region.intensity}
                  </span>
                  <span className="text-sm text-karbos-lavender">gCO₂/kWh</span>
                </div>
                <p className={`text-sm font-medium mt-1 ${color}`}>
                  {label} Intensity
                </p>
              </div>
            );
          })}
        </div>

        <div className="mt-6 flex items-center justify-center space-x-6 text-sm">
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 bg-green-500 rounded-full"></div>
            <span className="text-karbos-lavender">Green (Low Carbon)</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 bg-yellow-500 rounded-full"></div>
            <span className="text-karbos-lavender">Yellow (Medium)</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 bg-red-500 rounded-full"></div>
            <span className="text-karbos-lavender">Red (High Carbon)</span>
          </div>
        </div>
      </div>

      {/* Generation Mix Chart */}
      <div className="bg-karbos-indigo p-6 rounded-lg border border-karbos-blue-purple">
        <h3 className="text-xl font-semibold text-karbos-light-blue mb-4">
          Current Generation Mix
        </h3>

        <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
          {[
            { name: "Wind", value: 25, color: "bg-cyan-500" },
            { name: "Solar", value: 15, color: "bg-yellow-500" },
            { name: "Gas", value: 35, color: "bg-orange-500" },
            { name: "Coal", value: 15, color: "bg-gray-700" },
            { name: "Nuclear", value: 10, color: "bg-purple-500" },
          ].map((source) => (
            <div key={source.name} className="text-center">
              <div className="relative w-24 h-24 mx-auto mb-2">
                <svg className="w-full h-full transform -rotate-90">
                  <circle
                    cx="48"
                    cy="48"
                    r="40"
                    className="fill-none stroke-karbos-navy"
                    strokeWidth="8"
                  />
                  <circle
                    cx="48"
                    cy="48"
                    r="40"
                    className={`fill-none ${source.color.replace("bg-", "stroke-")}`}
                    strokeWidth="8"
                    strokeDasharray={`${source.value * 2.51} ${251 - source.value * 2.51}`}
                  />
                </svg>
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className="text-xl font-bold text-karbos-light-blue">
                    {source.value}%
                  </span>
                </div>
              </div>
              <p className="text-karbos-lavender text-sm">{source.name}</p>
            </div>
          ))}
        </div>
      </div>

      {/* Forecast Table */}
      <div className="bg-karbos-indigo rounded-lg border border-karbos-blue-purple overflow-hidden">
        <div className="px-6 py-4 bg-karbos-navy">
          <h3 className="text-xl font-semibold text-karbos-light-blue">
            24-Hour Forecast Data
          </h3>
          <p className="text-sm text-karbos-lavender mt-1">
            Raw data used by the scheduler for optimization decisions
          </p>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-karbos-navy border-t border-karbos-blue-purple">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-karbos-lavender uppercase">
                  Time Window
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-karbos-lavender uppercase">
                  Wind %
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-karbos-lavender uppercase">
                  Solar %
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-karbos-lavender uppercase">
                  Gas %
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-karbos-lavender uppercase">
                  Coal %
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-karbos-lavender uppercase">
                  Nuclear %
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-karbos-lavender uppercase">
                  Intensity Score
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-karbos-blue-purple">
              {forecastData.map((data, index) => {
                const { color } = getIntensityLabel(data.intensity);
                return (
                  <tr
                    key={index}
                    className="hover:bg-karbos-navy transition-colors"
                  >
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-karbos-light-blue font-medium">
                      {data.timeWindow}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-cyan-400 text-right">
                      {data.wind}%
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-yellow-400 text-right">
                      {data.solar}%
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-orange-400 text-right">
                      {data.gas}%
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-400 text-right">
                      {data.coal}%
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-purple-400 text-right">
                      {data.nuclear}%
                    </td>
                    <td
                      className={`px-4 py-3 whitespace-nowrap text-sm font-semibold text-right ${color}`}
                    >
                      {Math.round(data.intensity)}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* Data Source Info */}
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
              Data sourced from{" "}
              <span className="text-karbos-light-blue font-semibold">
                Carbon Aware SDK
              </span>{" "}
              and regional grid operators. Updated every 5 minutes.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default GridIntelligence;
