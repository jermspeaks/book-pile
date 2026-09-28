import { Link } from "react-router-dom";
import type { SuggestedBook } from "../types";
import { CoverArt } from "./CoverArt";

export function ReadNext({
  books,
  onSelect,
}: {
  books: SuggestedBook[];
  onSelect: (id: number) => void;
}) {
  if (books.length === 0) return null;
  return (
    <section className="shrink-0 px-6 pt-4">
      <div className="mb-2 flex items-baseline justify-between">
        <h2 className="text-xs font-medium uppercase tracking-widest text-muted">Read next</h2>
        <Link to="/authors" className="text-xs text-accent no-underline">
          Suggested authors
        </Link>
      </div>
      <div className="flex gap-3 overflow-x-auto pb-2">
        {books.map((book) => (
          <button
            key={book.id}
            type="button"
            onClick={() => onSelect(book.id)}
            className="w-28 shrink-0 text-left"
          >
            <div className="aspect-[2/3] overflow-hidden rounded-md bg-surface">
              <CoverArt title={book.title} coverUrl={book.coverUrl} />
            </div>
            <div className="mt-1 truncate text-xs text-heading">{book.title}</div>
            <div className="line-clamp-2 text-[10px] leading-snug text-muted">{book.reason}</div>
          </button>
        ))}
      </div>
    </section>
  );
}
