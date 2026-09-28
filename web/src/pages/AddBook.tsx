import { useState } from "react";
import { useNavigate } from "react-router-dom";
import * as api from "../api";

export function AddBook() {
  const navigate = useNavigate();
  const [title, setTitle] = useState("");
  const [authors, setAuthors] = useState("");
  const [isbn, setIsbn] = useState("");
  const [ownership, setOwnership] = useState<"coveting" | "physical" | "digital">("coveting");
  const [fetchMetadata, setFetchMetadata] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  return (
    <div className="mx-auto max-w-xl px-6 py-10">
      <h1 className="font-serif text-3xl text-heading">Add a book</h1>
      <p className="mt-3 text-sm text-muted">
        For physical copies and titles you covet that are not in Calibre. ISBN lookup fills cover
        and metadata from Open Library when possible.
      </p>
      <form
        className="mt-6 flex flex-col gap-4"
        onSubmit={(e) => {
          e.preventDefault();
          setBusy(true);
          setError(null);
          api
            .createBook({
              title: title.trim(),
              authors: authors
                .split(/,|&/)
                .map((s) => s.trim())
                .filter(Boolean),
              isbn: isbn.trim() || undefined,
              ownedPhysical: ownership === "physical",
              ownedDigital: ownership === "digital",
              fetchMetadata,
            })
            .then((book) => navigate(`/?q=${encodeURIComponent(book.title)}`))
            .catch((err: Error) => setError(err.message))
            .finally(() => setBusy(false));
        }}
      >
        <label className="text-xs uppercase tracking-wide text-muted">
          Title
          <input
            required={!isbn.trim()}
            className="mt-1 w-full rounded border border-border bg-surface px-3 py-2 text-sm text-text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
        </label>
        <label className="text-xs uppercase tracking-wide text-muted">
          Authors
          <input
            className="mt-1 w-full rounded border border-border bg-surface px-3 py-2 text-sm text-text"
            placeholder="Comma-separated"
            value={authors}
            onChange={(e) => setAuthors(e.target.value)}
          />
        </label>
        <label className="text-xs uppercase tracking-wide text-muted">
          ISBN
          <input
            className="mt-1 w-full rounded border border-border bg-surface px-3 py-2 text-sm text-text"
            value={isbn}
            onChange={(e) => setIsbn(e.target.value)}
          />
        </label>
        <label className="text-xs uppercase tracking-wide text-muted">
          Ownership
          <select
            className="mt-1 w-full rounded border border-border bg-surface px-3 py-2 text-sm text-text"
            value={ownership}
            onChange={(e) => setOwnership(e.target.value as typeof ownership)}
          >
            <option value="coveting">Coveting</option>
            <option value="physical">Physical</option>
            <option value="digital">Digital</option>
          </select>
        </label>
        <label className="flex items-center gap-2 text-sm text-text">
          <input
            type="checkbox"
            checked={fetchMetadata}
            onChange={(e) => setFetchMetadata(e.target.checked)}
          />
          Fetch cover and metadata from Open Library
        </label>
        {error && <p className="text-sm text-accent">{error}</p>}
        <button
          type="submit"
          disabled={busy}
          className="rounded bg-accent px-3 py-2 text-sm text-bg disabled:opacity-50"
        >
          {busy ? "Saving…" : "Add book"}
        </button>
      </form>
    </div>
  );
}
