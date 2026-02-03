"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useTransition } from "react";

export type CategoryTagOption = { id: number; name: string };

const categoryTags: CategoryTagOption[] = [
  { id: 1, name: "音楽" },
  { id: 2, name: "お笑い" },
  { id: 3, name: "グルメ" },
  { id: 4, name: "その他" },
];

export const Tags = () => {
  const searchParams = useSearchParams();
  const pathname = usePathname();
  const { replace } = useRouter();
  const [isPending, startTransition] = useTransition();

  const selectedTagIds = searchParams
    .getAll("tag_ids")
    .map((v) => Number(v))
    .filter((v) => Number.isFinite(v))
    .map((v) => v as number);

  const onToggleTag = (tagId: number) => {
    const nextTagIds = selectedTagIds.includes(tagId)
      ? selectedTagIds.filter((id) => id !== tagId)
      : [...selectedTagIds, tagId];

    const params = new URLSearchParams(searchParams);
    params.delete("tag_ids");
    for (const id of nextTagIds) {
      params.append("tag_ids", String(id));
    }

    startTransition(() => {
      replace(`${pathname}?${params.toString()}`);
    });
  };

  if (categoryTags.length === 0) return null;

  return (
    <div className={`px-4 flex flex-wrap gap-x-2 ${isPending ? "opacity-60" : ""}`.trim()}>
      {categoryTags.map((t) => {
        const checked = selectedTagIds.includes(t.id);
        return (
          <button
            key={t.id}
            type="button"
            aria-pressed={checked}
            onClick={() => onToggleTag(t.id)}
            className={
              "px-3 py-1 rounded-full text-sm select-none border-2 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-foreground focus-visible:ring-offset-2 focus-visible:ring-offset-background " +
              (checked
                ? "bg-foreground text-background border-foreground hover:opacity-90"
                : "bg-background text-foreground border-muted hover:bg-muted")
            }
          >
            {t.name}
          </button>
        );
      })}
    </div>
  );
};
