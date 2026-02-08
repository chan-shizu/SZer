import { BottomTabBar } from "@/components/BottomTabBar";
import RequestForm from "./RequestForm";

export default function RequestPage() {
  return (
    <div className="min-h-screen pb-24">
      <div className="max-w-lg mx-auto px-4 py-6">
        <h1 className="text-xl font-bold mb-6">リクエスト</h1>
        <p className="text-sm text-gray-500 mb-6">
          見たい番組やご要望をお聞かせください！
          <br />
          また動画投稿してくれる人も大募集ですので連絡ください！
        </p>
        <RequestForm />
      </div>
      <BottomTabBar />
    </div>
  );
}
