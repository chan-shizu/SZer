import { backendFetchJson } from "./server";
import type { GetProgramsResponse } from "./programs";
import type { GetPointsResponse } from "./points";

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

export async function getMyPoints(): Promise<GetPointsResponse> {
  try {
    return await backendFetchJson<GetPointsResponse>("/me/points", { method: "GET", cache: "no-store" });
  } catch (err) {
    console.error(`[API通信エラー] getMyPoints:`, { err });
    throw err;
  }
}
