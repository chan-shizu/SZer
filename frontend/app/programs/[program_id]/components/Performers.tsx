import Image from "next/image";

type Performer = {
  id: number;
  full_name: string;
  full_name_kana: string;
  image_url: string | null;
};

type Props = { performers: Performer[] };

export const Performers = ({ performers }: Props) => {
  return (
    <div>
      <h3 className="font-bold mb-2">出演者</h3>
      <div className="flex flex-wrap gap-2 mt-2">
        {performers.map((performer) => (
          <div
            key={performer.id}
            className="flex gap-x-2 px-3 py-1 rounded-md border-2 border-input items-center text-muted-foreground"
          >
            {performer.image_url && <Image src={performer.image_url} alt="" width={28} height={28} />}
            <p className="font-semibold text-sm">{performer.full_name}</p>
          </div>
        ))}
      </div>
    </div>
  );
};
