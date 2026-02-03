import Link from "next/link";
import { redirect } from "next/navigation";

import { getLikedPrograms, getWatchingPrograms } from "@/lib/api/mypage";
import { ApiError } from "@/lib/api/error";
import { ProgramCard } from "@/app/programs/components/ProgramCard";
import { BottomTabBar } from "@/components/BottomTabBar";

type Tab = "watching" | "liked";

function normalizeTab(v?: string): Tab {
  return v === "liked" ? "liked" : "watching";
}

function TabLink({ href, label, selected }: { href: string; label: string; selected: boolean }) {
  return (
    <Link
      href={href}
      className={
        "px-3 py-2 text-sm font-medium border-b-2 " +
        (selected ? "border-foreground text-foreground" : "border-transparent text-muted-foreground")
      }
    >
      {label}
    </Link>
  );
}

export default async function Page(props: { searchParams: Promise<{ tab?: string }> }) {
  const searchParams = await props.searchParams;
  const tab = normalizeTab(searchParams.tab);

  const response = tab === "liked" ? await getLikedPrograms() : await getWatchingPrograms();
  const programs = response.programs;

  return (
    <div>
      <div className="p-3">
        <h1 className="text-lg font-semibold text-foreground">マイページ</h1>
      </div>

      <div className="px-3">
        <div className="flex gap-2 border-b border-muted">
          <TabLink href="/mypage?tab=watching" label="視聴中" selected={tab === "watching"} />
          <TabLink href="/mypage?tab=liked" label="いいね" selected={tab === "liked"} />
        </div>
      </div>

      <div className="space-y-4 p-3">
        {programs.length === 0 ? (
          <div className="text-sm text-muted-foreground">まだありません</div>
        ) : (
          programs.map((program) => <ProgramCard key={program.program_id} program={program} />)
        )}
      </div>
      <BottomTabBar />
    </div>
  );
}
