"use client";

import React, { useState } from "react";
import Navigation from "@/components/Navigation";
import Overview from "@/components/tabs/Overview";
import Workloads from "@/components/tabs/Workloads";
import GridIntelligence from "@/components/tabs/GridIntelligence";
import Infrastructure from "@/components/tabs/Infrastructure";
import Playground from "@/components/tabs/Playground";

const Home = () => {
  const [activeTab, setActiveTab] = useState("overview");

  const renderTab = () => {
    switch (activeTab) {
      case "overview":
        return <Overview />;
      case "workloads":
        return <Workloads />;
      case "grid":
        return <GridIntelligence />;
      case "infrastructure":
        return <Infrastructure />;
      case "playground":
        return <Playground />;
      default:
        return <Overview />;
    }
  };

  return (
    <div className="min-h-screen bg-karbos-navy">
      <Navigation activeTab={activeTab} onTabChange={setActiveTab} />
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {renderTab()}
      </main>
    </div>
  );
};

export default Home;
