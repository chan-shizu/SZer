import type { CreateRequestBody, CreateRequestResponse } from "./requests";

export async function postRequestClient(body: CreateRequestBody): Promise<CreateRequestResponse> {
  const res = await fetch("/api/request", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    credentials: "include",
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(text || `Request failed with status ${res.status}`);
  }

  return (await res.json()) as CreateRequestResponse;
}
