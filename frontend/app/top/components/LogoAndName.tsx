import Link from "next/link";
import { UserCircle } from "lucide-react";

import Logo from "./Logo";
import UserStatus from "./UserStatus";

export default function LogoAndName() {
  return (
    <div className="flex items-center justify-between py-3 pointer-events-none select-none">
      <div className="pointer-events-auto">
        <Logo />
      </div>
      <div className="pointer-events-auto flex items-center gap-3">
        <UserStatus />
        <Link href="/mypage/profile" className="text-zinc-700 dark:text-zinc-200">
          <UserCircle className="h-5 w-5" />
        </Link>
      </div>
    </div>
  );
}
