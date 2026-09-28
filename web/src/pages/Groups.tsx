import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import * as api from "../api";
import { Library } from "./Library";
import type { Named } from "../types";

type Kind = "piles" | "collections" | "genres";

export function GroupList({ kind }: { kind: Kind }) {
  const [items, setItems] = useState<Named[]>([]);
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const title = kind[0].toUpperCase() + kind.slice(1);

  useEffect(() => {
    const load =
      kind === "piles" ? api.listPiles : kind === "collections" ? api.listCollections : api.listGenres;
    load()
      .then(setItems)
      .catch((e: Error) => setError(e.message));
  }, [kind]);

  const canCreate = kind !== "genres";

  return (
    <div className="mx-auto max-w-xl px-6 py-8">
      <h1 className="font-serif text-3xl text-heading">{title}</h1>
      <p className="mt-2 text-sm text-muted">
        {kind === "piles" && "Named locations and sources — bedside, Kindle, Calibre."}
        {kind === "collections" && "Groups you make yourself."}
        {kind === "genres" && "Genres from Calibre tags and anything you assign later."}
      </p>
      {error && <p className="mt-3 text-sm text-accent">{error}</p>}
      {canCreate && (
        <form
          className="mt-4 flex gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            const n = name.trim();
            if (!n) return;
            const create = kind === "piles" ? api.createPile : api.createCollection;
            void create(n).then((item) => {
              setItems((cur) => [...cur, item].sort((a, b) => a.name.localeCompare(b.name)));
              setName("");
            });
          }}
        >
          <input
            className="min-w-0 flex-1 rounded border border-border bg-surface px-3 py-2 text-sm"
            placeholder={`New ${kind.slice(0, -1)}`}
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <button type="submit" className="rounded bg-accent px-3 py-2 text-sm text-bg">
            Create
          </button>
        </form>
      )}
      <ul className="mt-6 divide-y divide-border">
        {items.map((item) => (
          <li key={item.id}>
            <Link to={`/${kind}/${item.id}`} className="block py-3 text-heading no-underline">
              {item.name}
            </Link>
          </li>
        ))}
        {items.length === 0 && <li className="py-6 text-sm text-muted">Nothing here yet.</li>}
      </ul>
    </div>
  );
}

export function GroupBrowse({ kind }: { kind: Kind }) {
  const { id } = useParams();
  const [named, setNamed] = useState<Named | null>(null);

  useEffect(() => {
    if (!id) return;
    const n = Number(id);
    const load =
      kind === "piles" ? api.getPile : kind === "collections" ? api.getCollection : api.getGenre;
    load(n).then(setNamed).catch(() => setNamed({ id: n, name: titleCase(kind) }));
  }, [kind, id]);

  const locked = useMemo(() => {
    const filterKey = kind === "piles" ? "pileId" : kind === "collections" ? "collectionId" : "genreId";
    return { [filterKey]: id };
  }, [kind, id]);

  if (!id) return null;
  return (
    <Library
      heading={named?.name}
      showNext={false}
      locked={locked}
    />
  );
}

function titleCase(s: string) {
  return s[0].toUpperCase() + s.slice(1);
}
