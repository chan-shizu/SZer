import Link from "next/link";
import Image from "next/image";
import { ThumbsUp } from "lucide-react";

import { TopProgramItem } from "@/lib/api/programs";

type Props = {
  program: TopProgramItem;
};

export function TopProgramCard({ program }: Props) {
  return (
    <div className="snap-start">
      <Link href={`/programs/${program.program_id}`} className="block">
        {program.thumbnail_url ? (
          <div className="relative w-40 h-20">
            <Image
              src={program.thumbnail_url}
              alt={program.title}
              fill
              sizes="160px"
              className="rounded object-cover"
            />
            {/* 限定公開ラベル */}
            {program.is_limited_release && (
              <span className="absolute top-1 left-1 bg-brand text-white text-[10px] font-bold px-2 py-0.5 rounded z-10 shadow">
                限定公開
              </span>
            )}
          </div>
        ) : (
          <div className="w-40 h-24 rounded bg-subtle flex items-center justify-center text-foreground">No Image</div>
        )}
        <div className="mt-2 text-sm font-semibold text-foreground line-clamp-2">{program.title}</div>
        {/* 料金表示 */}
        {program.is_limited_release && program.price > 0 && (
          <div className="mt-1 text-xs font-bold text-brand">¥{program.price.toLocaleString()}（税込）</div>
        )}
        <div className="mt-1 flex items-center gap-x-3 text-xs text-muted-foreground">
          <div>視聴回数: {program.view_count}回</div>
          <div className="flex items-center gap-x-1">
            <ThumbsUp className="h-4 w-4 text-foreground" strokeWidth={2} />
            <span>{program.like_count}</span>
          </div>
        </div>
      </Link>
    </div>
  );
}
