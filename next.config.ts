import type { NextConfig } from "next";

const config: NextConfig = {
  output: "standalone",
  devIndicators: false,
  poweredByHeader: false,
  agentRules: false,
  serverExternalPackages: ["better-sqlite3", "sharp"],
  experimental: {
    serverActions: { bodySizeLimit: "22mb" },
    proxyClientMaxBodySize: "22mb",
  },
  async headers() {
    return [
      {
        source: "/static/:caminho*",
        headers: [{ key: "Cache-Control", value: "public, max-age=31536000, immutable" }],
      },
      {
        source: "/:caminho*",
        headers: [
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
        ],
      },
    ];
  },
};

export default config;
