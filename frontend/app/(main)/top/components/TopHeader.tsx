import Logo from "./Logo";
import UserStatus from "./UserStatus";

export default function TopHeader() {
  return (
    <header className="absolute left-0 right-0 top-0 flex items-center justify-between px-4 py-3 z-10 pointer-events-none select-none">
      <div className="pointer-events-auto">
        <Logo />
      </div>
      <div className="pointer-events-auto">
        <UserStatus />
      </div>
    </header>
  );
}
