import { betterAuth } from "better-auth";
import { nextCookies } from "better-auth/next-js";

import { pool } from "@/lib/auth/db";

export const auth = betterAuth({
  baseURL: process.env.BETTER_AUTH_URL ?? process.env.NEXT_PUBLIC_APP_URL,
  secret: process.env.BETTER_AUTH_SECRET ?? process.env.AUTH_SECRET,
  database: pool,
  emailAndPassword: {
    enabled: true,
  },
  plugins: [nextCookies()],
});
