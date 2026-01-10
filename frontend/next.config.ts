/** @type {import('next').NextConfig} */
const nextConfig = {
  turbopack: {},
  webpack: (config: any) => {
    config.watchOptions = {
      poll: 300,
      aggregateTimeout: 300,
    };
    return config;
  },
};

export default nextConfig;
