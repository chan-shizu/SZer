type Props = { videoUrl: string };

export const Video = ({ videoUrl }: Props) => {
  return (
    <video controls style={{ width: "100%", maxWidth: 960 }} src={videoUrl}>
      お使いのブラウザは video タグに対応していません。
    </video>
  );
};
