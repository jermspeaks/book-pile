import type { ReactNode } from "react";
import type { BookFilters, Named } from "../types";

export function Filters({
  value,
  piles,
  collections,
  genres,
  onChange,
}: {
  value: BookFilters;
  piles: Named[];
  collections: Named[];
  genres: Named[];
  onChange: (next: BookFilters) => void;
}) {
  function set(key: keyof BookFilters, v: string) {
    onChange({ ...value, [key]: v || undefined });
  }
  const pileOptions = piles ?? [];
  const collectionOptions = collections ?? [];
  const genreOptions = genres ?? [];
  return (
    <div className="flex flex-wrap items-end gap-2 px-6 py-3">
      <Field label="Search">
        <input
          className="rounded border border-border bg-surface px-2 py-1 text-sm"
          value={value.q ?? ""}
          onChange={(e) => set("q", e.target.value)}
          placeholder="Title or author"
        />
      </Field>
      <Field label="Ownership">
        <select
          className="rounded border border-border bg-surface px-2 py-1 text-sm"
          value={value.ownership ?? ""}
          onChange={(e) => set("ownership", e.target.value)}
        >
          <option value="">All</option>
          <option value="physical">Physical</option>
          <option value="digital">Digital</option>
          <option value="both">Both</option>
          <option value="coveting">Coveting</option>
        </select>
      </Field>
      <Field label="Read">
        <select
          className="rounded border border-border bg-surface px-2 py-1 text-sm"
          value={value.readStatus ?? ""}
          onChange={(e) => set("readStatus", e.target.value)}
        >
          <option value="">All</option>
          <option value="unread">Unread</option>
          <option value="reading">Reading</option>
          <option value="read">Read</option>
        </select>
      </Field>
      <Field label="Want ≥">
        <select
          className="rounded border border-border bg-surface px-2 py-1 text-sm"
          value={value.minWant ?? ""}
          onChange={(e) => set("minWant", e.target.value)}
        >
          <option value="">Any</option>
          <option value="3">3</option>
          <option value="4">4</option>
          <option value="5">5</option>
        </select>
      </Field>
      <Field label="Pile">
        <select
          className="rounded border border-border bg-surface px-2 py-1 text-sm"
          value={value.pileId ?? ""}
          onChange={(e) => set("pileId", e.target.value)}
        >
          <option value="">All</option>
          {pileOptions.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}
        </select>
      </Field>
      <Field label="Collection">
        <select
          className="rounded border border-border bg-surface px-2 py-1 text-sm"
          value={value.collectionId ?? ""}
          onChange={(e) => set("collectionId", e.target.value)}
        >
          <option value="">All</option>
          {collectionOptions.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>
      </Field>
      <Field label="Genre">
        <select
          className="rounded border border-border bg-surface px-2 py-1 text-sm"
          value={value.genreId ?? ""}
          onChange={(e) => set("genreId", e.target.value)}
        >
          <option value="">All</option>
          {genreOptions.map((g) => (
            <option key={g.id} value={g.id}>
              {g.name}
            </option>
          ))}
        </select>
      </Field>
      <Field label="Sort">
        <select
          className="rounded border border-border bg-surface px-2 py-1 text-sm"
          value={value.sort ?? "want"}
          onChange={(e) => set("sort", e.target.value)}
        >
          <option value="want">Want</option>
          <option value="title">Title</option>
          <option value="author">Author</option>
          <option value="added">Added</option>
        </select>
      </Field>
    </div>
  );
}

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="flex flex-col gap-1 text-[10px] uppercase tracking-wide text-muted">
      {label}
      {children}
    </label>
  );
}
