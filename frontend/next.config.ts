/** @type {import('next').NextConfig} */
const nextConfig = {
  turbopack: {},
  images: {
    remotePatterns: [
      { protocol: "http", hostname: process.env.S3_VIDEO_BUCKET_HOST, pathname: "/**" },
      { protocol: "https", hostname: process.env.S3_VIDEO_BUCKET_HOST, pathname: "/**" },
      { protocol: "http", hostname: process.env.S3_PUBLIC_FILE_BUCKET_HOST, pathname: "/**" },
      { protocol: "https", hostname: process.env.S3_PUBLIC_FILE_BUCKET_HOST, pathname: "/**" },
    ],
  },
  webpack: (config: any) => {
    config.watchOptions = {
      poll: 300,
      aggregateTimeout: 300,
    };
    return config;
  },
};

export default nextConfig;
