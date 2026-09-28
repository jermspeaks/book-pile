import { useEffect, useState } from "react";
import type { Book, Named } from "../types";
import * as api from "../api";
import { StarRating } from "./StarRating";

export function BookPanel({
  bookId,
  onClose,
  onChanged,
}: {
  bookId: number;
  onClose: () => void;
  onChanged: (book: Book) => void;
}) {
  const [book, setBook] = useState<Book | null>(null);
  const [piles, setPiles] = useState<Named[]>([]);
  const [collections, setCollections] = useState<Named[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [newPile, setNewPile] = useState("");
  const [newCollection, setNewCollection] = useState("");

  useEffect(() => {
    let cancelled = false;
    Promise.all([api.getBook(bookId), api.listPiles(), api.listCollections()])
      .then(([b, p, c]) => {
        if (cancelled) return;
        setBook(b);
        setPiles(p);
        setCollections(c);
      })
      .catch((e: Error) => {
        if (!cancelled) setError(e.message);
      });
    return () => {
      cancelled = true;
    };
  }, [bookId]);

  async function patch(body: Record<string, unknown>) {
    const updated = await api.patchBook(bookId, body);
    setBook(updated);
    onChanged(updated);
  }

  async function savePiles(ids: number[]) {
    const updated = await api.putBookPiles(bookId, ids);
    setBook(updated);
    onChanged(updated);
  }

  async function saveCollections(ids: number[]) {
    const updated = await api.putBookCollections(bookId, ids);
    setBook(updated);
    onChanged(updated);
  }

  if (error) {
    return (
      <aside className="flex h-full w-[min(100%,22rem)] flex-col border-l border-border bg-bg p-4">
        <p className="text-sm text-accent">{error}</p>
        <button type="button" className="mt-4 text-sm" onClick={onClose}>
          Close
        </button>
      </aside>
    );
  }
  if (!book) {
    return (
      <aside className="h-full w-[min(100%,22rem)] border-l border-border bg-bg p-4 text-muted">
        Loading…
      </aside>
    );
  }

  const pileIds = new Set((book.piles ?? []).map((p) => p.id));
  const collectionIds = new Set((book.collections ?? []).map((c) => c.id));

  return (
    <aside className="flex h-full w-[min(100%,22rem)] shrink-0 flex-col overflow-auto border-l border-border bg-bg p-4">
      <div className="mb-3 flex items-start justify-between gap-2">
        <h2 className="font-serif text-xl leading-tight text-heading">{book.title}</h2>
        <button type="button" className="text-muted" onClick={onClose} aria-label="Close">
          ×
        </button>
      </div>
      <p className="mb-1 text-sm text-muted">
        {book.authors.map((a) => a.name).join(", ") || "Unknown author"}
      </p>
      {book.seriesName && (
        <p className="mb-3 text-xs text-dim">
          {book.seriesName}
          {book.seriesIndex ? ` ${book.seriesIndex}` : ""}
        </p>
      )}

      <label className="mb-1 text-xs uppercase tracking-wide text-muted">Want to read</label>
      <StarRating value={book.wantRating} onChange={(n) => void patch({ wantRating: n })} />

      <label className="mt-4 mb-1 text-xs uppercase tracking-wide text-muted">Read status</label>
      <select
        className="rounded border border-border bg-surface px-2 py-1 text-sm"
        value={book.readStatus}
        onChange={(e) => void patch({ readStatus: e.target.value })}
      >
        <option value="unread">Unread</option>
        <option value="reading">Currently reading</option>
        <option value="read">Read</option>
      </select>

      <label className="mt-4 mb-1 text-xs uppercase tracking-wide text-muted">Ownership</label>
      <div className="flex gap-2">
        <Toggle
          label="Physical"
          on={book.ownedPhysical}
          onToggle={() => void patch({ ownedPhysical: !book.ownedPhysical })}
        />
        <Toggle
          label="Digital"
          on={book.ownedDigital}
          onToggle={() => void patch({ ownedDigital: !book.ownedDigital })}
        />
      </div>
      <p className="mt-1 text-xs text-dim">
        {book.ownedPhysical || book.ownedDigital ? "Owned" : "Coveting — not owned on any platform"}
      </p>

      <label className="mt-4 mb-1 text-xs uppercase tracking-wide text-muted">Piles</label>
      <div className="flex flex-col gap-1">
        {piles.map((p) => (
          <label key={p.id} className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={pileIds.has(p.id)}
              onChange={() => {
                const next = new Set(pileIds);
                if (next.has(p.id)) next.delete(p.id);
                else next.add(p.id);
                void savePiles([...next]);
              }}
            />
            {p.name}
          </label>
        ))}
      </div>
      <form
        className="mt-2 flex gap-1"
        onSubmit={(e) => {
          e.preventDefault();
          const name = newPile.trim();
          if (!name) return;
          void api.createPile(name).then((p) => {
            setPiles((cur) => [...cur, p].sort((a, b) => a.name.localeCompare(b.name)));
            setNewPile("");
            void savePiles([...pileIds, p.id]);
          });
        }}
      >
        <input
          className="min-w-0 flex-1 rounded border border-border bg-surface px-2 py-1 text-sm"
          placeholder="New pile"
          value={newPile}
          onChange={(e) => setNewPile(e.target.value)}
        />
        <button type="submit" className="rounded border border-border px-2 text-sm">
          Add
        </button>
      </form>

      <label className="mt-4 mb-1 text-xs uppercase tracking-wide text-muted">Collections</label>
      <div className="flex flex-col gap-1">
        {collections.map((c) => (
          <label key={c.id} className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={collectionIds.has(c.id)}
              onChange={() => {
                const next = new Set(collectionIds);
                if (next.has(c.id)) next.delete(c.id);
                else next.add(c.id);
                void saveCollections([...next]);
              }}
            />
            {c.name}
          </label>
        ))}
      </div>
      <form
        className="mt-2 flex gap-1"
        onSubmit={(e) => {
          e.preventDefault();
          const name = newCollection.trim();
          if (!name) return;
          void api.createCollection(name).then((c) => {
            setCollections((cur) => [...cur, c].sort((a, b) => a.name.localeCompare(b.name)));
            setNewCollection("");
            void saveCollections([...collectionIds, c.id]);
          });
        }}
      >
        <input
          className="min-w-0 flex-1 rounded border border-border bg-surface px-2 py-1 text-sm"
          placeholder="New collection"
          value={newCollection}
          onChange={(e) => setNewCollection(e.target.value)}
        />
        <button type="submit" className="rounded border border-border px-2 text-sm">
          Add
        </button>
      </form>

      {book.genres && book.genres.length > 0 && (
        <p className="mt-4 text-xs text-muted">
          {book.genres.map((g) => g.name).join(" · ")}
        </p>
      )}
      {book.description && (
        <p className="mt-4 text-sm leading-relaxed text-text">{book.description}</p>
      )}
    </aside>
  );
}

function Toggle({
  label,
  on,
  onToggle,
}: {
  label: string;
  on: boolean;
  onToggle: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onToggle}
      className={`rounded-full border px-3 py-1 text-sm ${
        on ? "border-accent bg-accent-dim text-heading" : "border-border bg-surface text-muted"
      }`}
    >
      {label}
    </button>
  );
}
