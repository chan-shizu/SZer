import { ProgramListItem } from "@/lib/api/programs";
import Link from "next/link";
import Image from "next/image";
import { ThumbsUp } from "lucide-react";

type Props = { program: ProgramListItem };

const getTagColor = (name: string) => {
  if (name === "音楽") return "bg-yellow-100 text-yellow-800 border-yellow-200";
  if (name === "グルメ") return "bg-red-100 text-red-800 border-red-200";
  return "bg-subtle text-foreground border-border";
};

export const ProgramCard = ({ program }: Props) => {
  return (
    <Link
      href={`/programs/${program.program_id}`}
      className="flex justify-between items-center p-2 hover:bg-subtle rounded transition-colors"
    >
      <div className="flex-1 pr-4">
        <h3 className="text-sm font-semibold text-foreground mb-1">{program.title}</h3>
        {/* 料金表示 */}
        {program.is_limited_release && program.price > 0 && (
          <div className="text-xs font-bold text-brand mb-1">¥{program.price.toLocaleString()}（税込）</div>
        )}
        <div className="flex items-center gap-x-3 text-xs text-muted-foreground mb-1">
          <div>視聴回数: {program.view_count}回</div>
          <div className="flex items-center gap-x-1">
            <ThumbsUp className="h-4 w-4 text-foreground" strokeWidth={2} />
            <span>{program.like_count}</span>
          </div>
        </div>
        <div className="flex flex-wrap gap-1">
          {program.category_tags.map((tag) => {
            const colorClass = getTagColor(tag.name);
            return (
              <span key={tag.id} className={`text-xs px-1 py-0.5 rounded border ${colorClass}`}>
                {tag.name}
              </span>
            );
          })}
        </div>
      </div>
      {program.thumbnail_url ? (
        <div className="relative w-32 h-16">
          <Image src={program.thumbnail_url} alt={program.title} fill sizes="128px" className="rounded object-cover" />
          {/* 限定公開ラベル */}
          {program.is_limited_release && (
            <span className="absolute top-1 left-1 bg-brand text-white text-[10px] font-bold px-2 py-0.5 rounded z-10 shadow">
              限定公開
            </span>
          )}
        </div>
      ) : (
        <div className="">
          <span className="text-foreground">No Image</span>
        </div>
      )}
    </Link>
  );
};
