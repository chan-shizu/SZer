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

import { backendFetchJson } from "./server";

export async function getProgramDetail(id: number | string): Promise<GetProgramDetailResponse> {
  const encodedId = encodeURIComponent(String(id));
  return backendFetchJson<GetProgramDetailResponse>(`/programs/${encodedId}`, { method: "GET" });
}

export default getProgramDetail;
