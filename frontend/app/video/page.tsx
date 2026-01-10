"use client";

import React from "react";

export default function Page() {
  const videoUrl = `${process.env.NEXT_PUBLIC_S3_ENDPOINT ?? ""}/video/ReInventAI.mp4`;

  return (
    <main style={{ display: "flex", justifyContent: "center", padding: "2rem" }}>
      <video controls style={{ width: "100%", maxWidth: 960 }} src={videoUrl}>
        お使いのブラウザは video タグに対応していません。
      </video>
    </main>
  );
}
