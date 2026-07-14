import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // 1. Required for Tauri: Converts Next.js into static HTML/CSS/JS
  output: 'export', 
  
  // 2. Required for Tauri: Local desktop apps can't use Next.js image optimization
  images: {
    unoptimized: true,
  },

  webpack: (config) => {
    config.experiments = {
      ...config.experiments,
      asyncWebAssembly: true,
      layers: true,
    };
    return config;
  },
};

export default nextConfig;
