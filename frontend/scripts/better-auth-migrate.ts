import { getMigrations } from "better-auth/db";

import { auth } from "@/lib/auth/auth";

async function main() {
  const { compileMigrations } = await getMigrations(auth.options);
  const sql = await compileMigrations();
  console.log(sql);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
