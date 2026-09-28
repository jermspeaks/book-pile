import { useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import * as api from "../api";
import { BookPanel } from "../components/BookPanel";
import { CoverGrid } from "../components/CoverGrid";
import { Filters } from "../components/Filters";
import { ReadNext } from "../components/ReadNext";
import type { Book, BookFilters, Named, SuggestedBook } from "../types";

function fromParams(params: URLSearchParams): BookFilters {
  const keys = [
    "ownership",
    "pileId",
    "collectionId",
    "genreId",
    "authorId",
    "readStatus",
    "minWant",
    "sort",
    "q",
  ] as const;
  const out: BookFilters = {};
  for (const k of keys) {
    const v = params.get(k);
    if (v) out[k] = v;
  }
  if (!out.sort) out.sort = "want";
  return out;
}

export function Library({
  locked,
  heading,
  showNext = true,
}: {
  locked?: Partial<BookFilters>;
  heading?: string;
  showNext?: boolean;
}) {
  const [params, setParams] = useSearchParams();
  const filters = useMemo(() => ({ ...fromParams(params), ...locked }), [params, locked]);
  const [books, setBooks] = useState<Book[]>([]);
  const [next, setNext] = useState<SuggestedBook[]>([]);
  const [piles, setPiles] = useState<Named[]>([]);
  const [collections, setCollections] = useState<Named[]>([]);
  const [genres, setGenres] = useState<Named[]>([]);
  const [selectedId, setSelectedId] = useState<number | undefined>();
  const [error, setError] = useState<string | null>(null);

  const filterKey = [
    filters.ownership,
    filters.pileId,
    filters.collectionId,
    filters.genreId,
    filters.authorId,
    filters.readStatus,
    filters.minWant,
    filters.sort,
    filters.q,
    String(showNext),
  ].join("|");

  useEffect(() => {
    let cancelled = false;
    Promise.all([
      api.listBooks(filters),
      showNext ? api.suggestNext(12) : Promise.resolve([]),
      api.listPiles(),
      api.listCollections(),
      api.listGenres(),
    ])
      .then(([b, n, p, c, g]) => {
        if (cancelled) return;
        setBooks(b ?? []);
        setNext(n ?? []);
        setPiles(p ?? []);
        setCollections(c ?? []);
        setGenres(g ?? []);
      })
      .catch((e: Error) => {
        if (!cancelled) setError(e.message);
      });
    return () => {
      cancelled = true;
    };
  }, [filterKey, filters, showNext]);

  function updateFilters(nextFilters: BookFilters) {
    const merged = { ...nextFilters, ...locked };
    const sp = new URLSearchParams();
    for (const [k, v] of Object.entries(merged)) {
      if (v && !(locked && k in locked)) sp.set(k, v);
    }
    setParams(sp);
  }

  function mergeBook(updated: Book) {
    setBooks((cur) => cur.map((b) => (b.id === updated.id ? { ...b, ...updated } : b)));
    setNext((cur) => cur.map((b) => (b.id === updated.id ? { ...b, ...updated } : b)));
  }

  return (
    <div className="flex h-full">
      <div className="flex min-w-0 flex-1 flex-col">
        {heading && (
          <h1 className="px-6 pt-4 font-serif text-2xl text-heading">{heading}</h1>
        )}
        {error && <p className="px-6 pt-3 text-sm text-accent">{error}</p>}
        {showNext && (
          <ReadNext books={next} onSelect={(id) => setSelectedId(id)} />
        )}
        <Filters
          value={filters}
          piles={piles}
          collections={collections}
          genres={genres}
          onChange={updateFilters}
        />
        <div className="min-h-0 flex-1">
          <CoverGrid
            books={books}
            selectedId={selectedId}
            onSelect={(b) => setSelectedId(b.id)}
          />
        </div>
      </div>
      {selectedId !== undefined && (
        <BookPanel
          bookId={selectedId}
          onClose={() => setSelectedId(undefined)}
          onChanged={mergeBook}
        />
      )}
    </div>
  );
}
