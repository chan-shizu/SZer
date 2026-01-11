import "server-only";

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
