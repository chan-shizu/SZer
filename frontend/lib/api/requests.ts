import { backendFetchJson } from "./server";

export type CreateRequestBody = {
  content: string;
  name: string;
  contact: string;
  note: string;
};

export type CreateRequestResponse = {
  request: {
    id: number;
    created_at: string;
  };
};

export async function postRequest(
  body: CreateRequestBody,
  cookie?: string,
): Promise<CreateRequestResponse> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (cookie) headers["Cookie"] = cookie;
  return backendFetchJson<CreateRequestResponse>("/requests", {
    method: "POST",
    headers,
    body: JSON.stringify(body),
  });
}
