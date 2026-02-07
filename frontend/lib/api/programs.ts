export type ProgramDetailsCategoryTag = {
  id: number;
  name: string;
};

export type ProgramDetailsPerformer = {
  id: number;
  full_name: string;
  full_name_kana: string;
  image_url: string | null;
};

export type ProgramWatchHistory = {
  position_seconds: number;
  is_completed: boolean;
  last_watched_at: string;
};

export type ProgramDetails = {
  program_id: number;
  title: string;
  video_url: string;
  view_count: number;
  like_count: number;
  liked: boolean;
  is_limited_release: boolean;
  price: number;
  thumbnail_url: string | null;
  description: string | null;
  program_created_at: string;
  program_updated_at: string;
  category_tags: ProgramDetailsCategoryTag[];
  performers: ProgramDetailsPerformer[];
  watch_history: ProgramWatchHistory | null;
};

export type GetProgramDetailResponse = {
  program: ProgramDetails;
  is_permitted: boolean;
};

export type ProgramListItem = {
  program_id: number;
  title: string;
  view_count: number;
  like_count: number;
  is_limited_release: boolean;
  price: number;
  thumbnail_url: string | null;
  category_tags: ProgramDetailsCategoryTag[];
};

export type GetProgramsResponse = {
  programs: ProgramListItem[];
};

export type GetTopProgramsResponse = {
  programs: TopProgramItem[];
};

export type GetTopLikedProgramsResponse = {
  programs: TopProgramItem[];
};

export type GetTopViewedProgramsResponse = {
  programs: TopProgramItem[];
};

export type TopProgramItem = {
  program_id: number;
  title: string;
  view_count: number;
  like_count: number;
  is_limited_release: boolean;
  price: number;
  thumbnail_url: string | null;
};

import { backendFetchJson } from "./server";

export async function getProgramDetail(id: number | string): Promise<GetProgramDetailResponse> {
  const encodedId = encodeURIComponent(String(id));
  try {
    return await backendFetchJson<GetProgramDetailResponse>(`/programs/${encodedId}`, {
      method: "GET",
      cache: "no-store",
    });
  } catch (err) {
    console.error(`[API通信エラー] getProgramDetail:`, { id, err });
    throw err;
  }
}

export async function getPrograms(title?: string, tagIds?: Array<number | string>): Promise<GetProgramsResponse> {
  const params = new URLSearchParams();
  if (title) {
    params.set("title", title);
  }

  if (tagIds) {
    for (const id of tagIds) {
      const v = String(id);
      if (v) {
        params.append("tag_ids", v);
      }
    }
  }

  const queryString = params.toString();
  const path = queryString ? `/programs?${queryString}` : "/programs";
  try {
    return await backendFetchJson<GetProgramsResponse>(path, { method: "GET", cache: "no-store" });
  } catch (err) {
    console.error(`[API通信エラー] getPrograms:`, { title, tagIds, err });
    throw err;
  }
}

export async function getTopPrograms(): Promise<GetTopProgramsResponse> {
  try {
    return await backendFetchJson<GetTopProgramsResponse>("/top", { method: "GET", cache: "no-store" });
  } catch (err) {
    console.error(`[API通信エラー] getTopPrograms:`, { err });
    throw err;
  }
}

export async function getTopLikedPrograms(): Promise<GetTopLikedProgramsResponse> {
  try {
    return await backendFetchJson<GetTopLikedProgramsResponse>("/top/liked", { method: "GET", cache: "no-store" });
  } catch (err) {
    console.error(`[API通信エラー] getTopLikedPrograms:`, { err });
    throw err;
  }
}

export async function getTopViewedPrograms(): Promise<GetTopViewedProgramsResponse> {
  return backendFetchJson<GetTopViewedProgramsResponse>("/top/viewed", { method: "GET", cache: "no-store" });
}
