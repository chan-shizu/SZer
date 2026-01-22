/** @type {import('next').NextConfig} */
const toRemotePatterns = (value: string | undefined, opts?: { pathname?: string }) => {
  if (!value) return [];

  const pathname = opts?.pathname ?? "/**";

  // Accept formats:
  // - https://example.com:9000/public-file
  // - http://localhost:9000
  // - localhost:9000
  // - example.com
  const raw = value.trim();
  if (!raw) return [];

  const add = (protocol: "http" | "https", hostname: string, port?: string) => {
    if (!hostname) return [];
    const p = port?.trim();
    return p ? [{ protocol, hostname, port: p, pathname }] : [{ protocol, hostname, pathname }];
  };

  try {
    if (raw.startsWith("http://") || raw.startsWith("https://")) {
      const url = new URL(raw);
      const urlPath = url.pathname && url.pathname !== "/" ? `${url.pathname.replace(/\/$/, "")}/**` : pathname;
      const p = url.port || undefined;
      return p
        ? [{ protocol: url.protocol.replace(":", ""), hostname: url.hostname, port: p, pathname: urlPath }]
        : [{ protocol: url.protocol.replace(":", ""), hostname: url.hostname, pathname: urlPath }];
    }
  } catch {
    // fall through
  }

  // Handle host[:port][/path]
  const firstSlash = raw.indexOf("/");
  const hostPort = firstSlash >= 0 ? raw.slice(0, firstSlash) : raw;
  const hostPath = firstSlash >= 0 ? raw.slice(firstSlash) : "";
  const hostPathname = hostPath ? `${hostPath.replace(/\/$/, "")}/**` : pathname;

  const colonIdx = hostPort.lastIndexOf(":");
  const hasPort = colonIdx > 0 && colonIdx < hostPort.length - 1;
  const hostname = hasPort ? hostPort.slice(0, colonIdx) : hostPort;
  const port = hasPort ? hostPort.slice(colonIdx + 1) : undefined;

  // If protocol isn't given, allow both http/https.
  return [...add("http", hostname, port), ...add("https", hostname, port)].map((p) => ({
    ...p,
    pathname: hostPathname,
  }));
};

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
      // Back-end constructs URLs using *_ENDPOINT (can include protocol/port/path).
      ...toRemotePatterns(process.env.S3_PUBLIC_FILE_BUCKET_ENDPOINT),

      // Keep compatibility with older envs that provide only host.
      ...toRemotePatterns(process.env.S3_PUBLIC_FILE_BUCKET_HOST),

      // Not strictly needed for next/image today, but kept for parity.
      ...toRemotePatterns(process.env.S3_VIDEO_BUCKET_ENDPOINT),
      ...toRemotePatterns(process.env.S3_VIDEO_BUCKET_HOST),
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
