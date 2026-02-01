import "server-only";

import { headers as nextHeaders } from "next/headers";

import { ApiError } from "./error";

function normalizeBaseUrl(url: string): string {
  return url.replace(/\/+$/, "");
}

function assertLeadingSlash(path: string): void {
  if (!path.startsWith("/")) {
    throw new Error(`API path must start with "/": ${path}`);
  }
}

export function buildBackendUrl(path: string): string {
  assertLeadingSlash(path);
  const baseUrl = process.env.API_BASE_URL || "http://backend:8080";
  return `${normalizeBaseUrl(baseUrl)}${path}`;
}

export async function backendFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const url = buildBackendUrl(path);

  const headers = new Headers(init.headers);
  if (!headers.has("Accept")) {
    headers.set("Accept", "application/json");
  }

  // Forward incoming request cookies (better-auth session) to the backend.
  // This is required because Next server-to-server fetch does not automatically
  // include browser cookies.
  if (!headers.has("Cookie")) {
    try {
      const reqHeaders = await nextHeaders();
      const cookie = reqHeaders.get("cookie");
      if (cookie) {
        headers.set("Cookie", cookie);
      }
    } catch {
      // `headers()` throws when called outside a request context (e.g. build time).
    }
  }

  return fetch(url, {
    ...init,
    headers,
  });
}

export async function backendFetchJson<T>(path: string, init: RequestInit = {}): Promise<T> {
  const url = buildBackendUrl(path);
  const res = await backendFetch(path, init);

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    console.error(`[API通信エラー] backendFetchJson:`, {
      path,
      url,
      status: res.status,
      response: text,
      init,
    });
    throw new ApiError({
      status: res.status,
      code: `HTTP_${res.status}`,
      message: text || `Request failed with status ${res.status}`,
      url,
      details: text ? { body: text } : undefined,
    });
  }

  return res.json() as Promise<T>;
}
