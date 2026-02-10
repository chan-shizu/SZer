"use client";

import { useSearchParams, usePathname, useRouter } from "next/navigation";
import { useTransition, useRef, useState, useEffect } from "react";

export const SearchBar = ({}) => {
  const searchParams = useSearchParams();
  const pathname = usePathname();
  const { replace } = useRouter();
  const [isPending, startTransition] = useTransition();
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const [showClear, setShowClear] = useState(false);

  useEffect(() => {
    const title = searchParams.get("title")?.toString() || "";
    if (inputRef.current && inputRef.current.value !== title) {
      inputRef.current.value = title;
    }
  }, [searchParams]);

  function updateParams(term: string) {
    const params = new URLSearchParams(searchParams);
    if (term) {
      params.set("title", term);
    } else {
      params.delete("title");
    }

    startTransition(() => {
      replace(`${pathname}?${params.toString()}`);
    });
  }

  function handleSearch(formData: FormData) {
    const term = formData.get("title") as string;
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
    updateParams(term);
  }

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const term = e.target.value;
    setShowClear(!!term);
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
    timeoutRef.current = setTimeout(() => {
      updateParams(term);
    }, 300);
  }

  function handleClear() {
    if (inputRef.current) {
      inputRef.current.value = "";
      inputRef.current.focus();
    }
    setShowClear(false);
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
    updateParams("");
  }

  return (
    <form action={handleSearch} className="w-full p-4 shadow-md bg-background rounded-lg space-y-3">
      <div className="relative flex items-center w-full">
        {/* Input container for better absolute positioning if needed */}
        <input
          ref={inputRef}
          name="title"
          type="text"
          onChange={handleChange}
          className="w-full py-2 pl-6 pr-20 bg-subtle rounded text-foreground placeholder-muted-foreground transition-all duration-300 text-sm focus:outline-none focus:ring-2 focus:ring-blue-100"
          placeholder="番組名で検索"
          defaultValue={searchParams.get("title")?.toString()}
        />
        {isPending && (
          <div className="absolute left-1/2 top-1/2 transform -translate-x-1/2 -translate-y-1/2 pointer-events-none">
            <div className="w-5 h-5 border-2 border-input border-t-muted-foreground rounded-full animate-spin"></div>
          </div>
        )}

        <div className="absolute right-6 top-1/2 transform -translate-y-1/2 flex items-center gap-2">
          {showClear && (
            <button
              type="button"
              onClick={handleClear}
              className="text-muted-foreground hover:text-foreground focus:outline-none p-1 rounded-full hover:bg-subtle transition-colors"
              aria-label="検索キーワードをクリア"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                strokeWidth={2}
                stroke="currentColor"
                className="w-4 h-4"
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          )}
        </div>
      </div>
    </form>
  );
};
