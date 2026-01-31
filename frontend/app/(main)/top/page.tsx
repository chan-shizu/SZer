import { getTopLikedPrograms, getTopPrograms, getTopViewedPrograms } from "@/lib/api/programs";
import { TopProgramCard } from "./components/TopProgramCard";
import LogoAndName from "./components/LogoAndName";

export const dynamic = "force-dynamic";

export default async function TopPage() {
  let programs: Awaited<ReturnType<typeof getTopPrograms>>["programs"];
  let topLikedPrograms: Awaited<ReturnType<typeof getTopLikedPrograms>>["programs"];
  let topViewedPrograms: Awaited<ReturnType<typeof getTopViewedPrograms>>["programs"];
  try {
    const [topRes, likedRes, viewedRes] = await Promise.all([
      getTopPrograms(),
      getTopLikedPrograms(),
      getTopViewedPrograms(),
    ]);
    programs = topRes.programs;
    topLikedPrograms = likedRes.programs;
    topViewedPrograms = viewedRes.programs;
  } catch (err) {
    throw err;
  }

  return (
    <div className="min-h-screen w-full max-w-3xl pb-12 px-4 bg-white dark:bg-black sm:px-6">
      <LogoAndName />
      <div className="space-y-10">
        <section>
          <h1 className="text-2xl font-extrabold text-zinc-900 dark:text-zinc-100">新着番組</h1>

          <div className="mt-4 -mx-4 px-4 overflow-x-auto no-scrollbar">
            <div className="flex gap-4 snap-x snap-mandatory">
              {programs.map((program) => (
                <TopProgramCard key={program.program_id} program={program} />
              ))}
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-extrabold text-zinc-900 dark:text-zinc-100">いいね数上位</h2>

          <div className="mt-4 -mx-4 px-4 overflow-x-auto no-scrollbar">
            <div className="flex gap-4 snap-x snap-mandatory">
              {topLikedPrograms.map((program) => (
                <TopProgramCard key={program.program_id} program={program} />
              ))}
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-extrabold text-zinc-900 dark:text-zinc-100">視聴回数上位</h2>

          <div className="mt-4 -mx-4 px-4 overflow-x-auto no-scrollbar">
            <div className="flex gap-4 snap-x snap-mandatory">
              {topViewedPrograms.map((program) => (
                <TopProgramCard key={program.program_id} program={program} />
              ))}
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}
