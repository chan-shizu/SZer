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

export type ProgramDetails = {
  program_id: number;
  title: string;
  video_url: string;
  thumbnail_url: string | null;
  description: string | null;
  program_created_at: string;
  program_updated_at: string;
  category_tags: ProgramDetailsCategoryTag[];
  performers: ProgramDetailsPerformer[];
};

export type GetProgramDetailResponse = {
  program: ProgramDetails;
};

export type ProgramListItem = {
  program_id: number;
  title: string;
  thumbnail_url: string | null;
  category_tags: ProgramDetailsCategoryTag[];
};

export type GetProgramsResponse = {
  programs: ProgramListItem[];
};

import { backendFetchJson } from "./server";

export async function getProgramDetail(id: number | string): Promise<GetProgramDetailResponse> {
  const encodedId = encodeURIComponent(String(id));
  return backendFetchJson<GetProgramDetailResponse>(`/programs/${encodedId}`, { method: "GET" });
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
  return backendFetchJson<GetProgramsResponse>(path, { method: "GET" });
}
