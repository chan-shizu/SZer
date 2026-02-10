import CommentForm from "./CommentForm";
import CommentsList from "./CommentsList";

export default async function Comments({ programId }: { programId: number }) {
  return (
    <div className="relative h-full">
      <h3 className="font-bold mb-2">コメント</h3>
      <div className="flex-1 overflow-y-auto pb-16">
        <CommentsList programId={programId} />
      </div>
      <div className="fixed bottom-0 left-0 right-0 z-10 bg-background border-t border-border shadow-[0_-4px_16px_-4px_rgba(0,0,0,0.10)]">
        <CommentForm programId={programId} />
      </div>
    </div>
  );
}
