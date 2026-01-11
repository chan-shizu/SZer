type Tag = { id: number; name: string };

type Props = { categoryTags: Tag[] };

export const CategoryTags = ({ categoryTags }: Props) => {
  return (
    <div className="flex flex-wrap gap-2">
      {categoryTags.map((tag) => (
        <span
          key={tag.id}
          className="inline-block bg-gray-200 rounded-full px-3 py-1 text-sm font-semibold text-gray-700"
        >
          #{tag.name}
        </span>
      ))}
    </div>
  );
};
