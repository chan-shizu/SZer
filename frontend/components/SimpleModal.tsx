import { ReactNode } from "react";

export type SimpleModalProps = {
  open: boolean;
  onClose: () => void;
  children: ReactNode;
};

export function SimpleModal({ open, onClose, children }: SimpleModalProps) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="bg-white rounded-lg shadow-lg p-6 min-w-[300px] max-w-[90vw] relative">
        <button
          className="absolute top-2 right-2 text-gray-400 hover:text-gray-700"
          onClick={onClose}
          aria-label="閉じる"
        >
          ×
        </button>
        {children}
      </div>
    </div>
  );
}
