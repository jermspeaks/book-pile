import { useLayoutEffect, useRef, useState } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import type { Book } from "../types";
import { CoverArt } from "./CoverArt";
import { ownershipLabel } from "../lib/ownership";

const MIN_COL = 132;

export function CoverGrid({
  books,
  selectedId,
  onSelect,
}: {
  books: Book[];
  selectedId?: number;
  onSelect: (book: Book) => void;
}) {
  const parentRef = useRef<HTMLDivElement>(null);
  const [width, setWidth] = useState(800);

  useLayoutEffect(() => {
    const el = parentRef.current;
    if (!el) return;
    const ro = new ResizeObserver(() => setWidth(el.clientWidth));
    ro.observe(el);
    setWidth(el.clientWidth);
    return () => ro.disconnect();
  }, []);

  const colCount = Math.max(1, Math.floor((width - 16) / MIN_COL));
  const rowCount = Math.max(1, Math.ceil(books.length / colCount));
  const virtualizer = useVirtualizer({
    count: books.length === 0 ? 0 : rowCount,
    getScrollElement: () => parentRef.current,
    estimateSize: () => Math.round((MIN_COL - 8) * 1.5 + 56),
    overscan: 6,
  });

  return (
    <div ref={parentRef} className="h-full overflow-auto px-4 pb-8">
      {books.length === 0 ? (
        <p className="pt-12 text-center text-muted">No books in this view yet.</p>
      ) : (
        <div
          className="relative w-full"
          style={{ height: virtualizer.getTotalSize() }}
        >
          {virtualizer.getVirtualItems().map((row) => {
            const start = row.index * colCount;
            const slice = books.slice(start, start + colCount);
            return (
              <div
                key={row.key}
                className="absolute top-0 left-0 w-full"
                style={{
                  transform: `translateY(${row.start}px)`,
                  height: row.size,
                }}
              >
                <div
                  className="grid h-full gap-2 px-1"
                  style={{ gridTemplateColumns: `repeat(${colCount}, minmax(0, 1fr))` }}
                >
                  {slice.map((book) => (
                    <button
                      key={book.id}
                      type="button"
                      onClick={() => onSelect(book)}
                      className={`flex flex-col overflow-hidden rounded-md border bg-surface text-left transition hover:-translate-y-0.5 hover:bg-surface-hover ${
                        selectedId === book.id ? "border-accent" : "border-transparent"
                      }`}
                    >
                      <div className="aspect-[2/3] w-full overflow-hidden bg-surface-hover">
                        <CoverArt title={book.title} coverUrl={book.coverUrl} />
                      </div>
                      <div className="truncate px-1.5 pt-1.5 text-xs text-heading">
                        {book.title}
                      </div>
                      <div className="truncate px-1.5 pb-1.5 text-[10px] text-muted">
                        {book.authors?.map((a) => a.name).join(", ") || "Unknown"}
                        {" · "}
                        {ownershipLabel(book.ownership)}
                        {book.wantRating ? ` · ${"★".repeat(book.wantRating)}` : ""}
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
