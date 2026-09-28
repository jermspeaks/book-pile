import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import * as api from "../api";
import type { Author } from "../types";

export function Authors() {
  const [authors, setAuthors] = useState<Author[]>([]);
  const [suggested, setSuggested] = useState<Author[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([api.listAuthors(), api.suggestAuthors(12)])
      .then(([all, s]) => {
        setAuthors(all);
        setSuggested(s);
      })
      .catch((e: Error) => setError(e.message));
  }, []);

  return (
    <div className="mx-auto max-w-3xl px-6 py-8">
      <h1 className="font-serif text-3xl text-heading">Authors</h1>
      {error && <p className="mt-3 text-sm text-accent">{error}</p>}
      {suggested.length > 0 && (
        <section className="mt-8">
          <h2 className="text-xs font-medium uppercase tracking-widest text-muted">
            Suggested authors
          </h2>
          <ul className="mt-3 grid gap-2 sm:grid-cols-2">
            {suggested.map((a) => (
              <AuthorCard key={a.id} author={a} />
            ))}
          </ul>
        </section>
      )}
      <section className="mt-10">
        <h2 className="text-xs font-medium uppercase tracking-widest text-muted">All authors</h2>
        <ul className="mt-3 divide-y divide-border">
          {authors.map((a) => (
            <li key={a.id}>
              <AuthorRow author={a} />
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}

function AuthorCard({ author }: { author: Author }) {
  return (
    <Link
      to={`/authors/${author.id}`}
      className="rounded-md border border-border bg-surface p-3 no-underline hover:bg-surface-hover"
    >
      <div className="text-heading">{author.name}</div>
      <div className="mt-1 text-xs text-muted">
        {author.unreadOwned} unread owned · {author.unreadCoveting} coveting
      </div>
    </Link>
  );
}

function AuthorRow({ author }: { author: Author }) {
  return (
    <Link
      to={`/authors/${author.id}`}
      className="flex items-baseline justify-between gap-3 py-2 no-underline"
    >
      <span className="text-heading">{author.name}</span>
      <span className="text-xs text-muted">
        {author.bookCount} books · {author.readCount} read
      </span>
    </Link>
  );
}
