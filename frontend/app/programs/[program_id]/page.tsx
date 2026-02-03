import { getProgramDetail } from "@/lib/api/programs";
import { Video } from "./components/Video";
import { Title } from "./components/Title";
import { Description } from "./components/Description";
import { CategoryTags } from "./components/Tags";
import { Performers } from "./components/Performers";
import { LikeButton } from "./components/LikeButton";
import Comments from "./components/Comments";

type Props = { params: Promise<{ program_id: string }> };

export default async function Page({ params }: Props) {
  const { program_id } = await params;
  if (!program_id) {
    throw new Error("program_id is required");
  }
  const programDetail = await getProgramDetail(program_id);
  const programIdNumber = Number(programDetail.program.program_id);

  return (
    <div className="flex flex-col h-screen">
      {/* 動画は常に上部に固定表示 */}
      <div className="shrink-0">
        <Video
          programId={programIdNumber}
          videoUrl={programDetail.program.video_url}
          startPositionSeconds={programDetail.program.watch_history?.position_seconds ?? undefined}
        />
      </div>
      {/* 下部だけスクロール可能 */}
      <div className="flex-1 min-h-0 overflow-y-auto p-4 grid gap-y-4 bg-white">
        <Title title={programDetail.program.title} />
        <div className="flex items-center justify-between gap-x-4">
          <div className="text-sm text-gray-600">視聴回数: {programDetail.program.view_count}回</div>
          <LikeButton
            programId={programIdNumber}
            initialLiked={programDetail.program.liked}
            initialLikeCount={programDetail.program.like_count}
          />
        </div>
        <Description description={programDetail.program.description ?? "説明文はありません"} />
        <CategoryTags categoryTags={programDetail.program.category_tags} />
        <Performers performers={programDetail.program.performers} />
        <Comments programId={programIdNumber} />
      </div>
    </div>
  );
}
