import { SimpleModal } from "@/components/SimpleModal";

export type AuthModalProps = {
  open: boolean;
  onClose: () => void;
};

export function AuthModal({ open, onClose }: AuthModalProps) {
  return (
    <SimpleModal open={open} onClose={onClose}>
      <div className="text-center space-y-4">
        <div className="text-lg font-bold">ログインまたは会員登録が必要です</div>
        <div className="flex gap-4 justify-center">
          <a
            href="/login"
            className="px-4 py-2 rounded bg-foreground text-white font-semibold shadow hover:bg-zinc-800 transition-colors"
          >
            ログイン
          </a>
          <a
            href="/signup"
            className="px-4 py-2 rounded bg-brand text-white font-semibold shadow hover:bg-orange-700 transition-colors"
          >
            新規登録
          </a>
        </div>
      </div>
    </SimpleModal>
  );
}
