import { getProgramDetail } from "@/lib/api/programs";
import { Video } from "./components/Video";
import { Title } from "./components/Title";
import { Description } from "./components/Description";
import { CategoryTags } from "./components/Tags";
import { Performers } from "./components/Performers";

type Props = { params: Promise<{ program_id: string }> };

export default async function Page({ params }: Props) {
  const { program_id } = await params;
  if (!program_id) {
    throw new Error("program_id is required");
  }
  const programDetail = await getProgramDetail(program_id);

  return (
    <div>
      <Video videoUrl={programDetail.program.video_url} />
      <div className="p-4 grid gap-y-4">
        <Title title={programDetail.program.title} />
        <Description description={programDetail.program.description ?? "説明文はありません"} />
        <CategoryTags categoryTags={programDetail.program.category_tags} />
        <Performers performers={programDetail.program.performers} />
      </div>
    </div>
  );
}
