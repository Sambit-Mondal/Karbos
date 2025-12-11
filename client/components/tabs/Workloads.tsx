"use client";

import React, { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import useSWR from "swr";
import { apiClient } from "@/lib/api";
import type { Job } from "@/lib/types";
import { RefreshCw, Play, Clock, CheckCircle, XCircle, Loader } from "lucide-react";

const Workloads = () => {
    const [selectedJob, setSelectedJob] = useState<Job | null>(null);
    
    // Fetch jobs with SWR (auto-refresh every 5 seconds)
    const { data: jobs, error, isLoading, mutate } = useSWR<Job[]>(
        '/api/jobs',
        () => apiClient.getJobs(),
        {
            refreshInterval: 5000,
            revalidateOnFocus: true,
        }
    );

    const handleRefresh = () => {
        mutate();
    };

    const getStatusIcon = (status: string) => {
        switch (status) {
            case 'RUNNING':
                return <Play className="w-4 h-4" />;
            case 'PENDING':
            case 'DELAYED':
                return <Clock className="w-4 h-4" />;
            case 'COMPLETED':
                return <CheckCircle className="w-4 h-4" />;
            case 'FAILED':
                return <XCircle className="w-4 h-4" />;
            default:
                return <Loader className="w-4 h-4" />;
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'RUNNING':
                return 'text-blue-400 bg-blue-900/20 border-blue-500/30';
            case 'PENDING':
            case 'DELAYED':
                return 'text-yellow-400 bg-yellow-900/20 border-yellow-500/30';
            case 'COMPLETED':
                return 'text-green-400 bg-green-900/20 border-green-500/30';
            case 'FAILED':
                return 'text-red-400 bg-red-900/20 border-red-500/30';
            default:
                return 'text-gray-400 bg-gray-900/20 border-gray-500/30';
        }
    };

    const formatDate = (dateString?: string) => {
        if (!dateString) return 'N/A';
        const date = new Date(dateString);
        return date.toLocaleString('en-US', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    if (error) {
        return (
            <div className="bg-red-900/20 border border-red-500/30 rounded-lg p-4">
                <p className="text-red-400">Failed to load jobs: {error.message}</p>
                <button
                    onClick={handleRefresh}
                    className="mt-2 px-4 py-2 bg-red-500 hover:bg-red-600 rounded-md text-white transition-colors"
                >
                    Retry
                </button>
            </div>
        );
    }

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
                    <h2 className="text-2xl font-bold text-karbos-light-blue">Job Queue</h2>
                    <p className="text-sm text-gray-400 mt-1">
                        {isLoading ? 'Loading...' : `${jobs?.length || 0} jobs total`}
                    </p>
                </div>
                <div className="flex space-x-2">
                    <button
                        onClick={handleRefresh}
                        disabled={isLoading}
                        className="px-4 py-2 bg-karbos-blue-purple text-white rounded-md hover:bg-opacity-80 transition-colors flex items-center gap-2 disabled:opacity-50"
                    >
                        <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
                        Refresh
                    </button>
                </div>
            </div>

            {/* Loading State */}
            {isLoading && !jobs && (
                <div className="flex justify-center items-center py-12">
                    <Loader className="w-8 h-8 text-karbos-blue-purple animate-spin" />
                </div>
            )}

            {/* Jobs List */}
            {jobs && jobs.length === 0 && (
                <div className="bg-karbos-indigo/50 border border-karbos-blue-purple/30 rounded-lg p-8 text-center">
                    <p className="text-gray-400">No jobs found</p>
                </div>
            )}

            {jobs && jobs.length > 0 && (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {jobs.map((job) => (
                        <motion.div
                            key={job.id}
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            whileHover={{ scale: 1.02 }}
                            onClick={() => setSelectedJob(job)}
                            className="bg-karbos-indigo rounded-lg p-4 border border-karbos-blue-purple/50 hover:border-karbos-blue-purple cursor-pointer transition-all"
                        >
                            {/* Job Header */}
                            <div className="flex justify-between items-start mb-3">
                                <div className="flex-1">
                                    <p className="text-xs text-gray-500 font-mono">
                                        {job.id.slice(0, 13)}...
                                    </p>
                                    <h3 className="text-sm font-semibold text-karbos-light-blue mt-1 truncate">
                                        {job.docker_image}
                                    </h3>
                                </div>
                                <div className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs border ${getStatusColor(job.status)}`}>
                                    {getStatusIcon(job.status)}
                                    <span>{job.status}</span>
                                </div>
                            </div>

                            {/* Job Details */}
                            <div className="space-y-2 text-xs text-gray-400">
                                <div className="flex justify-between">
                                    <span>User:</span>
                                    <span className="text-gray-300">{job.user_id}</span>
                                </div>
                                <div className="flex justify-between">
                                    <span>Submitted:</span>
                                    <span className="text-gray-300">{formatDate(job.created_at)}</span>
                                </div>
                                {job.started_at && (
                                    <div className="flex justify-between">
                                        <span>Started:</span>
                                        <span className="text-gray-300">{formatDate(job.started_at)}</span>
                                    </div>
                                )}
                                {job.completed_at && (
                                    <div className="flex justify-between">
                                        <span>Completed:</span>
                                        <span className="text-gray-300">{formatDate(job.completed_at)}</span>
                                    </div>
                                )}
                                <div className="flex justify-between">
                                    <span>Deadline:</span>
                                    <span className="text-gray-300">{formatDate(job.deadline)}</span>
                                </div>
                                {job.region && (
                                    <div className="flex justify-between">
                                        <span>Region:</span>
                                        <span className="text-gray-300">{job.region}</span>
                                    </div>
                                )}
                            </div>

                            {/* Progress Indicator */}
                            {job.status === 'RUNNING' && (
                                <div className="mt-3 pt-3 border-t border-karbos-blue-purple/30">
                                    <div className="flex items-center gap-2">
                                        <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                                        <span className="text-xs text-green-400">Executing...</span>
                                    </div>
                                </div>
                            )}
                        </motion.div>
                    ))}
                </div>
            )}

            {/* Job Details Modal */}
            <AnimatePresence>
                {selectedJob && (
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        onClick={() => setSelectedJob(null)}
                        className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4"
                    >
                        <motion.div
                            initial={{ scale: 0.9, opacity: 0 }}
                            animate={{ scale: 1, opacity: 1 }}
                            exit={{ scale: 0.9, opacity: 0 }}
                            onClick={(e) => e.stopPropagation()}
                            className="bg-karbos-indigo border border-karbos-blue-purple rounded-lg max-w-2xl w-full max-h-[80vh] overflow-auto"
                        >
                            {/* Modal Header */}
                            <div className="p-6 border-b border-karbos-blue-purple/30">
                                <div className="flex justify-between items-start">
                                    <div>
                                        <h3 className="text-xl font-bold text-karbos-light-blue">
                                            Job Details
                                        </h3>
                                        <p className="text-sm text-gray-400 font-mono mt-1">
                                            {selectedJob.id}
                                        </p>
                                    </div>
                                    <button
                                        onClick={() => setSelectedJob(null)}
                                        className="text-gray-400 hover:text-white transition-colors"
                                    >
                                        ✕
                                    </button>
                                </div>
                            </div>

                            {/* Modal Body */}
                            <div className="p-6 space-y-4">
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <p className="text-xs text-gray-500">Docker Image</p>
                                        <p className="text-sm text-white mt-1">{selectedJob.docker_image}</p>
                                    </div>
                                    <div>
                                        <p className="text-xs text-gray-500">Status</p>
                                        <div className="mt-1">
                                            <span className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs border ${getStatusColor(selectedJob.status)}`}>
                                                {getStatusIcon(selectedJob.status)}
                                                {selectedJob.status}
                                            </span>
                                        </div>
                                    </div>
                                    <div>
                                        <p className="text-xs text-gray-500">User ID</p>
                                        <p className="text-sm text-white mt-1">{selectedJob.user_id}</p>
                                    </div>
                                    <div>
                                        <p className="text-xs text-gray-500">Region</p>
                                        <p className="text-sm text-white mt-1">{selectedJob.region || 'N/A'}</p>
                                    </div>
                                    <div>
                                        <p className="text-xs text-gray-500">Created</p>
                                        <p className="text-sm text-white mt-1">{formatDate(selectedJob.created_at)}</p>
                                    </div>
                                    <div>
                                        <p className="text-xs text-gray-500">Deadline</p>
                                        <p className="text-sm text-white mt-1">{formatDate(selectedJob.deadline)}</p>
                                    </div>
                                    {selectedJob.started_at && (
                                        <div>
                                            <p className="text-xs text-gray-500">Started</p>
                                            <p className="text-sm text-white mt-1">{formatDate(selectedJob.started_at)}</p>
                                        </div>
                                    )}
                                    {selectedJob.completed_at && (
                                        <div>
                                            <p className="text-xs text-gray-500">Completed</p>
                                            <p className="text-sm text-white mt-1">{formatDate(selectedJob.completed_at)}</p>
                                        </div>
                                    )}
                                </div>

                                {/* Command */}
                                {selectedJob.command && (
                                    <div>
                                        <p className="text-xs text-gray-500">Command</p>
                                        <pre className="mt-1 p-3 bg-karbos-navy rounded text-xs text-gray-300 overflow-x-auto">
                                            {selectedJob.command}
                                        </pre>
                                    </div>
                                )}

                                {/* Metadata */}
                                {selectedJob.metadata && (
                                    <div>
                                        <p className="text-xs text-gray-500">Metadata</p>
                                        <pre className="mt-1 p-3 bg-karbos-navy rounded text-xs text-gray-300 overflow-x-auto">
                                            {JSON.stringify(JSON.parse(selectedJob.metadata), null, 2)}
                                        </pre>
                                    </div>
                                )}
                            </div>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>
        </motion.div>
    );
};

export default Workloads;
