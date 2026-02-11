import { backendFetchJson } from "./server";
import type { GetProgramsResponse } from "./programs";

export async function getWatchingPrograms(): Promise<GetProgramsResponse> {
  try {
    return await backendFetchJson<GetProgramsResponse>("/me/watching-programs", { method: "GET", cache: "no-store" });
  } catch (err) {
    console.error(`[API通信エラー] getWatchingPrograms:`, { err });
    throw err;
  }
}

export async function getLikedPrograms(): Promise<GetProgramsResponse> {
  try {
    return await backendFetchJson<GetProgramsResponse>("/me/liked-programs", { method: "GET", cache: "no-store" });
  } catch (err) {
    console.error(`[API通信エラー] getLikedPrograms:`, { err });
    throw err;
  }
}

export async function getPurchasedPrograms(): Promise<GetProgramsResponse> {
  try {
    return await backendFetchJson<GetProgramsResponse>("/me/purchased-programs", { method: "GET", cache: "no-store" });
  } catch (err) {
    console.error(`[API通信エラー] getPurchasedPrograms:`, { err });
    throw err;
  }
}
