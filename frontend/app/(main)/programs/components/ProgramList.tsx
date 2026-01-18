import { getPrograms } from "@/lib/api/programs";
import { ProgramCard } from "./ProgramCard";

export const ProgramList = async ({ title, tagIds }: { title?: string; tagIds?: number[] }) => {
  const response = await getPrograms(title, tagIds);
  const programs = response.programs;

  return (
    <div className="space-y-4 p-3">
      {programs && programs.map((program) => <ProgramCard key={program.program_id} program={program} />)}
    </div>
  );
};
