
import { BottomTabBar } from "@/components/BottomTabBar";

export default function MainLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <>
      <div className="pb-16">
        {children}
      </div>
      <BottomTabBar />
    </>
  );
}
