import { useState } from "react";
import { useNavigate } from "react-router-dom";
import * as api from "../api";

export function ImportPage() {
  const [path, setPath] = useState("");
  const [preview, setPreview] = useState<number | null>(null);
  const [result, setResult] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  return (
    <div className="mx-auto max-w-xl px-6 py-10">
      <h1 className="font-serif text-3xl text-heading">Import from Calibre</h1>
      <p className="mt-3 text-sm leading-relaxed text-muted">
        Paste the path to your Calibre library folder — the one that contains{" "}
        <code className="text-heading">metadata.db</code>. Covers are copied locally. Re-running
        the import updates catalog metadata without overwriting want ratings, read status, or
        collections.
      </p>
      <label className="mt-6 block text-xs uppercase tracking-wide text-muted">
        Library path
      </label>
      <input
        className="mt-1 w-full rounded border border-border bg-surface px-3 py-2 text-sm"
        placeholder="/Users/you/Calibre Library"
        value={path}
        onChange={(e) => {
          setPath(e.target.value);
          setPreview(null);
          setResult(null);
        }}
      />
      {error && <p className="mt-3 text-sm text-accent">{error}</p>}
      {preview !== null && (
        <p className="mt-3 text-sm text-text">Calibre reports {preview} books.</p>
      )}
      {result && <p className="mt-3 text-sm text-text">{result}</p>}
      <div className="mt-5 flex gap-2">
        <button
          type="button"
          disabled={busy || !path.trim()}
          className="rounded border border-border px-3 py-2 text-sm disabled:opacity-50"
          onClick={() => {
            setBusy(true);
            setError(null);
            api
              .previewCalibre(path.trim())
              .then((r) => setPreview(r.books))
              .catch((e: Error) => setError(e.message))
              .finally(() => setBusy(false));
          }}
        >
          Preview
        </button>
        <button
          type="button"
          disabled={busy || !path.trim()}
          className="rounded bg-accent px-3 py-2 text-sm text-bg disabled:opacity-50"
          onClick={() => {
            setBusy(true);
            setError(null);
            api
              .importCalibre(path.trim())
              .then((r) => {
                setResult(
                  `Imported ${r.imported}, updated ${r.updated}, copied ${r.covers} covers.`,
                );
                setTimeout(() => navigate("/"), 800);
              })
              .catch((e: Error) => setError(e.message))
              .finally(() => setBusy(false));
          }}
        >
          {busy ? "Importing…" : "Import"}
        </button>
      </div>
    </div>
  );
}
