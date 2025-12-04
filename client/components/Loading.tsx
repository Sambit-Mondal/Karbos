"use client";

import React from "react";

interface LoadingProps {
  message?: string;
}

const Loading: React.FC<LoadingProps> = ({ message = "Loading..." }) => {
  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] space-y-4">
      <div className="relative w-16 h-16">
        <div className="absolute inset-0 border-4 border-karbos-blue-purple border-t-transparent rounded-full animate-spin" />
        <div className="absolute inset-2 border-4 border-karbos-lavender border-t-transparent rounded-full animate-spin animation-delay-150" />
      </div>
      <p className="text-karbos-lavender text-lg">{message}</p>
    </div>
  );
};

export default Loading;
