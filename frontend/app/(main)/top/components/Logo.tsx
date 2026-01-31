export default function Logo() {
  return (
    <div className="flex items-end select-none" style={{ userSelect: "none" }}>
      <span className="flex items-end gap-1 align-bottom">
        {/* SZ: オレンジ系グラデーション＋斜め＋やや大きめ */}
        <span
          className="text-4xl font-extrabold "
          style={{
            background: "linear-gradient(180deg, #FFB300 0%, #FF6F00 100%)",
            WebkitBackgroundClip: "text",
            WebkitTextFillColor: "transparent",
            backgroundClip: "text",
            fontFamily: "Montserrat, Arial, sans-serif",
            transform: "skew(-10deg)",
            textShadow: "0 2px 8px #FFB30033",
            alignSelf: "flex-end",
          }}
        >
          SZ
        </span>
        {/* er: グレー＋細字＋やや小さめ */}
        <span
          className="text-xl"
          style={{
            color: "#222",
            fontFamily: "Montserrat, Arial, sans-serif",
            letterSpacing: "-0.04em",
            fontWeight: 500,
            alignSelf: "flex-end",
          }}
        >
          er
        </span>
      </span>
      <div className="flex flex-col justify-end ml-3 font-semibold tracking-widest" style={{ letterSpacing: "0.15em" }}>
        <div className="flex items-end">
          <span className="text-2xl leading-none">28</span>
          <span className="text-xs align-super ml-0.5">th</span>
        </div>
        <span className="text-xs leading-none">Anniversary</span>
      </div>
    </div>
  );
}
