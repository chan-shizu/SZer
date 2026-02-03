import { BottomTabBar } from "@/components/BottomTabBar";
import { ProgramList } from "./components/ProgramList";
import { SearchBar } from "./components/SearchBar";
import { Tags } from "./components/Tags";

export default async function Page(props: { searchParams: Promise<{ title?: string; tag_ids?: string | string[] }> }) {
  const searchParams = await props.searchParams;

  const tagIdsRaw = searchParams.tag_ids;
  const tagIds = Array.isArray(tagIdsRaw) ? tagIdsRaw : tagIdsRaw ? [tagIdsRaw] : [];
  const tagIdsNumber = tagIds
    .map((v) => Number(v))
    .filter((v) => Number.isFinite(v))
    .map((v) => v as number);

  return (
    <div>
      <SearchBar />
      <div className="mt-4">
        <Tags />
      </div>
      <ProgramList title={searchParams.title} tagIds={tagIdsNumber} />
      <BottomTabBar />
    </div>
  );
}
