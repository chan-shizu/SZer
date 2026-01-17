/** @type {import('next').NextConfig} */
const nextConfig = {
  turbopack: {},
  images: {
    // In local Docker dev, MinIO (and host.docker.internal) resolves to private IPs.
    // Next.js blocks these by default for SSRF protection.
    // Enable only in development.
    dangerouslyAllowLocalIP: process.env.NODE_ENV === "development",
    remotePatterns: [
      { protocol: "http", hostname: "host.docker.internal", port: "9000", pathname: "/**" },
      { protocol: "http", hostname: "localhost", port: "9000", pathname: "/**" },
      { protocol: "http", hostname: "127.0.0.1", port: "9000", pathname: "/**" },
      { protocol: "http", hostname: process.env.S3_VIDEO_BUCKET_HOST, pathname: "/**" },
      { protocol: "https", hostname: process.env.S3_VIDEO_BUCKET_HOST, pathname: "/**" },
      { protocol: "http", hostname: process.env.S3_PUBLIC_FILE_BUCKET_HOST, pathname: "/**" },
      { protocol: "https", hostname: process.env.S3_PUBLIC_FILE_BUCKET_HOST, pathname: "/**" },
    ],
  },
  webpack: (config: { watchOptions?: { poll?: number; aggregateTimeout?: number } }) => {
    config.watchOptions = {
      poll: 300,
      aggregateTimeout: 300,
    };
    return config;
  },
};

export default nextConfig;
