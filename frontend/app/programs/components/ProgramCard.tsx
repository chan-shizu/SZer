import { ProgramListItem } from "@/lib/api/programs";
import Link from "next/link";
import Image from "next/image";

type Props = { program: ProgramListItem };

const getTagColor = (name: string) => {
  if (name === "音楽") return "bg-yellow-100 text-yellow-800 border-yellow-200";
  if (name === "グルメ") return "bg-red-100 text-red-800 border-red-200";
  return "bg-gray-100 text-gray-800 border-gray-200";
};

export const ProgramCard = ({ program }: Props) => {
  return (
    <Link
      href={`/programs/${program.program_id}`}
      className="flex justify-between items-center p-2 hover:bg-gray-50 rounded transition-colors"
    >
      <div className="flex-1 pr-4">
        <h3 className="text-sm font-semibold text-foreground mb-1">{program.title}</h3>
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
        </div>
      ) : (
        <div className="">
          <span className="text-foreground">No Image</span>
        </div>
      )}
    </Link>
  );
};
