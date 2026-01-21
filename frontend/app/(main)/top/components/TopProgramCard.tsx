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
          <div className="relative w-40 h-32">
            <Image
              src={program.thumbnail_url}
              alt={program.title}
              fill
              sizes="160px"
              className="rounded object-cover"
            />
          </div>
        ) : (
          <div className="w-48 h-24 rounded bg-gray-100 flex items-center justify-center text-foreground">No Image</div>
        )}
        <div className="mt-2 text-sm font-semibold text-foreground line-clamp-2">{program.title}</div>
        <div className="mt-1 flex items-center gap-x-3 text-xs text-gray-600">
          <div>視聴回数: {program.view_count}回</div>
          <div className="flex items-center gap-x-1">
            <ThumbsUp className="h-4 w-4 text-gray-900" strokeWidth={2} />
            <span>{program.like_count}</span>
          </div>
        </div>
      </Link>
    </div>
  );
}
