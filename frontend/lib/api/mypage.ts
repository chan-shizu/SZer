import { backendFetchJson } from "./server";
import type { GetProgramsResponse } from "./programs";

export async function getWatchingPrograms(): Promise<GetProgramsResponse> {
  return backendFetchJson<GetProgramsResponse>("/me/watching-programs", { method: "GET", cache: "no-store" });
}

export async function getLikedPrograms(): Promise<GetProgramsResponse> {
  return backendFetchJson<GetProgramsResponse>("/me/liked-programs", { method: "GET", cache: "no-store" });
}
