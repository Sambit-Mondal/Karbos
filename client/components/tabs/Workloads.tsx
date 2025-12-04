"use client";

import React, { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";

interface Job {
    id: string;
    image: string;
    status: "RUNNING" | "DELAYED" | "COMPLETED" | "FAILED";
    carbonImpact?: number;
    slaDeadline: string;
    submittedAt: string;
    startTime?: string;
    estimatedStart?: string;
    logs?: string;
}

const Workloads = () => {
    const [selectedJob, setSelectedJob] = useState<Job | null>(null);

    // Mock data
    const jobs: Job[] = [
        {
            id: "abc-123-def",
            image: "python:3.9-script",
            status: "RUNNING",
            slaDeadline: "2025-12-04T18:00:00",
            submittedAt: "2025-12-04T12:30:00",
            startTime: "2025-12-04T14:00:00",
            logs: "Starting job execution...\nProcessing data...\n[INFO] Progress: 45%",
        },
        {
            id: "def-456-ghi",
            image: "node:18-alpine",
            status: "DELAYED",
            slaDeadline: "2025-12-04T20:00:00",
            submittedAt: "2025-12-04T13:00:00",
            estimatedStart: "2025-12-04T15:30:00",
            logs: "Job queued for optimal execution window",
        },
        {
            id: "ghi-789-jkl",
            image: "ubuntu:22.04",
            status: "COMPLETED",
            carbonImpact: 34,
            slaDeadline: "2025-12-04T16:00:00",
            submittedAt: "2025-12-04T11:00:00",
            startTime: "2025-12-04T13:00:00",
            logs: "Job completed successfully\n[SUCCESS] Output saved to /results",
        },
        {
            id: "jkl-012-mno",
            image: "postgres:15",
            status: "FAILED",
            slaDeadline: "2025-12-04T15:00:00",
            submittedAt: "2025-12-04T12:00:00",
            startTime: "2025-12-04T14:30:00",
            logs: "[ERROR] Connection timeout\n[ERROR] Job failed after 3 retries",
        },
        {
            id: "mno-345-pqr",
            image: "python:3.11-slim",
            status: "DELAYED",
            slaDeadline: "2025-12-05T10:00:00",
            submittedAt: "2025-12-04T14:00:00",
            estimatedStart: "2025-12-04T16:00:00",
            logs: "Waiting for green window - Starts in 45m",
        },
    ];

    const getStatusBadge = (status: Job["status"]) => {
        const styles = {
            RUNNING: "bg-green-500/20 text-green-400 border-green-500",
            DELAYED: "bg-yellow-500/20 text-yellow-400 border-yellow-500",
            COMPLETED: "bg-blue-500/20 text-blue-400 border-blue-500",
            FAILED: "bg-red-500/20 text-red-400 border-red-500",
        };

        return (
            <span
                className={`px-3 py-1 rounded-full text-xs font-medium border ${styles[status]}`}
            >
                {status}
            </span>
        );
    };

    const calculateTimeRemaining = (estimatedStart: string) => {
        const now = new Date("2025-12-04T14:45:00"); // Mock current time
        const start = new Date(estimatedStart);
        const diff = start.getTime() - now.getTime();
        const minutes = Math.floor(diff / 60000);
        const hours = Math.floor(minutes / 60);

        if (hours > 0) {
            return `Starts in ${hours}h ${minutes % 60}m`;
        }
        return `Starts in ${minutes}m`;
    };

    return (
        <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.5 }}
            className="space-y-6"
        >
            {/* Header */}
            <div className="flex justify-between items-center">
                <h2 className="text-2xl font-bold text-karbos-light-blue">Job Queue</h2>
                <div className="flex space-x-2">
                    <button className="px-4 py-2 bg-karbos-blue-purple text-white rounded-md hover:bg-opacity-80 transition-colors">
                        Refresh
                    </button>
                    <button className="px-4 py-2 bg-karbos-indigo text-karbos-light-blue rounded-md border border-karbos-blue-purple hover:bg-karbos-blue-purple transition-colors">
                        Filters
                    </button>
                </div>
            </div>

            {/* Jobs Table */}
            <div className="bg-karbos-indigo rounded-lg border border-karbos-blue-purple overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="w-full">
                        <thead className="bg-karbos-navy">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-karbos-lavender uppercase tracking-wider">
                                    Job ID
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-karbos-lavender uppercase tracking-wider">
                                    Image
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-karbos-lavender uppercase tracking-wider">
                                    Status
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-karbos-lavender uppercase tracking-wider">
                                    Carbon Impact
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-karbos-lavender uppercase tracking-wider">
                                    SLA Deadline
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-karbos-lavender uppercase tracking-wider">
                                    Actions
                                </th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-karbos-blue-purple">
                            {jobs.map((job, index) => (
                                <motion.tr
                                    key={job.id}
                                    initial={{ opacity: 0, x: -20 }}
                                    animate={{ opacity: 1, x: 0 }}
                                    transition={{ duration: 0.3, delay: index * 0.05 }}
                                    whileHover={{ backgroundColor: "rgba(41, 46, 111, 0.5)" }}
                                    className="cursor-pointer"
                                    onClick={() => setSelectedJob(job)}
                                >
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <span className="text-sm font-mono text-karbos-light-blue">
                                            {job.id}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4">
                                        <span className="text-sm text-karbos-lavender">
                                            {job.image}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <div className="flex flex-col space-y-1">
                                            {getStatusBadge(job.status)}
                                            {job.status === "DELAYED" && job.estimatedStart && (
                                                <span className="text-xs text-yellow-400">
                                                    {calculateTimeRemaining(job.estimatedStart)}
                                                </span>
                                            )}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        {job.carbonImpact !== undefined ? (
                                            <span className="text-sm text-green-400">
                                                ↓ {job.carbonImpact}% reduction
                                            </span>
                                        ) : (
                                            <span className="text-sm text-karbos-lavender">-</span>
                                        )}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <span className="text-sm text-karbos-lavender">
                                            {new Date(job.slaDeadline).toLocaleString()}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                setSelectedJob(job);
                                            }}
                                            className="text-karbos-blue-purple hover:text-karbos-lavender text-sm"
                                        >
                                            View Details
                                        </button>
                                    </td>
                                </motion.tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Job Details Drawer */}
            <AnimatePresence>
                {selectedJob && (
                    <>
                        <motion.div
                            initial={{ x: "100%" }}
                            animate={{ x: 0 }}
                            exit={{ x: "100%" }}
                            transition={{ type: "spring", damping: 25, stiffness: 200 }}
                            className="fixed inset-y-0 right-0 w-full md:w-1/2 lg:w-1/3 bg-karbos-navy border-l border-karbos-blue-purple shadow-xl z-50 overflow-y-auto"
                        >
                            <div className="p-6">
                                <div className="flex justify-between items-start mb-6">
                                    <div>
                                        <h3 className="text-xl font-bold text-karbos-light-blue">
                                            Job Details
                                        </h3>
                                        <p className="text-sm font-mono text-karbos-lavender mt-1">
                                            {selectedJob.id}
                                        </p>
                                    </div>
                                    <button
                                        onClick={() => setSelectedJob(null)}
                                        className="text-karbos-lavender hover:text-white"
                                    >
                                        <svg
                                            className="w-6 h-6"
                                            fill="none"
                                            stroke="currentColor"
                                            viewBox="0 0 24 24"
                                        >
                                            <path
                                                strokeLinecap="round"
                                                strokeLinejoin="round"
                                                strokeWidth={2}
                                                d="M6 18L18 6M6 6l12 12"
                                            />
                                        </svg>
                                    </button>
                                </div>

                                <div className="space-y-4">
                                    <div>
                                        <label className="text-sm text-karbos-lavender">Image</label>
                                        <p className="text-karbos-light-blue font-mono">
                                            {selectedJob.image}
                                        </p>
                                    </div>

                                    <div>
                                        <label className="text-sm text-karbos-lavender">Status</label>
                                        <div className="mt-1">{getStatusBadge(selectedJob.status)}</div>
                                    </div>

                                    {selectedJob.carbonImpact !== undefined && (
                                        <div>
                                            <label className="text-sm text-karbos-lavender">
                                                Carbon Impact
                                            </label>
                                            <p className="text-green-400 font-semibold">
                                                {selectedJob.carbonImpact}% reduction vs immediate execution
                                            </p>
                                        </div>
                                    )}

                                    <div>
                                        <label className="text-sm text-karbos-lavender">Timeline</label>
                                        <div className="mt-2 space-y-3">
                                            <div className="flex items-start space-x-3">
                                                <div className="w-2 h-2 rounded-full bg-blue-400 mt-1"></div>
                                                <div>
                                                    <p className="text-karbos-light-blue text-sm">
                                                        Submitted
                                                    </p>
                                                    <p className="text-xs text-karbos-lavender">
                                                        {new Date(selectedJob.submittedAt).toLocaleString()}
                                                    </p>
                                                </div>
                                            </div>
                                            {selectedJob.status === "DELAYED" &&
                                                selectedJob.estimatedStart && (
                                                    <div className="flex items-start space-x-3">
                                                        <div className="w-2 h-2 rounded-full bg-yellow-400 mt-1"></div>
                                                        <div>
                                                            <p className="text-karbos-light-blue text-sm">
                                                                Optimized Delay
                                                            </p>
                                                            <p className="text-xs text-karbos-lavender">
                                                                Scheduled for{" "}
                                                                {new Date(
                                                                    selectedJob.estimatedStart,
                                                                ).toLocaleString()}
                                                            </p>
                                                        </div>
                                                    </div>
                                                )}
                                            {selectedJob.startTime && (
                                                <div className="flex items-start space-x-3">
                                                    <div className="w-2 h-2 rounded-full bg-green-400 mt-1"></div>
                                                    <div>
                                                        <p className="text-karbos-light-blue text-sm">
                                                            Executed
                                                        </p>
                                                        <p className="text-xs text-karbos-lavender">
                                                            {new Date(selectedJob.startTime).toLocaleString()}
                                                        </p>
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                    </div>

                                    <div>
                                        <label className="text-sm text-karbos-lavender">
                                            SLA Deadline
                                        </label>
                                        <p className="text-karbos-light-blue">
                                            {new Date(selectedJob.slaDeadline).toLocaleString()}
                                        </p>
                                    </div>

                                    <div>
                                        <label className="text-sm text-karbos-lavender mb-2 block">
                                            Logs
                                        </label>
                                        <div className="bg-black rounded-md p-4 font-mono text-sm text-green-400 max-h-64 overflow-y-auto">
                                            <pre className="whitespace-pre-wrap">{selectedJob.logs}</pre>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </motion.div>

                        {/* Overlay */}
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            transition={{ duration: 0.2 }}
                            className="fixed inset-0 bg-black bg-opacity-50 z-40"
                            onClick={() => setSelectedJob(null)}
                        />
                    </>
                )}
            </AnimatePresence>
        </motion.div>
    );
};

export default Workloads;
