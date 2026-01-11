type Props = { description: string };

export const Description = ({ description }: Props) => {
  return <p className="whitespace-pre-wrap">{description}</p>;
};
