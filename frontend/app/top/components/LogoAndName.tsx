import Logo from "./Logo";
import UserStatus from "./UserStatus";

export default function LogoAndName() {
  return (
    <div className="flex items-center justify-between py-3 pointer-events-none select-none">
      <div className="pointer-events-auto">
        <Logo />
      </div>
      <div className="pointer-events-auto">
        <UserStatus />
      </div>
    </div>
  );
}
