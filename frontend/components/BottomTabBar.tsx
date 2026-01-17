import Link from "next/link";
import { Home, Search } from "lucide-react";

const tabs = [
  {
    href: "/top",
    label: "トップ",
    icon: Home,
  },
  {
    href: "/programs",
    label: "探す",
    icon: Search,
  },
];

export const BottomTabBar = () => {
  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 border-t border-muted bg-background">
      <div className="flex items-center justify-around">
        {tabs.map((tab) => (
          <BottomTabBarItem key={tab.href} href={tab.href} label={tab.label} icon={tab.icon} />
        ))}
      </div>
    </nav>
  );
};

const BottomTabBarItem = ({
  href,
  label,
  icon: Icon,
}: {
  href: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
}) => {
  return (
    <Link
      href={href}
      className="flex flex-col w-full items-center px-3 py-4 rounded-full text-sm border-muted bg-background text-foreground "
    >
      <Icon className="h-4 w-4" aria-hidden="true" />
      {label}
    </Link>
  );
};
